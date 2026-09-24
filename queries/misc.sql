-- name: get-dashboard-charts
SELECT data FROM mat_dashboard_charts;

-- name: get-dashboard-counts
SELECT data FROM mat_dashboard_counts;

-- name: get-dashboard-events
-- Unified recent events feed for the dashboard card: campaign sends,
-- opens, link clicks and opt-ins to public lists, newest first.
SELECT * FROM (
    SELECT 'campaign_sent' AS type, COALESCE(started_at, updated_at) AS created_at,
        id AS campaign_id, name AS campaign_name,
        NULL::INTEGER AS subscriber_id, NULL::TEXT AS email, NULL::TEXT AS subscriber_name,
        NULL::TEXT AS list_name, NULL::TEXT AS url, NULL::TEXT AS track,
        NULL::INTEGER AS open_count
    FROM campaigns WHERE started_at IS NOT NULL
    UNION ALL
    SELECT 'open', v.created_at, v.campaign_id, c.name,
        v.subscriber_id, s.email, s.name, NULL, NULL, NULL,
        CASE WHEN v.subscriber_id IS NULL THEN NULL
            ELSE ROW_NUMBER() OVER (
                PARTITION BY v.campaign_id, v.subscriber_id
                ORDER BY v.created_at, v.id)
        END::INTEGER
    FROM campaign_views v
    JOIN campaigns c ON c.id = v.campaign_id
    LEFT JOIN subscribers s ON s.id = v.subscriber_id
    -- Scanner filter the earliest view of a (campaign, subscriber) pair within 4
    -- minutes of the send is mail-gateway noise, not an open. Copies the canonical
    -- fragment documented in queries/campaigns.sql (get-campaign-stats).
    WHERE NOT (v.subscriber_id IS NOT NULL
        AND c.started_at IS NOT NULL
        AND v.created_at <= c.started_at + INTERVAL '4 minutes'
        AND NOT EXISTS (
            SELECT 1 FROM campaign_views earlier
            WHERE earlier.campaign_id = v.campaign_id
            AND earlier.subscriber_id = v.subscriber_id
            AND (earlier.created_at, earlier.id) < (v.created_at, v.id)
        ))
    -- Burst filter, rapid duplicate opens within 5 seconds of each other
    -- collapse to the latest view of the burst.
    AND NOT EXISTS (
        SELECT 1 FROM campaign_views later
        WHERE later.campaign_id = v.campaign_id
        AND later.subscriber_id = v.subscriber_id
        AND (later.created_at, later.id) > (v.created_at, v.id)
        AND later.created_at <= v.created_at + INTERVAL '5 seconds'
    )
    UNION ALL
    SELECT 'click', lc.created_at, lc.campaign_id, c.name,
        lc.subscriber_id, s.email, s.name, NULL, l.url, NULL, NULL
    FROM link_clicks lc
    JOIN links l ON l.id = lc.link_id
    LEFT JOIN campaigns c ON c.id = lc.campaign_id
    LEFT JOIN subscribers s ON s.id = lc.subscriber_id
    UNION ALL
    SELECT 'optin', sl.created_at, NULL, NULL,
        sl.subscriber_id, s.email, s.name, l.name, NULL, NULL, NULL
    FROM subscriber_lists sl
    JOIN lists l ON l.id = sl.list_id
    JOIN subscribers s ON s.id = sl.subscriber_id
    WHERE l.type = 'public' AND sl.status != 'unsubscribed'
    UNION ALL
    SELECT 'unsubscribe', sl.updated_at, NULL, NULL,
        sl.subscriber_id, s.email, s.name, l.name, NULL, NULL, NULL
    FROM subscriber_lists sl
    JOIN lists l ON l.id = sl.list_id
    JOIN subscribers s ON s.id = sl.subscriber_id
    WHERE sl.status = 'unsubscribed'
) e ORDER BY created_at DESC LIMIT $1;

-- name: get-settings
SELECT JSON_OBJECT_AGG(key, value) AS settings FROM (SELECT * FROM settings ORDER BY key) t;

-- name: update-settings
UPDATE settings AS s SET value = c.value
    -- For each key in the incoming JSON map, update the row with the key and its value.
    FROM(SELECT * FROM JSONB_EACH($1)) AS c(key, value) WHERE s.key = c.key;

-- name: update-settings-by-key
UPDATE settings SET value = $2, updated_at = NOW() WHERE key = $1;

-- name: get-db-info
SELECT JSON_BUILD_OBJECT('version', (SELECT VERSION()),
                        'size_mb', (SELECT ROUND(pg_database_size((SELECT CURRENT_DATABASE()))/(1024^2)))) AS info;
