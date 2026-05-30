package analytics

import "strings"

const (
	BrowserChrome  = "Chrome"
	BrowserFirefox = "Firefox"
	BrowserSafari  = "Safari"
	BrowserEdge    = "Edge"
	BrowserOpera   = "Opera"
	BrowserOther   = "Other"
)

func ClassifyBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case strings.Contains(ua, "opr/") || strings.Contains(ua, "opios/") || strings.Contains(ua, "opera"):
		return BrowserOpera
	case strings.Contains(ua, "edg/") || strings.Contains(ua, "edga/") || strings.Contains(ua, "edgios/") || strings.Contains(ua, "edge/"):
		return BrowserEdge
	case strings.Contains(ua, "firefox/") || strings.Contains(ua, "fxios/"):
		return BrowserFirefox
	case strings.Contains(ua, "chrome/") || strings.Contains(ua, "crios/"):
		return BrowserChrome
	case strings.Contains(ua, "safari/") && !strings.Contains(ua, "chrome/") && !strings.Contains(ua, "crios/"):
		return BrowserSafari
	default:
		return BrowserOther
	}
}

func emptyBrowserStats() map[string]int64 {
	return map[string]int64{
		BrowserChrome:  0,
		BrowserFirefox: 0,
		BrowserSafari:  0,
		BrowserEdge:    0,
		BrowserOpera:   0,
		BrowserOther:   0,
	}
}
