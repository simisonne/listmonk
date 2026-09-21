package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

func V6_3_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	// One-time cleanup of automated scanner opens that were recorded before the
	// write path started rejecting them (see `register-campaign-view` in
	// queries/campaigns.sql). Mail gateways (Safe Links, Proofpoint, Barracuda and
	// friends) prefetch the tracking pixel seconds after a send starts, which lands
	// one bogus "open" per subscriber in the first minutes of a campaign.
	//
	// A row is scanner noise when it is the EARLIEST view of its
	// (campaign_id, subscriber_id) pair and it was recorded within 4 minutes of
	// campaigns.started_at. Only that single earliest row per pair is deleted:
	// rows with a NULL subscriber_id and any later view inside the same window are
	// kept, because a second view is a real open.
	//
	// This DELETE is not idempotent on its own: once the earliest rows are gone, the
	// next view of a pair inside the window becomes "the earliest" and a second run
	// would delete it too. Migration versions are recorded in settings, so this runs
	// once per database.
	if _, err := db.Exec(`
		DELETE FROM campaign_views
		WHERE id IN (
			SELECT v.id
			FROM campaign_views v
			JOIN campaigns c ON (c.id = v.campaign_id)
			WHERE v.subscriber_id IS NOT NULL
			AND c.started_at IS NOT NULL
			AND v.created_at <= c.started_at + INTERVAL '4 minutes'
			AND NOT EXISTS (
				SELECT 1 FROM campaign_views earlier
				WHERE earlier.campaign_id = v.campaign_id
				AND earlier.subscriber_id = v.subscriber_id
				AND (earlier.created_at, earlier.id) < (v.created_at, v.id)
			)
		);
	`); err != nil {
		return err
	}

	return nil
}
