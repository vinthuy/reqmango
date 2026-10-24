// Package version exposes the build identity of the running server.
//
// The values are injected at build time, e.g.
//
//	go build -ldflags "-X github.com/reqmango/backend/internal/version.Version=v0.1.0 \
//	  -X github.com/reqmango/backend/internal/version.Commit=$(git rev-parse --short HEAD) \
//	  -X github.com/reqmango/backend/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
//
// A binary built without them reports "dev", which is what a local `go run`
// should say.
package version

import "runtime"

var (
	// Version is the release tag, without or with the leading "v" as the
	// release workflow passes it.
	Version = "dev"
	// Commit is the source revision the binary was built from.
	Commit = "unknown"
	// BuildDate is an RFC 3339 UTC timestamp.
	BuildDate = "unknown"
)

// Info is the JSON payload of GET /version.
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
}

// Current returns the build identity of this process.
func Current() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
	}
}
