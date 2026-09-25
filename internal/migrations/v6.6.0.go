package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/knadh/koanf/v2"

	"github.com/knadh/stuffbin"
)

func V6_6_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	// Rebuild the dashboard counts materialized view so the messages card
	// also reports total opens and clicks next to messages sent. Opens reuse
	// the canonical read filters (4 minute mail gateway scanner window plus
	// 5 second burst collapse) and clicks reuse the 5 second burst collapse,
	// in sync with schema.sql (mat_dashboard_counts) and the analytics and
	// charts counts. No rows are touched, only the derived view is
	// redefined, so this is safe to rerun and exports stay raw.
	if _, err := db.Exec(`
		DROP MATERIALIZED VIEW IF EXISTS mat_dashboard_counts;
		CREATE MATERIALIZED VIEW mat_dashboard_counts AS
		    WITH subs AS (
		        SELECT COUNT(*) AS num, status FROM subscribers GROUP BY status
		    )
		    SELECT NOW() AS updated_at,
		        JSON_BUILD_OBJECT(
		            'subscribers', JSON_BUILD_OBJECT(
		                'total', (SELECT SUM(num) FROM subs),
		                'blocklisted', (SELECT num FROM subs WHERE status='blocklisted'),
		                'orphans', (
		                    SELECT COUNT(id) FROM subscribers
		                    LEFT JOIN subscriber_lists ON (subscribers.id = subscriber_lists.subscriber_id)
		                    WHERE subscriber_lists.subscriber_id IS NULL
		                )
		            ),
		            'lists', JSON_BUILD_OBJECT(
		                'total', (SELECT COUNT(*) FROM lists),
		                'private', (SELECT COUNT(*) FROM lists WHERE type='private'),
		                'public', (SELECT COUNT(*) FROM lists WHERE type='public'),
		                'optin_single', (SELECT COUNT(*) FROM lists WHERE optin='single'),
		                'optin_double', (SELECT COUNT(*) FROM lists WHERE optin='double')
		            ),
		            'campaigns', JSON_BUILD_OBJECT(
		                'total', (SELECT COUNT(*) FROM campaigns),
		                'by_status', (
		                    SELECT JSON_OBJECT_AGG (status, num) FROM
		                    (SELECT status, COUNT(*) AS num FROM campaigns GROUP BY status) r
		                )
		            ),
		            'messages', JSON_BUILD_OBJECT(
		                'total', (SELECT SUM(sent) AS messages FROM campaigns),
		                'opens', (
		                    SELECT COUNT(*) FROM campaign_views v
		                    JOIN campaigns c ON (c.id = v.campaign_id)
		                    WHERE NOT (v.subscriber_id IS NOT NULL
		                        AND c.started_at IS NOT NULL
		                        AND v.created_at <= c.started_at + INTERVAL '4 minutes'
		                        AND NOT EXISTS (
		                            SELECT 1 FROM campaign_views earlier
		                            WHERE earlier.campaign_id = v.campaign_id
		                            AND earlier.subscriber_id = v.subscriber_id
		                            AND (earlier.created_at, earlier.id) < (v.created_at, v.id)
		                        ))
		                    AND NOT EXISTS (
		                        SELECT 1 FROM campaign_views later
		                        WHERE later.campaign_id = v.campaign_id
		                        AND later.subscriber_id = v.subscriber_id
		                        AND (later.created_at, later.id) > (v.created_at, v.id)
		                        AND later.created_at <= v.created_at + INTERVAL '5 seconds'
		                    )
		                ),
		                'clicks', (
		                    SELECT COUNT(*) FROM link_clicks lc
		                    WHERE NOT EXISTS (
		                        SELECT 1 FROM link_clicks later
		                        WHERE later.campaign_id = lc.campaign_id
		                        AND later.subscriber_id = lc.subscriber_id
		                        AND later.link_id = lc.link_id
		                        AND (later.created_at, later.id) > (lc.created_at, lc.id)
		                        AND later.created_at <= lc.created_at + INTERVAL '5 seconds'
		                    )
		                )
		            )
		        ) AS data;
		DROP INDEX IF EXISTS mat_dashboard_stats_idx;
		CREATE UNIQUE INDEX mat_dashboard_stats_idx ON mat_dashboard_counts (updated_at);
	`); err != nil {
		return err
	}

	return nil
}
