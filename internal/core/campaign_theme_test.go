package core

import "testing"

// The names mirror the live campaigns on the Pi plus the shapes the scripts
// create, so a drift between the Go derivation, the SQL backfill and the UI
// fallback shows up as a failing colour.
func TestCampaignThemeColor(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"tt onboard: simone.picazzi@gmail.com", "#14b8a6"},
		{"tt onboard: lilrando933@gmail.com", "#14b8a6"},
		{"tt onboard", "#14b8a6"},
		{"ig onboard: prod.blaxor@gmail.com", "#14b8a6"},
		{"ig onboard", "#14b8a6"},
		{"Copy of tt onboard: someone@example.com", "#14b8a6"},
		{"  IG Onboard: someone@example.com  ", "#14b8a6"},
		{"onboarding blast 10/07", "#14b8a6"},
		{"tt one-off", "#6366f1"},
		{"tt one-off: cozyrated@gmail.com", "#6366f1"},
		{"copy of tt one-off: cozyrated@gmail.com", "#6366f1"},
		{"one-off: cozyrated@gmail.com", "#6366f1"},
		{"tt loopkit drop", "#f59e0b"},
		{"tt loopkit drop 10/07", "#f59e0b"},
		{"tt loopkit drop \"MOONDROP\"", "#f59e0b"},
		{"tt loopkit template", "#f59e0b"},
		{"bs loopkit drop", "#f59e0b"},
		{"bs loopkit drop \"MOONDROP\"", "#f59e0b"},
		{"slayr email list", "#3e6fbf"},
		{"slayr email list - sep week 2", "#3e6fbf"},
		{"slayr email list - oct week 1", "#3e6fbf"},
		{"8digit weekly loops email list", "#3e6fbf"},
		{"8digit weekly loops - oct week 1", "#3e6fbf"},
		{"8digit weekly loops email list 10/07", "#3e6fbf"},

		// No theme: stays empty so the UI renders neutral grey.
		{"simonlmao@yahoo.com", ""},
		{"", ""},
		{"newsletter", ""},
		{"Copy of newsletter", ""},
	}

	for _, c := range cases {
		if got := CampaignThemeColor(c.name); got != c.want {
			t.Errorf("CampaignThemeColor(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

// Every theme must be reachable from the color map, so a typo in a key cannot
// leave a whole category grey.
func TestCampaignThemeKeysCovered(t *testing.T) {
	for _, t2 := range campaignThemePrefixes {
		if _, ok := campaignThemeColors[t2.key]; !ok {
			t.Errorf("theme %q has prefixes but no color", t2.key)
		}
	}
}
