package buildversion

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{
			name:    "empty",
			version: "",
			want:    "v0.0.0-unknown",
		},
		{
			name:    "version string",
			version: "v0.1.2",
			want:    "v0.1.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			version = tt.version
			initBuild()

			got := Version()
			req.Equal(tt.want, got)
		})
	}
}

func TestGitSHA(t *testing.T) {
	tests := []struct {
		name string
		sha  string
		want string
	}{
		{
			name: "empty",
			sha:  "",
			want: "",
		},
		{
			name: "too short",
			sha:  "123456",
			want: "",
		},
		{
			name: "7 chars",
			sha:  "1234567",
			want: "1234567",
		},
		{
			name: "full sha",
			sha:  "e21cf800acca2aa972e7f5f65f7134b5da92f05f",
			want: "e21cf80",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			gitSHA = tt.sha
			initBuild()

			got := GitSHA()
			req.Equal(tt.want, got)
		})
	}
}

func TestBuildTime(t *testing.T) {
	req := require.New(t)
	aTime, err := time.Parse(time.RFC3339, "2019-06-26T18:53:19Z")
	req.NoError(err, "parse constant time")

	tests := []struct {
		name       string
		timestring string
		want       time.Time
	}{
		{
			name:       "empty",
			timestring: "",
			want:       time.Time{},
		},
		{
			name:       "proper format",
			timestring: "2019-06-26T18:53:19Z",
			want:       aTime,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			buildTime = tt.timestring
			initBuild()

			got := BuildTime()
			req.Equal(tt.want, got)
		})
	}
}

func TestGetBuild(t *testing.T) {
	tests := []struct {
		name      string
		gitSHA    string
		version   string
		buildTime string
		want      Build
	}{
		{
			name:   "goInfo",
			gitSHA: "12345678",
			want: Build{
				GitSHA: "1234567",
				GoInfo: getGoInfo(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			version = tt.version
			gitSHA = tt.gitSHA
			buildTime = tt.buildTime

			initBuild()

			got := GetBuild()
			got.RunAt = nil
			req.Equal(tt.want, got)
		})
	}
}

func TestIsLatestRelease(t *testing.T) {
	tests := []struct {
		name            string
		version         string
		upstreamVersion string
		isLatest        bool
		latest          string
		wantErr         bool
	}{
		{
			name:            "not the latest",
			version:         "v0.9.1",
			upstreamVersion: "v0.10.0",
			isLatest:        false,
			latest:          "v0.10.0",
		},
		{
			name:            "local version",
			version:         "v0.10.0-dirty",
			upstreamVersion: "v0.10.0",
			isLatest:        false,
			latest:          "v0.10.0",
		},
		{
			name:            "actually the latest version",
			version:         "v0.10.1",
			upstreamVersion: "v0.10.1",
			isLatest:        true,
			latest:          "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.upstreamVersion))
			})
			server := httptest.NewTLSServer(h)
			defer server.Close()
			client := server.Client()
			upstreamURL := server.URL

			version = tt.version
			initBuild()

			isLatest, latest, err := isLatestRelease(client, upstreamURL)
			req.NoError(err)

			req.Equal(tt.isLatest, isLatest)
			req.Equal(tt.latest, latest)
		})
	}
}
