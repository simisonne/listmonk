package core

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadSiteEventsKeepsHiddenVisits covers what the Website page adds to the
// log reader: Simon's own egress (Regensburg, and the private network token) is
// kept and flagged Hidden when the caller asks for it, instead of dropped the
// way the dashboard feed drops it. His own logins stay out either way.
func TestReadSiteEventsKeepsHiddenVisits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	lines := `[2026-10-20 10:00:00] user=guest action=melodies_page loc=Regensburg_DE ip=203.0.113.1
[2026-10-20 09:59:00] user=guest action=beats_page loc=Berlin_DE ip=203.0.113.2
[2026-10-20 09:58:00] user=guest action=song_play song=foo.mp3 site=loops loc=Paris_FR ip=203.0.113.3
[2026-10-20 09:57:00] user=guest action=portfolio_visit path=/hire-me/ loc=London_GB ip=203.0.113.5
[2026-10-20 09:55:00] user=guest action=weekly_loops_page loc=local ip=10.0.0.9
[2026-10-20 09:54:00] user=simon action=melodies_page loc=Berlin_DE ip=203.0.113.6
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	// Dashboard path: his own network's visits and his login are both gone.
	plain := readMelodiesEvents(50)
	if len(plain) != 3 {
		t.Fatalf("plain reader got %d events, want 3: %+v", len(plain), plain)
	}
	for _, e := range plain {
		if e.Hidden {
			t.Errorf("nothing from the dashboard feed may be flagged Hidden: %+v", e)
		}
	}

	// Website page: his own network's visits come back, flagged.
	all := readSiteEvents(50, siteEventOptions{includeHiddenLocations: true})
	if len(all) != 5 {
		t.Fatalf("inclusive reader got %d events, want 5: %+v", len(all), all)
	}

	byIP := map[string]bool{}
	for _, e := range all {
		if e.IP == nil {
			t.Fatalf("event with no ip: %+v", e)
		}
		byIP[*e.IP] = e.Hidden
	}
	if hidden, ok := byIP["203.0.113.1"]; !ok || !hidden {
		t.Error("the Regensburg visit must come back flagged Hidden")
	}
	if hidden, ok := byIP["10.0.0.9"]; !ok || !hidden {
		t.Error("the private network visit must come back flagged Hidden")
	}
	if hidden := byIP["203.0.113.2"]; hidden {
		t.Error("a real visitor location must not be flagged Hidden")
	}

	// Simon's own login never comes back, even from the inclusive reader.
	if _, ok := byIP["203.0.113.6"]; ok {
		t.Fatal("user=simon line leaked into the site events")
	}
}
