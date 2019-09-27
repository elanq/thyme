package thyme

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// Snapshot represents the current state of all in-use application
// windows at a moment in time.
type Snapshot struct {
	Time    time.Time
	Windows []*Window
	Active  int64
	Visible []int64
}

// Print returns a pretty-printed representation of the snapshot.
func (s Snapshot) Print() string {
	var b bytes.Buffer

	var active *Window
	visible := make([]*Window, 0, len(s.Windows))
	other := make([]*Window, 0, len(s.Windows))
s_Windows:
	for _, w := range s.Windows {
		if w.ID == s.Active {
			active = w
			continue s_Windows
		}
		for _, v := range s.Visible {
			if w.ID == v {
				visible = append(visible, w)
				continue s_Windows
			}
		}
		other = append(other, w)
	}

	fmt.Fprintf(&b, "%s\n", s.Time.Format("Mon Jan 2 15:04:05 -0700 MST 2006"))
	if active != nil {
		fmt.Fprintf(&b, "\tActive: %s\n", active.Info().Print())
	}
	if len(visible) > 0 {
		fmt.Fprintf(&b, "\tVisible: ")
		for _, v := range visible {
			fmt.Fprintf(&b, "%s, ", v.Info().Print())
		}
		fmt.Fprintf(&b, "\n")
	}
	if len(other) > 0 {
		fmt.Fprintf(&b, "\tOther: ")
		for _, w := range other {
			fmt.Fprintf(&b, "%s, ", w.Info().Print())
		}
		fmt.Fprintf(&b, "\n")
	}
	return string(b.Bytes())
}

// Window represents an application window.
type Window struct {
	// ID is the numerical identifier of the window.
	ID int64

	// Desktop is the numerical identifier of the desktop the
	// window belongs to.  Equal to -1 if the window is sticky.
	Desktop int64

	// Name is the display name of the window (typically what the
	// windowing system shows in the top bar of the window).
	Name string
}

// systemNames is a set of blacklisted window names that are known to
// be used by system windows that aren't visible to the user.
var systemNames = map[string]struct{}{
	"XdndCollectionWindowImp": {},
	"unity-launcher":          {},
	"unity-panel":             {},
	"unity-dash":              {},
	"Hud":                     {},
	"Desktop":                 {},
}

// IsSystem returns true if the window is a system window (like
// "unity-panel" and thus shouldn't be considered an application
// visible to the end-users)
func (w *Window) IsSystem() bool {
	if _, is := systemNames[w.Name]; is {
		return true
	}
	return false
}

// IsSticky returns true if the window is a sticky window (i.e.
// present on all desktops)
func (w *Window) IsSticky() bool {
	return w.Desktop == -1
}

// IsOnDesktop returns true if the window is present on the specified
// desktop
func (w *Window) IsOnDesktop(desktop int64) bool {
	return w.IsSticky() || w.Desktop == desktop
}

const (
	defaultWindowTitleSeparator       = " - "
	microsoftEdgeWindowTitleSeparator = "\u200e- "
)

// Info returns more structured metadata about a window. The metadata
// is extracted using heuristics.
//
// Assumptions:
//     1) Most windows use " - " to separate their window names from their content
//     2) Most windows use the " - " with the application name at the end.
//     3) The few programs that reverse this convention only reverse it.
func (w *Window) Info() *Winfo {
	// Special Cases
	fields := strings.Split(w.Name, defaultWindowTitleSeparator)
	if len(fields) > 1 {
		last := strings.TrimSpace(fields[len(fields)-1])
		if last == "Google Chrome" {
			return &Winfo{
				App:    "Google Chrome",
				SubApp: strings.TrimSpace(fields[len(fields)-2]),
				Title:  strings.Join(fields[0:len(fields)-2], defaultWindowTitleSeparator),
			}
		}
	}

	if strings.Contains(w.Name, microsoftEdgeWindowTitleSeparator) {
		// App Name Last
		beforeSep := strings.LastIndex(w.Name, microsoftEdgeWindowTitleSeparator)
		afterSep := beforeSep + len(microsoftEdgeWindowTitleSeparator)
		return &Winfo{
			App:   strings.TrimSpace(w.Name[afterSep:]),
			Title: strings.TrimSpace(w.Name[:beforeSep]),
		}
	}

	// Normal Cases
	if beforeSep := strings.Index(w.Name, defaultWindowTitleSeparator); beforeSep > -1 {
		// App Name First
		if w.Name[:beforeSep] == "Slack" {
			afterSep := beforeSep + len(defaultWindowTitleSeparator)
			return &Winfo{
				App:   strings.TrimSpace(w.Name[:beforeSep]),
				Title: strings.TrimSpace(w.Name[afterSep:]),
			}
		}

		// App Name Last
		beforeSep := strings.LastIndex(w.Name, defaultWindowTitleSeparator)
		afterSep := beforeSep + len(defaultWindowTitleSeparator)
		return &Winfo{
			App:   strings.TrimSpace(w.Name[afterSep:]),
			Title: strings.TrimSpace(w.Name[:beforeSep]),
		}
	}

	// No Application name separator
	return &Winfo{
		Title: w.Name,
	}
}

// Winfo is structured metadata info about a window.
type Winfo struct {
	// App is the application that controls the window.
	App string

	// SubApp is the sub-application that controls the window. An
	// example is a web app (e.g., Sourcegraph) that runs
	// inside a Chrome tab. In this case, the App field would be
	// "Google Chrome" and the SubApp field would be "Sourcegraph".
	SubApp string

	// Title is the title of the window after the App and SubApp name
	// have been stripped.
	Title string
}

// Print returns a pretty-printed representation of the snapshot.
func (w Winfo) Print() string {
	return fmt.Sprintf("[%s|%s|%s]", w.App, w.SubApp, w.Title)
}
