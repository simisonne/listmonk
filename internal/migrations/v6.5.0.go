package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/knadh/koanf/v2"

	"github.com/knadh/stuffbin"
)

func V6_5_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	// Rebuild the dashboard charts materialized view so its per-day click
	// counts also collapse rapid duplicate clicks. A click is a burst
	// duplicate when the same (campaign_id, subscriber_id, link_id) triple
	// recorded a later click within 5 seconds of it, and only the latest
	// click of a burst counts. The views CTE is carried over unchanged from
	// V6_4_0 (scanner plus burst filters). Kept in sync with schema.sql
	// (mat_dashboard_charts clicks CTE). No rows are touched, only the
	// derived view is redefined, so this is safe to re-run and exports
	// stay raw.
	if _, err := db.Exec(`
		DROP MATERIALIZED VIEW IF EXISTS mat_dashboard_charts;
		CREATE MATERIALIZED VIEW mat_dashboard_charts AS
		    WITH clicks AS (
		        SELECT JSON_AGG(ROW_TO_JSON(row))
		        FROM (
		            WITH viewDates AS (
		              SELECT created_at::DATE AS to_date,
		                     created_at::DATE - INTERVAL '30 DAY' AS from_date
		                     FROM link_clicks ORDER BY id DESC LIMIT 1
		            )
		            SELECT COUNT(*) AS count, lc.created_at::DATE as date FROM link_clicks lc
		              WHERE lc.created_at >= (SELECT from_date FROM viewDates)
		                AND lc.created_at < (SELECT to_date FROM viewDates) + INTERVAL '1 day'
		                AND NOT EXISTS (
		                    SELECT 1 FROM link_clicks later
		                    WHERE later.campaign_id = lc.campaign_id
		                    AND later.subscriber_id = lc.subscriber_id
		                    AND later.link_id = lc.link_id
		                    AND (later.created_at, later.id) > (lc.created_at, lc.id)
		                    AND later.created_at <= lc.created_at + INTERVAL '5 seconds'
		                )
		              GROUP by date ORDER BY date
		        ) row
		    ),
		    views AS (
		        SELECT JSON_AGG(ROW_TO_JSON(row))
		        FROM (
		            WITH viewDates AS (
		              SELECT created_at::DATE AS to_date,
		                     created_at::DATE - INTERVAL '30 DAY' AS from_date
		                     FROM campaign_views ORDER BY id DESC LIMIT 1
		            )
		            SELECT COUNT(*) AS count, v.created_at::DATE as date FROM campaign_views v
		              JOIN campaigns c ON (c.id = v.campaign_id)
		              WHERE v.created_at >= (SELECT from_date FROM viewDates)
		                AND v.created_at < (SELECT to_date FROM viewDates) + INTERVAL '1 day'
		                AND NOT (v.subscriber_id IS NOT NULL
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
		              GROUP by date ORDER BY date
		        ) row
		    )
		    SELECT NOW() AS updated_at, JSON_BUILD_OBJECT('link_clicks', COALESCE((SELECT * FROM clicks), '[]'),
		                                  'campaign_views', COALESCE((SELECT * FROM views), '[]')
		                                ) AS data;
		DROP INDEX IF EXISTS mat_dashboard_charts_idx;
		CREATE UNIQUE INDEX mat_dashboard_charts_idx ON mat_dashboard_charts (updated_at);
	`); err != nil {
		return err
	}

	return nil
}
