package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsPrerelease(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"v0.20.2", false},
		{"v2.0.0", false},
		{"v2.0.0-beta.1", true},
		{"v2.0.0-beta.6", true},
		{"v1.2.3-alpha+build", true},
		{"v2.0.0-beta.2", true},
		{"0.1.0", false},
	}
	for _, tt := range tests {
		if got := isPrerelease(tt.tag); got != tt.want {
			t.Errorf("isPrerelease(%q) = %v, want %v", tt.tag, got, tt.want)
		}
	}
}

func TestSelectLatestStable(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		want    string
		wantErr bool
	}{
		{
			name: "stable upgrade available",
			tags: []string{"v0.19.0", "v0.20.0", "v0.20.2"},
			want: "v0.20.2",
		},
		{
			name: "stable does not follow betas",
			tags: []string{"v0.20.2", "v2.0.0-beta.4", "v2.0.0-beta.1"},
			want: "v0.20.2",
		},
		{
			name: "stable channel ignores betas",
			tags: []string{"v0.20.2", "v2.0.0-beta.4"},
			want: "v0.20.2",
		},
		{
			name: "newer stable version on remote (stable channel)",
			tags: []string{"v0.20.2", "v0.21.0", "v2.0.0-beta.4"},
			want: "v0.21.0",
		},
		{
			name:    "no stable tag found",
			tags:    []string{"v2.0.0-beta.4", "v2.0.0-beta.6"},
			wantErr: true,
		},
		{
			name: "skips invalid tags",
			tags: []string{"v0.20.2", "invalid", "v0.21.0"},
			want: "v0.21.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectLatestStable(tt.tags)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("selectLatestStable = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSelectLatestOverall(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "beta upgrade to newer beta",
			tags: []string{"v0.20.2", "v2.0.0-beta.2", "v2.0.0-beta.4"},
			want: "v2.0.0-beta.4",
		},
		{
			name: "beta sees highest overall including stable",
			tags: []string{"v2.0.0-beta.4", "v2.0.0"},
			want: "v2.0.0",
		},
		{
			name: "picks stable when newer than beta line",
			tags: []string{"v2.0.0-beta.6", "v2.1.0"},
			want: "v2.1.0",
		},
		{
			name: "single stable is latest",
			tags: []string{"v0.20.2"},
			want: "v0.20.2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectLatestOverall(tt.tags)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("selectLatestOverall = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSelectLatestForChannel(t *testing.T) {
	tests := []struct {
		name         string
		tags         []string
		localVersion string
		want         string
	}{
		{
			name:         "stable does not follow betas",
			tags:         []string{"v0.20.2", "v2.0.0-beta.4"},
			localVersion: "v0.20.2",
			want:         "v0.20.2",
		},
		{
			name:         "stable channel ignores betas",
			tags:         []string{"v0.20.2", "v2.0.0-beta.4"},
			localVersion: "v0.20.2",
			want:         "v0.20.2",
		},
		{
			name:         "beta channel sees betas",
			tags:         []string{"v0.20.2", "v2.0.0-beta.4"},
			localVersion: "v2.0.0-beta.2",
			want:         "v2.0.0-beta.4",
		},
		{
			name:         "beta upgrade to newer beta",
			tags:         []string{"v2.0.0-beta.2", "v2.0.0-beta.4"},
			localVersion: "v2.0.0-beta.2",
			want:         "v2.0.0-beta.4",
		},
		{
			name:         "beta upgrade to stable with same base",
			tags:         []string{"v2.0.0-beta.4", "v2.0.0"},
			localVersion: "v2.0.0-beta.2",
			want:         "v2.0.0",
		},
		{
			name:         "newer stable on remote stable channel",
			tags:         []string{"v0.20.2", "v0.21.0"},
			localVersion: "v0.20.2",
			want:         "v0.21.0",
		},
		{
			name:         "stable ignores higher beta line",
			tags:         []string{"v0.21.0", "v2.0.0-beta.6"},
			localVersion: "v0.20.2",
			want:         "v0.21.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectLatestForChannel(tt.tags, tt.localVersion)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("selectLatestForChannel(local=%q) = %q, want %q", tt.localVersion, got, tt.want)
			}
		})
	}
}

func TestCompareSemverPrerelease(t *testing.T) {
	tests := []struct {
		local, remote string
		want          int
	}{
		{"v2.0.0-beta.2", "v2.0.0", -1},       // prerelease < stable same base
		{"v2.0.0", "v2.0.0-beta.2", 1},        // stable > prerelease same base
		{"v2.0.0-beta.2", "v2.0.0-beta.4", -1}, // beta.2 < beta.4
		{"v2.0.0-beta.4", "v2.0.0-beta.2", 1},  // beta.4 > beta.2
		{"v2.0.0-beta.6", "v2.0.0-beta.6", 0},  // equal prereleases
		{"v2.0.0-beta.2", "v2.0.0-beta.2", 0},  // equal
		{"v0.20.2", "v0.20.2", 0},              // equal stables
		{"v0.20.2", "v2.0.0-beta.4", -1},       // lower major
		{"v2.0.0-beta.10", "v2.0.0-beta.6", 1}, // numeric prerelease 10 > 6
		{"v1.2.3-alpha", "v1.2.3-beta", -1},    // alpha < beta lexicographically
	}
	for _, tt := range tests {
		got, err := compareSemver(tt.local, tt.remote)
		if err != nil {
			t.Errorf("compareSemver(%q, %q): %v", tt.local, tt.remote, err)
			continue
		}
		if got != tt.want {
			t.Errorf("compareSemver(%q, %q) = %d, want %d", tt.local, tt.remote, got, tt.want)
		}
	}
}

func TestUpgradeChannelDecisions(t *testing.T) {
	tests := []struct {
		name         string
		localVersion string
		tags         []string
		shouldUpdate bool
	}{
		{
			name:         "stable upgrade available",
			localVersion: "v0.19.0",
			tags:         []string{"v0.19.0", "v0.20.0", "v0.20.2"},
			shouldUpdate: true,
		},
		{
			name:         "already up to date stable",
			localVersion: "v0.20.2",
			tags:         []string{"v0.20.2", "v0.20.0"},
			shouldUpdate: false,
		},
		{
			name:         "local version newer than remote",
			localVersion: "v2.0.0-beta.2",
			tags:         []string{"v0.20.2"},
			shouldUpdate: false,
		},
		{
			name:         "equal versions",
			localVersion: "v0.20.2",
			tags:         []string{"v0.20.2", "v0.19.0"},
			shouldUpdate: false,
		},
		{
			name:         "local version greater than remote",
			localVersion: "v1.0.0",
			tags:         []string{"v0.20.2", "v0.19.0"},
			shouldUpdate: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latest, err := selectLatestForChannel(tt.tags, tt.localVersion)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cmp, err := compareSemver(tt.localVersion, latest)
			if err != nil {
				t.Fatalf("compareSemver: %v", err)
			}
			got := cmp < 0
			if got != tt.shouldUpdate {
				t.Errorf("shouldUpdate(local=%q, latest=%q) = %v, want %v (cmp=%d)", tt.localVersion, latest, got, tt.shouldUpdate, cmp)
			}
		})
	}
}

func TestFetchTagNamesFiltering(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[{"name": "v0.20.2"}, {"name": ""}, {"name": "v2.0.0-beta.4"}]`))
	}))
	defer server.Close()

	tags, err := fetchTagNames(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := containsStr(tags, ""); ok {
		t.Error("expected empty tag name to be excluded")
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags (empty excluded), got %d: %v", len(tags), tags)
	}
}

func containsStr(slice []string, s string) (string, bool) {
	for _, v := range slice {
		if v == s {
			return v, true
		}
		if v != "" && strings.Contains(v, strings.Trim(s, " ")) && s == "" {
			// direct match handled above; this branch is unused but keeps strings import referenced
			_ = strings.Contains
		}
	}
	return "", false
}

func TestFetchTagNamesNetworkError(t *testing.T) {
	_, err := fetchTagNames("http://127.0.0.1:1")
	if err == nil {
		t.Error("expected error for network failure")
	}
	if !strings.Contains(err.Error(), "failed to check") {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}

func TestParseSemverPrereleaseAndBuildMeta(t *testing.T) {
	v, err := parseSemver("v1.2.3-beta.1+build.123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.prerelease != "beta.1" {
		t.Errorf("prerelease = %q, want %q", v.prerelease, "beta.1")
	}
	if v.major != 1 || v.minor != 2 || v.patch != 3 {
		t.Errorf("major/minor/patch = %d.%d.%d, want 1.2.3", v.major, v.minor, v.patch)
	}
}

func TestComparePrerelease_NumericVsAlpha(t *testing.T) {
	if comparePrerelease("10", "alpha") != -1 {
		t.Error("numeric prerelease segment should sort before alpha per semver")
	}
	if comparePrerelease("alpha", "10") != 1 {
		t.Error("alpha prerelease segment should sort after numeric per semver")
	}
}

func TestComparePrerelease_DotCount(t *testing.T) {
	if comparePrerelease("beta.1", "beta") != 1 {
		t.Error("beta.1 should be greater than beta (more identifiers)")
	}
	if comparePrerelease("beta", "beta.1") != -1 {
		t.Error("beta should be less than beta.1")
	}
}
