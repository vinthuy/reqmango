package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCurrent_ReportsDefaults(t *testing.T) {
	defer func(v, c, b string) {
		Version, Commit, BuildDate = v, c, b
	}(Version, Commit, BuildDate)

	Version, Commit, BuildDate = "dev", "unknown", "unknown"

	info := Current()
	require.Equal(t, "dev", info.Version)
	require.Equal(t, "unknown", info.Commit)
	require.Equal(t, "unknown", info.BuildDate)
	require.True(t, strings.HasPrefix(info.GoVersion, "go"), "go_version = %q", info.GoVersion)
}

func TestCurrent_ReportsInjectedValues(t *testing.T) {
	defer func(v, c, b string) {
		Version, Commit, BuildDate = v, c, b
	}(Version, Commit, BuildDate)

	Version = "v0.1.0"
	Commit = "dd079de"
	BuildDate = "2026-10-24T10:00:00Z"

	info := Current()
	require.Equal(t, "v0.1.0", info.Version)
	require.Equal(t, "dd079de", info.Commit)
	require.Equal(t, "2026-10-24T10:00:00Z", info.BuildDate)
}
