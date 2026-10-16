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
	// The loc=local line is a private address, which means the Pi's own network,
	// so it is dropped at read time: three of the four lines survive.
	if len(evs) != 3 {
		t.Fatalf("got %d events, want 3", len(evs))
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

	// The private network row from the fixture must not appear at all.
	for _, ev := range evs {
		if ev.Location != nil && *ev.Location == "local" {
			t.Errorf("a loc=local row reached the feed: %+v", ev)
		}
	}

	// Country only location, path must stay nil.
	if evs[2].Type != "portfolio_referral" || evs[2].Ref == nil || *evs[2].Ref != "google" ||
		evs[2].Location == nil || *evs[2].Location != "DE" ||
		evs[2].Path != nil || evs[2].Device != nil || evs[2].Browser != nil {
		t.Errorf("event 2 = %+v, want portfolio_referral google loc=DE with nil path, device, browser", evs[2])
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

func TestReadMelodiesEventsWeeklyLoops(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	// The opt in line is not a feed type any more: listmonk writes its own
	// optin row for a signup. Regensburg is Simon's own egress and "local" is
	// his own network, both dropped at read time. The sparse line, the plain
	// city and Hamburg all survive.
	lines := `[2026-10-05 09:10:00] user=guest action=email_list_subscribe result=ok device=Pixel browser=Chrome ip=198.51.100.7 loc=Berlin_DE
[2026-10-05 09:09:00] user=guest action=weekly_loops_page device=iPhone browser=Safari ip=203.0.113.9 loc=Berlin_DE
[2026-10-05 09:08:00] user=guest action=weekly_loops_page ip=203.0.113.9
[2026-10-05 09:07:00] user=guest action=melodies_page device=iPhone browser=Safari ip=203.0.113.9 loc=Regensburg_DE
[2026-10-05 09:06:00] user=guest action=weekly_loops_page device=iPhone browser=Safari ip=192.168.1.20 loc=local
[2026-10-05 09:05:00] user=guest action=weekly_loops_page device=Mac browser=Chrome ip=203.0.113.10 loc=Hamburg_DE
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)

	// Three of the six lines survive: Regensburg and local are hidden and the
	// signup line is not a feed type.
	if len(evs) != 3 {
		t.Fatalf("got %d events, want 3: %+v", len(evs), evs)
	}

	// Page visit, full meta parsed the way melodies rows do it.
	if evs[0].Type != "weekly_loops_page" ||
		evs[0].Location == nil || *evs[0].Location != "Berlin_DE" ||
		evs[0].Device == nil || *evs[0].Device != "iPhone" ||
		evs[0].Browser == nil || *evs[0].Browser != "Safari" ||
		evs[0].IP == nil || *evs[0].IP != "203.0.113.9" {
		t.Errorf("event 0 = %+v, want weekly_loops_page Berlin_DE iPhone/Safari ip 203.0.113.9", evs[0])
	}

	// Sparse line: no device, browser or location, ip still parsed. A blank
	// location must never be treated as hidden.
	if evs[1].Type != "weekly_loops_page" || evs[1].Device != nil ||
		evs[1].Browser != nil || evs[1].Location != nil ||
		evs[1].IP == nil || *evs[1].IP != "203.0.113.9" {
		t.Errorf("event 1 = %+v, want weekly_loops_page with nil device, browser, location and ip 203.0.113.9", evs[1])
	}

	// A city that is not hidden still comes through.
	if evs[2].Type != "weekly_loops_page" ||
		evs[2].Location == nil || *evs[2].Location != "Hamburg_DE" {
		t.Errorf("event 2 = %+v, want weekly_loops_page Hamburg_DE", evs[2])
	}

	// The log reader must never emit a signup row again.
	for _, ev := range evs {
		if ev.Type == "email_list_subscribe" {
			t.Errorf("the log reader still emits signup rows: %+v", ev)
		}
	}
}

func TestReadMelodiesEventsBeatsAndSiteTag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	lines := `[2026-10-16 12:02:00] user=guest action=song_download song=foo bar.mp3 ip=203.0.113.7 site=melodies
[2026-10-16 12:01:00] user=guest action=song_play song=DJ DJ.mp3 device=Desktop browser=Chrome ip=203.0.113.7 site=beats
[2026-10-16 12:00:30] user=guest action=beat_page device=Desktop browser=Chrome ip=203.0.113.7
[2026-10-16 12:00:00] user=guest action=beats_page device=Desktop browser=Chrome ip=203.0.113.7 loc=Berlin_DE
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)
	// beat_page is the login gated page and must not reach the feed, so three
	// of the four lines survive.
	if len(evs) != 3 {
		t.Fatalf("got %d events, want 3: %+v", len(evs), evs)
	}

	// song= with the site tag last: the track must stop before it.
	if evs[0].Type != "track_downloaded" || evs[0].Track == nil || *evs[0].Track != "foo bar.mp3" ||
		evs[0].Site == nil || *evs[0].Site != "melodies" {
		t.Errorf("event 0 = %+v, want track_downloaded foo bar.mp3 site=melodies", evs[0])
	}

	// A beats play: its own site tag plus the usual fields.
	if evs[1].Type != "track_played" || evs[1].Track == nil || *evs[1].Track != "DJ DJ.mp3" ||
		evs[1].Site == nil || *evs[1].Site != "beats" || evs[1].Device == nil || *evs[1].Device != "Desktop" {
		t.Errorf("event 1 = %+v, want track_played DJ DJ.mp3 site=beats Desktop", evs[1])
	}

	// The public beats page view is its own type now.
	if evs[2].Type != "beats_visit" || evs[2].Location == nil || *evs[2].Location != "Berlin_DE" ||
		evs[2].Site != nil {
		t.Errorf("event 2 = %+v, want beats_visit Berlin_DE and no site tag", evs[2])
	}
}

func TestReadMelodiesEventsOldLinesHaveNoSite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	// The shape every line written before the site tag existed.
	lines := `[2026-09-01 09:00:00] user=guest action=song_play song=old.mp3 device=iPhone browser=Safari ip=1.2.3.4
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)
	if len(evs) != 1 {
		t.Fatalf("got %d events, want 1", len(evs))
	}
	if evs[0].Track == nil || *evs[0].Track != "old.mp3" {
		t.Errorf("track = %v, want old.mp3", evs[0].Track)
	}
	if evs[0].Site != nil {
		t.Errorf("site = %v, want nil for a line written before the tag existed", *evs[0].Site)
	}
}

func TestReadMelodiesEventsHidesSiteOwner(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "activity_log.txt")
	lines := `[2026-10-16 12:02:00] user=guest action=song_play song=visitor beat.mp3 device=Desktop browser=Chrome ip=203.0.113.8 site=beats
[2026-10-16 12:01:00] user=simon action=song_play song=simon played this.mp3 device=Desktop browser=Chrome ip=203.0.113.8 site=beats
[2026-10-16 12:00:30] user=Simon action=melodies_page device=Desktop browser=Chrome ip=203.0.113.8 loc=Berlin_DE
[2026-10-16 12:00:00] user=simon action=beats_page device=Desktop browser=Chrome ip=203.0.113.8 loc=Berlin_DE
[2026-10-16 11:59:00] user=g33cmin action=beats_page device=iPhone browser=Safari ip=203.0.113.9 loc=Berlin_DE
`
	if err := os.WriteFile(path, []byte(lines), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MELODIES_ACTIVITY_LOG", path)

	evs := readMelodiesEvents(10)
	// The three simon lines are dropped, the guest and the other account stay.
	if len(evs) != 2 {
		t.Fatalf("got %d events, want 2: %+v", len(evs), evs)
	}
	if evs[0].Type != "track_played" || evs[0].Track == nil || *evs[0].Track != "visitor beat.mp3" {
		t.Errorf("event 0 = %+v, want the guest track_played", evs[0])
	}
	// Another account is a real visitor and must survive.
	if evs[1].Type != "beats_visit" || evs[1].Device == nil || *evs[1].Device != "iPhone" {
		t.Errorf("event 1 = %+v, want the g33cmin beats_visit", evs[1])
	}
}
