package core

import (
	"regexp"
	"strings"
)

// Campaign theme colors. A campaign's chip color is decided by its theme, not
// by its row id, so every send from the same script and every send in the same
// category (all onboards, all one offs, all loopkit drops, all email list
// sends) shows one color. Campaigns whose name matches no theme keep an empty
// color and render neutral grey in the UI.
//
// This table is mirrored by CAMPAIGN_PALETTE and CAMPAIGN_THEMES in
// frontend/src/utils.js (the UI fallback for rows with no stored color) and by
// the backfill in internal/migrations/v6.8.0.go (existing rows). Edit all three
// together.
var campaignThemeColors = map[string]string{
	"onboard":    "#14b8a6",
	"one-off":    "#6366f1",
	"loopkit":    "#f59e0b",
	"email-list": "#3e6fbf",
}

// Theme prefixes, most specific first, matched against the lowercased name
// after a leading "copy of " is stripped. The bare theme words catch scripts
// and hand made campaigns named without a brand prefix.
var campaignThemePrefixes = []struct {
	key      string
	prefixes []string
}{
	{"onboard", []string{"tt onboard", "ig onboard", "onboard"}},
	{"one-off", []string{"tt one-off", "one-off", "one off"}},
	{"loopkit", []string{"tt loopkit", "bs loopkit", "loopkit"}},
	{"email-list", []string{"slayr email list", "email list", "8digit weekly loops", "weekly loops"}},
}

var reCopyOf = regexp.MustCompile(`^\s*copy\s+of\s+`)

// CampaignTheme returns the theme key for a campaign name, "" when it matches
// no known theme.
func CampaignTheme(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	// Collapse runs of whitespace so "tt  onboard" matches like "tt onboard".
	n = strings.Join(strings.Fields(n), " ")
	n = reCopyOf.ReplaceAllString(n, "")

	for _, t := range campaignThemePrefixes {
		for _, p := range t.prefixes {
			if strings.HasPrefix(n, p) {
				return t.key
			}
		}
	}

	return ""
}

// CampaignThemeColor returns the color for a campaign name, "" when the name
// matches no theme.
func CampaignThemeColor(name string) string {
	return campaignThemeColors[CampaignTheme(name)]
}
