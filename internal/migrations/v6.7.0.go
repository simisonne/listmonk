package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/knadh/koanf/v2"

	"github.com/knadh/stuffbin"
)

// V6_7_0 adds campaigns.color, the display colour shown as a chip next to
// every campaign name, and backfills the rows that already exist.
//
// Backfill rules, mirrored by the frontend fallback in frontend/src/utils.js
// (campaignPalette / campaignChipStyle), so both sides agree for the same row:
//
//   - A campaign whose name, lowercased and stripped of a leading "copy of ",
//     starts with one of the known prefixes gets that group's base colour.
//     The email list prefix covers both "email list" and the "slayr email
//     list" names the list automation creates.
//   - Every other row gets the palette entry at (id % 8), listed in the same
//     order as CAMPAIGN_PALETTE in frontend/src/utils.js.
//
// Only rows still holding the empty column default are touched, so rerunning
// the migration never overwrites a colour that was set later.
func V6_7_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	if _, err := db.Exec(`ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS color TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}

	if _, err := db.Exec(`
		UPDATE campaigns SET color = CASE
			WHEN src.name LIKE 'tt onboard%' THEN '#14B8A6'
			WHEN src.name LIKE 'tt one-off%' THEN '#6366F1'
			WHEN src.name LIKE 'ig onboard%' THEN '#EC4899'
			WHEN src.name LIKE 'email list%' THEN '#3E6FBF'
			WHEN src.name LIKE 'slayr email list%' THEN '#3E6FBF'
			WHEN src.name LIKE 'bs loopkit%' THEN '#F59E0B'
			WHEN campaigns.id % 8 = 0 THEN '#14B8A6'
			WHEN campaigns.id % 8 = 1 THEN '#6366F1'
			WHEN campaigns.id % 8 = 2 THEN '#EC4899'
			WHEN campaigns.id % 8 = 3 THEN '#3E6FBF'
			WHEN campaigns.id % 8 = 4 THEN '#F59E0B'
			WHEN campaigns.id % 8 = 5 THEN '#10B981'
			WHEN campaigns.id % 8 = 6 THEN '#8B5CF6'
			WHEN campaigns.id % 8 = 7 THEN '#0891B2'
		END
		FROM (
			SELECT id,
				regexp_replace(lower(name), '^[[:space:]]*copy[[:space:]]+of[[:space:]]+', '') AS name
			FROM campaigns
			WHERE color = ''
		) src
		WHERE campaigns.id = src.id;
	`); err != nil {
		return err
	}

	return nil
}
