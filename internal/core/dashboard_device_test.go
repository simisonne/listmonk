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
