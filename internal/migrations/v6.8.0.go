package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"

	"github.com/knadh/koanf/v2"

	"github.com/knadh/stuffbin"
)

// V6_8_0 reworks campaign colours from per palette entry picked off the row id
// into one colour per THEME, and rewrites every existing row to match.
//
// v6.7.0 backfilled a colour per campaign type and gave every other row the
// palette entry at (id % 8). That meant two sends from the same script landed
// on different colours, so the chip could not be read as a category. This
// migration recomputes the colour of EVERY campaign from its name, discarding
// the old id based colour, because that colour carried no meaning:
//
//   onboard    '#14b8a6' every onboarding send (tt onboard, ig onboard)
//   one-off    '#6366f1' 1:1 sends (tt one-off)
//   loopkit    '#f59e0b' loopkit drops (tt loopkit drop, bs loopkit drop)
//   email list '#3e6fbf' email list and weekly loops sends
//   no match   ''        neutral grey in the UI, never a random palette colour
//
// The match mirrors cmd/campaign_theme.go and CAMPAIGN_THEMES in
// frontend/src/utils.js, so a draft and its sent copy, and a campaign created
// before and after the upgrade, all resolve to the same colour. Colours picked
// by hand in the UI are overwritten here and stay untouched afterwards, as the
// migration runs once.
func V6_8_0(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	if _, err := db.Exec(`
		UPDATE campaigns SET color = CASE
			WHEN src.n LIKE 'onboard%' OR src.n LIKE 'tt onboard%' OR src.n LIKE 'ig onboard%' THEN '#14b8a6'
			WHEN src.n LIKE 'one-off%' OR src.n LIKE 'one off%' OR src.n LIKE 'tt one-off%' THEN '#6366f1'
			WHEN src.n LIKE 'loopkit%' OR src.n LIKE 'tt loopkit%' OR src.n LIKE 'bs loopkit%' THEN '#f59e0b'
			WHEN src.n LIKE 'email list%' OR src.n LIKE 'slayr email list%' OR src.n LIKE '8digit weekly loops%' OR src.n LIKE 'weekly loops%' THEN '#3e6fbf'
			ELSE ''
		END
		FROM (
			SELECT id,
				btrim(regexp_replace(
					regexp_replace(lower(name), '\s+', ' ', 'g'),
					'^\s*copy\s+of\s+', ''
				)) AS n
			FROM campaigns
		) src
		WHERE campaigns.id = src.id;
	`); err != nil {
		return err
	}

	return nil
}
