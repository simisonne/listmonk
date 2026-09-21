package core

import (
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetDashboardCharts returns chart data points to render on the dashboard.
func (c *Core) GetDashboardCharts() (types.JSONText, error) {
	_ = c.refreshCache(matDashboardCharts, false)

	var out types.JSONText
	if err := c.q.GetDashboardCharts.Get(&out); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "dashboard charts", "error", pqErrMsg(err)))
	}

	return out, nil
}

// GetDashboardCounts returns stats counts to show on the dashboard.
func (c *Core) GetDashboardCounts() (types.JSONText, error) {
	_ = c.refreshCache(matDashboardCounts, false)

	var out types.JSONText
	if err := c.q.GetDashboardCounts.Get(&out); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "dashboard stats", "error", pqErrMsg(err)))
	}

	return out, nil
}

// GetDashboardEvents returns the newest dashboard events (campaign sends,
// opens, clicks and public list opt-ins from the DB, merged with melodies
// site activity from the PI-Website activity log), newest first, capped
// at lim (clamped to 1-20).
func (c *Core) GetDashboardEvents(lim int) ([]models.DashboardEvent, error) {
	if lim < 1 {
		lim = 5
	}
	if lim > 20 {
		lim = 20
	}

	out := []models.DashboardEvent{}
	// Ask the DB for extra rows so interleaved melodies events cannot
	// starve any DB event type out of the merged feed.
	if err := c.q.GetDashboardEvents.Select(&out, lim*3); err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "dashboard events", "error", pqErrMsg(err)))
	}

	out = append(out, readMelodiesEvents(lim)...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	if len(out) > lim {
		out = out[:lim]
	}

	return out, nil
}

// melodiesLogLine matches PI-Website activity_log.txt lines:
// [2006-01-02 15:04:05] user=guest action=song_play song=foo.mp3 ip=1.2.3.4
var melodiesLogLine = regexp.MustCompile(`^\[(.*?)\]\s+user=(\S*)\s+action=(\S*)(.*)$`)

// melodiesActionTypes maps PI-Website log actions to dashboard event types.
var melodiesActionTypes = map[string]string{
	"melodies_page": "site_visit",
	"song_play":     "track_played",
	"song_download": "track_downloaded",
}

// readMelodiesEvents tails the PI-Website activity log (newest lines first)
// and returns up to lim melodies site events. A missing or unreadable log
// simply yields no events.
func readMelodiesEvents(lim int) []models.DashboardEvent {
	path := os.Getenv("MELODIES_ACTIVITY_LOG")
	if path == "" {
		path = "/home/simon/shared/PI-Website/activity_log.txt"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	out := []models.DashboardEvent{}
	scanned := 0
	for _, line := range strings.Split(string(data), "\n") {
		if len(out) >= lim || scanned >= 500 {
			break
		}
		scanned++

		m := melodiesLogLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		typ, ok := melodiesActionTypes[m[3]]
		if !ok {
			continue
		}

		ts, err := time.Parse("2006-01-02 15:04:05", m[1])
		if err != nil {
			continue
		}

		ev := models.DashboardEvent{Type: typ, CreatedAt: ts}
		if typ == "track_played" || typ == "track_downloaded" {
			if t := parseMelodiesTrack(m[4]); t != "" {
				ev.Track = &t
			}
		}
		out = append(out, ev)
	}

	return out
}

// parseMelodiesTrack extracts the song= value from the trailing meta of a
// log line. Values may contain spaces and always precede the ip= pair.
func parseMelodiesTrack(meta string) string {
	i := strings.Index(meta, "song=")
	if i == -1 {
		return ""
	}
	rest := meta[i+len("song="):]
	if j := strings.Index(rest, " ip="); j != -1 {
		rest = rest[:j]
	}
	return strings.TrimSpace(rest)
}
