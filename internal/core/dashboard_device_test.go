package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadMelodiesEventsDeviceBrowser(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	lines := `[2026-09-23 10:00:00] user=guest action=melodies_page device=iPhone browser=Safari ip=1.2.3.4
[2026-09-23 10:01:00] user=guest action=song_play song=foo.mp3 device=Pixel browser=Chrome ip=1.2.3.4
[2026-09-23 10:02:00] user=guest action=song_download song=bar.mp3 browser=Firefox ip=1.2.3.4
[2026-09-23 10:03:00] user=guest action=melodies_page ip=1.2.3.4
[2026-09-23 10:04:00] user=guest action=song_play song=my cool song.mp3 device=iPad browser=Chrome ip=1.2.3.4
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)
	if len(evs) != 5 {
		t.Fatalf("got %d events, want 5", len(evs))
	}

	// New style line: both fields set.
	if evs[0].Type != "site_visit" || evs[0].Device == nil || *evs[0].Device != "iPhone" ||
		evs[0].Browser == nil || *evs[0].Browser != "Safari" {
		t.Errorf("event 0 = %+v, want site_visit iPhone/Safari", evs[0])
	}

	// Track still parsed alongside device.
	if evs[1].Type != "track_played" || evs[1].Track == nil || *evs[1].Track != "foo.mp3" ||
		evs[1].Device == nil || *evs[1].Device != "Pixel" || evs[1].Browser == nil || *evs[1].Browser != "Chrome" {
		t.Errorf("event 1 = %+v, want track_played foo.mp3 Pixel/Chrome", evs[1])
	}

	// Only browser present, device must stay nil.
	if evs[2].Type != "track_downloaded" || evs[2].Device != nil ||
		evs[2].Browser == nil || *evs[2].Browser != "Firefox" {
		t.Errorf("event 2 = %+v, want track_downloaded nil/Firefox", evs[2])
	}

	// Old style line: both nil.
	if evs[3].Type != "site_visit" || evs[3].Device != nil || evs[3].Browser != nil {
		t.Errorf("event 3 = %+v, want site_visit with nil device and browser", evs[3])
	}

	// Song name with spaces stops at device=.
	if evs[4].Type != "track_played" || evs[4].Track == nil || *evs[4].Track != "my cool song.mp3" ||
		evs[4].Device == nil || *evs[4].Device != "iPad" || evs[4].Browser == nil || *evs[4].Browser != "Chrome" {
		t.Errorf("event 4 = %+v, want track_played my cool song.mp3 iPad/Chrome", evs[4])
	}
}

func TestReadMelodiesEventsPortfolio(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	lines := `[2026-10-02 11:01:00] user=guest action=portfolio_referral ref=@tiktok device=Windows_PC browser=Edge ip=203.0.113.9 loc=Berlin_DE
[2026-10-02 11:00:00] user=guest action=portfolio_visit path=/hire-me/education.html device=iPhone browser=Safari ip=203.0.113.9 loc=Berlin_DE
[2026-10-02 10:59:00] user=guest action=portfolio_visit path=/hire-me/ loc=local ip=203.0.113.9
[2026-10-02 10:58:00] user=guest action=portfolio_referral ref=google loc=DE ip=203.0.113.9
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)
	if len(evs) != 4 {
		t.Fatalf("got %d events, want 4", len(evs))
	}

	// Referral: ref and loc parsed alongside device.
	if evs[0].Type != "portfolio_referral" || evs[0].Ref == nil || *evs[0].Ref != "@tiktok" ||
		evs[0].Location == nil || *evs[0].Location != "Berlin_DE" ||
		evs[0].Device == nil || *evs[0].Device != "Windows_PC" || evs[0].Browser == nil || *evs[0].Browser != "Edge" {
		t.Errorf("event 0 = %+v, want portfolio_referral @tiktok Berlin_DE Windows_PC/Edge", evs[0])
	}

	// Visit: path and loc parsed alongside device and browser.
	if evs[1].Type != "portfolio_visit" || evs[1].Path == nil || *evs[1].Path != "/hire-me/education.html" ||
		evs[1].Location == nil || *evs[1].Location != "Berlin_DE" ||
		evs[1].Device == nil || *evs[1].Device != "iPhone" || evs[1].Browser == nil || *evs[1].Browser != "Safari" {
		t.Errorf("event 1 = %+v, want portfolio_visit /hire-me/education.html Berlin_DE iPhone/Safari", evs[1])
	}

	// No device or browser, ref must stay nil.
	if evs[2].Type != "portfolio_visit" || evs[2].Path == nil || *evs[2].Path != "/hire-me/" ||
		evs[2].Location == nil || *evs[2].Location != "local" ||
		evs[2].Device != nil || evs[2].Browser != nil || evs[2].Ref != nil {
		t.Errorf("event 2 = %+v, want portfolio_visit /hire-me/ loc=local and nil device, browser, ref", evs[2])
	}

	// Country only location, path must stay nil.
	if evs[3].Type != "portfolio_referral" || evs[3].Ref == nil || *evs[3].Ref != "google" ||
		evs[3].Location == nil || *evs[3].Location != "DE" ||
		evs[3].Path != nil || evs[3].Device != nil || evs[3].Browser != nil {
		t.Errorf("event 3 = %+v, want portfolio_referral google loc=DE with nil path, device, browser", evs[3])
	}
}

func TestReadMelodiesEventsIPAndHiddenLocations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	lines := `[2026-10-02 12:00:00] user=guest action=portfolio_visit path=/hire-me/education.html device=Mac browser=Chrome ip=203.0.113.9 loc=Berlin_DE
[2026-10-02 11:59:00] user=guest action=portfolio_visit path=/hire-me/ device=Mac browser=Chrome ip=203.0.113.9 loc=Regensburg_DE
[2026-10-02 11:58:00] user=guest action=melodies_page device=iPhone browser=Safari ip=203.0.113.9 loc=Regensburg_DE
[2026-10-02 11:57:00] user=guest action=melodies_page device=iPhone browser=Safari ip=203.0.113.9 loc=regensburg_de
[2026-10-02 11:56:00] user=guest action=melodies_page device=iPhone browser=Safari ip=203.0.113.9
[2026-10-02 11:55:00] user=guest action=song_play song=foo.mp3 device=Pixel browser=Chrome ip=198.51.100.7 loc=Munich_DE
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)

	// The three Regensburg rows (portfolio, melodies, lower case) are dropped,
	// the other three survive.
	if len(evs) != 3 {
		t.Fatalf("got %d events, want 3 (Regensburg rows dropped): %+v", len(evs), evs)
	}
	for i, ev := range evs {
		if ev.Location != nil && isHiddenLocation(*ev.Location) {
			t.Errorf("event %d = %+v, hidden location survived the filter", i, ev)
		}
	}

	// Kept portfolio row: ip parsed next to loc.
	if evs[0].Type != "portfolio_visit" || evs[0].Location == nil || *evs[0].Location != "Berlin_DE" ||
		evs[0].IP == nil || *evs[0].IP != "203.0.113.9" {
		t.Errorf("event 0 = %+v, want portfolio_visit Berlin_DE with ip 203.0.113.9", evs[0])
	}

	// Kept melodies row with no loc at all: nil location, ip still parsed.
	if evs[1].Type != "site_visit" || evs[1].Location != nil ||
		evs[1].IP == nil || *evs[1].IP != "203.0.113.9" {
		t.Errorf("event 1 = %+v, want site_visit with nil location and ip 203.0.113.9", evs[1])
	}

	// Kept melodies row from another city, track parsed alongside ip.
	if evs[2].Type != "track_played" || evs[2].Location == nil || *evs[2].Location != "Munich_DE" ||
		evs[2].IP == nil || *evs[2].IP != "198.51.100.7" {
		t.Errorf("event 2 = %+v, want track_played Munich_DE with ip 198.51.100.7", evs[2])
	}
}
