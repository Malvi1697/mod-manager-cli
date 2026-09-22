package thunderstore

import (
	"strconv"
	"strings"
)

const hexiumBaseURL = "https://valheim.hexium.gg"

// Registry is a Thunderstore-compatible package registry.
//
// Hexium serves the same experimental API shape as Thunderstore, so both are
// queried through the same code path. Several Valheim mod authors publish
// their newest builds only on Hexium, which a Thunderstore-only lookup reports
// as a missing package or a missing version.
type Registry struct {
	// Name identifies the registry in errors and user-facing output.
	Name string
	// BaseURL is the registry root, without a trailing slash.
	BaseURL string
	// URLHosts are the hosts whose package URLs belong to this registry.
	URLHosts []string
	// URLMarker is the path segment directly before /{owner}/{name} in a
	// package URL: thunderstore.io/c/valheim/p/Owner/Name and
	// valheim.hexium.gg/mods/Owner/Name.
	URLMarker string
}

// registries are queried in this order. Thunderstore stays first so that it
// wins ties and remains the source for everything it already served.
var registries = []Registry{
	{
		Name:      "Thunderstore",
		BaseURL:   baseURL,
		URLHosts:  []string{"thunderstore.io"},
		URLMarker: "p",
	},
	{
		Name:      "Hexium",
		BaseURL:   hexiumBaseURL,
		URLHosts:  []string{"valheim.hexium.gg", "hexium.gg"},
		URLMarker: "mods",
	},
}

// Registries returns the configured registries in preference order.
func Registries() []Registry {
	return registries
}

func (r Registry) experimentalAPI() string {
	return r.BaseURL + "/api/experimental/package/"
}

// registryForURL returns the registry a package URL belongs to.
func registryForURL(rawURL string) (Registry, bool) {
	trimmed := strings.TrimPrefix(strings.TrimPrefix(rawURL, "https://"), "http://")
	if trimmed == rawURL {
		return Registry{}, false
	}
	host := trimmed
	if i := strings.IndexByte(host, '/'); i >= 0 {
		host = host[:i]
	}
	for _, r := range registries {
		for _, h := range r.URLHosts {
			if host == h {
				return r, true
			}
		}
	}
	return Registry{}, false
}

// ownerNameFromURL extracts the owner and name from a package URL belonging to
// r. It reports false when the URL does not carry both segments.
func ownerNameFromURL(r Registry, rawURL string) (owner, name string, ok bool) {
	parts := strings.Split(strings.Trim(rawURL, "/"), "/")
	for i, p := range parts {
		if p == r.URLMarker && i+2 < len(parts) {
			return parts[i+1], parts[i+2], true
		}
	}
	return "", "", false
}

// latestVersionNumber returns the version number a lookup resolved to.
func latestVersionNumber(pkg *Package) string {
	if pkg == nil || len(pkg.Versions) == 0 {
		return ""
	}
	return pkg.Versions[0].VersionNumber
}

// compareVersions compares two dotted numeric version strings. It returns a
// negative value when a sorts before b, zero when they are equal and a
// positive value when a sorts after b. Segments that are missing or not
// numeric count as zero, matching the numeric triples both registries publish.
func compareVersions(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		if d := versionSegment(as, i) - versionSegment(bs, i); d != 0 {
			return d
		}
	}
	return 0
}

func versionSegment(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	v, err := strconv.Atoi(strings.TrimSpace(parts[i]))
	if err != nil {
		return 0
	}
	return v
}
