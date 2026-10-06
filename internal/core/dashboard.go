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
// at lim (clamped to 1-100).
func (c *Core) GetDashboardEvents(lim int) ([]models.DashboardEvent, error) {
	if lim < 1 {
		lim = 5
	}
	if lim > 100 {
		lim = 100
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
// Sign ups (email_list_subscribe) are deliberately NOT here: listmonk records
// its own optin row for the same event, with the address and list name on it,
// so mapping the log line too showed every signup twice.
var melodiesActionTypes = map[string]string{
	"melodies_page":      "site_visit",
	"song_play":          "track_played",
	"song_download":      "track_downloaded",
	"portfolio_visit":    "portfolio_visit",
	"portfolio_referral": "portfolio_referral",
	"weekly_loops_page":  "weekly_loops_page",
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
		if len(out) >= lim || scanned >= 2000 {
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

		ts, err := time.ParseInLocation("2006-01-02 15:04:05", m[1], time.Local)
		if err != nil {
			continue
		}

		ev := models.DashboardEvent{Type: typ, CreatedAt: ts}
		if typ == "track_played" || typ == "track_downloaded" {
			if t := parseMelodiesTrack(m[4]); t != "" {
				ev.Track = &t
			}
		}
		if d := parseMelodiesMetaToken(m[4], "device"); d != "" {
			ev.Device = &d
		}
		if b := parseMelodiesMetaToken(m[4], "browser"); b != "" {
			ev.Browser = &b
		}
		if l := parseMelodiesMetaToken(m[4], "loc"); l != "" {
			if isHiddenLocation(l) {
				continue
			}
			ev.Location = &l
		}
		if r := parseMelodiesMetaToken(m[4], "ref"); r != "" {
			ev.Ref = &r
		}
		if p := parseMelodiesMetaToken(m[4], "path"); p != "" {
			ev.Path = &p
		}
		if ip := parseMelodiesMetaToken(m[4], "ip"); ip != "" {
			ev.IP = &ip
		}
		out = append(out, ev)
	}

	return out
}

// hiddenLocationCities lists the cities whose visits are kept out of the
// dashboard feed because they always come from Simon's own machine. The
// activity log keeps every line untouched, this only filters at read time.
var hiddenLocationCities = []string{"Regensburg"}

// hiddenLocationTokens are loc= values that can never describe a real visitor.
// "local" is what the site writes for a private or loopback address, so the
// only people it can be are on the Pi's own network, meaning Simon himself.
var hiddenLocationTokens = []string{"local"}

// isHiddenLocation reports whether a loc= token such as Regensburg_DE belongs
// to a hidden city, or is one of the private network tokens. For a city only
// the part before the last underscore is compared, case insensitively. Lines
// without a location are never hidden.
func isHiddenLocation(loc string) bool {
	if loc == "" {
		return false
	}
	for _, tok := range hiddenLocationTokens {
		if strings.EqualFold(loc, tok) {
			return true
		}
	}
	city := loc
	if i := strings.LastIndex(loc, "_"); i != -1 {
		city = loc[:i]
	}
	for _, c := range hiddenLocationCities {
		if strings.EqualFold(city, c) {
			return true
		}
	}
	return false
}

// parseMelodiesTrack extracts the song= value from the trailing meta of a
// log line. Values may contain spaces and stop at the first known follower
// pair (ip, device, browser).
func parseMelodiesTrack(meta string) string {
	i := strings.Index(meta, "song=")
	if i == -1 {
		return ""
	}
	rest := meta[i+len("song="):]
	for _, stop := range []string{" ip=", " device=", " browser="} {
		if j := strings.Index(rest, stop); j != -1 {
			rest = rest[:j]
		}
	}
	return strings.TrimSpace(rest)
}

// melodiesMetaTokenRes compiles the single token fields (device, browser)
// that newer PI-Website log lines carry in their trailing meta.
var melodiesMetaTokenRes = map[string]*regexp.Regexp{
	"device":  regexp.MustCompile(`(?:^|\s)device=(\S+)`),
	"browser": regexp.MustCompile(`(?:^|\s)browser=(\S+)`),
	"loc":     regexp.MustCompile(`(?:^|\s)loc=(\S+)`),
	"ref":     regexp.MustCompile(`(?:^|\s)ref=(\S+)`),
	"path":    regexp.MustCompile(`(?:^|\s)path=(\S+)`),
	"ip":      regexp.MustCompile(`(?:^|\s)ip=(\S+)`),
}

// parseMelodiesMetaToken extracts a single token value such as device=iPhone
// from the trailing meta of a log line. Lines written before the field
// existed return an empty string.
func parseMelodiesMetaToken(meta, key string) string {
	re, ok := melodiesMetaTokenRes[key]
	if !ok {
		return ""
	}
	m := re.FindStringSubmatch(meta)
	if m == nil {
		return ""
	}
	return m[1]
}
