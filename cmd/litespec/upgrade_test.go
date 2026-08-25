package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

// --- Scenario test infrastructure ---

type upgradeTestState struct {
	installCalled   bool
	installModule   string
	installTag      string
	installErr      error
	bgCalled        bool
	bgModule        string
	bgTag           string
	modulePathErr   error
	modulePathValue string
	stdout          *bytes.Buffer
	stdoutW         io.WriteCloser
	stdoutR         io.Reader
}

func (s *upgradeTestState) flushStdout() {
	if s.stdoutW == nil {
		return
	}
	s.stdoutW.Close()
	s.stdoutW = nil
	io.Copy(s.stdout, s.stdoutR)
}

func githubTagsJSON(names []string) []byte {
	tags := make([]githubTag, len(names))
	for i, n := range names {
		tags[i] = githubTag{Name: n}
	}
	b, _ := json.Marshal(tags)
	return b
}

func setupUpgradeTestHandler(t *testing.T, handler http.HandlerFunc, localVersion string) (*upgradeTestState, func()) {
	t.Helper()

	server := httptest.NewServer(handler)

	gobinDir := t.TempDir()
	exePath := filepath.Join(gobinDir, "litespec")
	if err := os.WriteFile(exePath, nil, 0o755); err != nil {
		t.Fatal(err)
	}

	homeDir := t.TempDir()
	cacheDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CACHE_HOME", cacheDir)
	t.Setenv("GOBIN", gobinDir)
	t.Setenv("GOPATH", "")

	state := &upgradeTestState{
		modulePathValue: "github.com/test/mod",
		stdout:          &bytes.Buffer{},
	}

	oldTagsURL := tagsURL
	oldExeFn := executableFn
	oldModuleFn := modulePathFn
	oldRunInstall := runGoInstall
	oldStartBg := startBackgroundInstall
	oldVersion := version

	tagsURL = server.URL
	executableFn = func() (string, error) { return exePath, nil }
	modulePathFn = func() (string, error) { return state.modulePathValue, state.modulePathErr }
	runGoInstall = func(mp, tag string) error {
		state.installCalled = true
		state.installModule = mp
		state.installTag = tag
		return state.installErr
	}
	startBackgroundInstall = func(mp, tag string) {
		state.bgCalled = true
		state.bgModule = mp
		state.bgTag = tag
	}
	version = localVersion

	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		t.Fatal(pipeErr)
	}
	os.Stdout = w
	state.stdoutW = w
	state.stdoutR = r

	cleanup := func() {
		state.flushStdout()
		tagsURL = oldTagsURL
		executableFn = oldExeFn
		modulePathFn = oldModuleFn
		runGoInstall = oldRunInstall
		startBackgroundInstall = oldStartBg
		version = oldVersion
		os.Stdout = oldStdout
		server.Close()
	}

	return state, cleanup
}

func setupUpgradeTest(t *testing.T, tags []string, localVersion string) (*upgradeTestState, func()) {
	t.Helper()
	return setupUpgradeTestHandler(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(githubTagsJSON(tags))
	}, localVersion)
}

func stampDirPath(t *testing.T) string {
	t.Helper()
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(cacheDir, "litespec")
}

func stampFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(stampDirPath(t), "last-update-check")
}

// --- Requirement: explicit upgrade command ---

func TestUpgradeCommandScenarios(t *testing.T) {
	t.Run("stable upgrade available", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.19.0", "v0.20.0", "v0.20.2"}, "0.19.0")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		if state.installTag != "v0.20.2" {
			t.Errorf("install tag = %q, want v0.20.2", state.installTag)
		}
		if !strings.Contains(state.stdout.String(), "v0.20.2") {
			t.Errorf("expected output to contain v0.20.2, got: %s", state.stdout.String())
		}
	})

	t.Run("beta upgrade to newer beta", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v2.0.0-beta.2", "v2.0.0-beta.4"}, "2.0.0-beta.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		if state.installTag != "v2.0.0-beta.4" {
			t.Errorf("install tag = %q, want v2.0.0-beta.4", state.installTag)
		}
	})

	t.Run("beta upgrade to stable", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v2.0.0-beta.2", "v2.0.0"}, "2.0.0-beta.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		if state.installTag != "v2.0.0" {
			t.Errorf("install tag = %q, want v2.0.0", state.installTag)
		}
	})

	t.Run("already up to date", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.19.0"}, "0.20.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if state.installCalled {
			t.Error("expected go install NOT to be called")
		}
		if !strings.Contains(state.stdout.String(), "Already up to date") {
			t.Errorf("expected 'Already up to date' in output, got: %s", state.stdout.String())
		}
	})

	t.Run("local version newer than remote", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2"}, "2.0.0-beta.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if state.installCalled {
			t.Error("expected go install NOT to be called")
		}
		if !strings.Contains(state.stdout.String(), "Already up to date") {
			t.Errorf("expected 'Already up to date', got: %s", state.stdout.String())
		}
	})

	t.Run("stable does not follow betas", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v2.0.0-beta.4", "v2.0.0-beta.1"}, "0.20.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if state.installCalled {
			t.Error("expected go install NOT to be called (stable ignores betas)")
		}
		if !strings.Contains(state.stdout.String(), "Already up to date") {
			t.Errorf("expected 'Already up to date', got: %s", state.stdout.String())
		}
	})

	t.Run("not installed via go install", func(t *testing.T) {
		t.Setenv("GOBIN", "/nonexistent")
		t.Setenv("GOPATH", "/nonexistent")
		t.Setenv("HOME", t.TempDir())

		err := cmdUpgrade([]string{})
		if err == nil {
			t.Fatal("expected error for non-go-install binary")
		}
		if !strings.Contains(err.Error(), "go install") {
			t.Errorf("expected 'go install' in error, got: %v", err)
		}
	})

	t.Run("go install failure", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()
		state.installErr = fmt.Errorf("go install failed: exit status 1")

		err := cmdUpgrade([]string{})
		if err == nil {
			t.Fatal("expected error from go install failure")
		}
		if !state.installCalled {
			t.Error("expected go install to be called")
		}
	})

	t.Run("network error fetching latest version", func(t *testing.T) {
		state, cleanup := setupUpgradeTestHandler(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}, "0.20.2")
		defer cleanup()

		err := cmdUpgrade([]string{})
		if err == nil {
			t.Fatal("expected error for network failure")
		}
		if !strings.Contains(err.Error(), "failed to check") {
			t.Errorf("expected 'failed to check' in error, got: %v", err)
		}
		if state.installCalled {
			t.Error("expected go install NOT to be called")
		}
	})
}

// --- Requirement: go install detection ---

func TestGoInstallDetectionScenarios(t *testing.T) {
	t.Run("binary in GOBIN", func(t *testing.T) {
		dir := t.TempDir()
		exePath := filepath.Join(dir, "litespec")
		if err := os.WriteFile(exePath, nil, 0o755); err != nil {
			t.Fatal(err)
		}

		oldFn := executableFn
		defer func() { executableFn = oldFn }()
		executableFn = func() (string, error) { return exePath, nil }

		t.Setenv("GOBIN", dir)
		t.Setenv("GOPATH", "")

		if !isGoInstall() {
			t.Error("expected true for binary in GOBIN")
		}
	})

	t.Run("binary in GOPATH/bin", func(t *testing.T) {
		dir := t.TempDir()
		gobinDir := filepath.Join(dir, "bin")
		if err := os.MkdirAll(gobinDir, 0o755); err != nil {
			t.Fatal(err)
		}
		exePath := filepath.Join(gobinDir, "litespec")
		if err := os.WriteFile(exePath, nil, 0o755); err != nil {
			t.Fatal(err)
		}

		oldFn := executableFn
		defer func() { executableFn = oldFn }()
		executableFn = func() (string, error) { return exePath, nil }

		t.Setenv("GOBIN", "")
		t.Setenv("GOPATH", dir)

		if !isGoInstall() {
			t.Error("expected true for binary in GOPATH/bin")
		}
	})

	t.Run("binary elsewhere", func(t *testing.T) {
		dir := t.TempDir()
		exePath := filepath.Join(dir, "litespec")
		if err := os.WriteFile(exePath, nil, 0o755); err != nil {
			t.Fatal(err)
		}

		oldFn := executableFn
		defer func() { executableFn = oldFn }()
		executableFn = func() (string, error) { return exePath, nil }

		t.Setenv("GOBIN", "/nonexistent")
		t.Setenv("GOPATH", "/nonexistent")

		if isGoInstall() {
			t.Error("expected false for binary outside GOBIN/GOPATH/bin")
		}
	})
}

// --- Requirement: module path discovery ---

func TestModulePathDiscoveryScenarios(t *testing.T) {
	t.Run("module path resolved", func(t *testing.T) {
		path, err := getModulePath()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if path == "" {
			t.Error("expected non-empty module path")
		}
	})

	t.Run("module path unavailable", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()
		state.modulePathErr = fmt.Errorf("could not determine module path from build info")

		err := cmdUpgrade([]string{})
		if err == nil {
			t.Fatal("expected error for unavailable module path")
		}
		if !strings.Contains(err.Error(), "module path") {
			t.Errorf("expected 'module path' in error, got: %v", err)
		}
		if state.installCalled {
			t.Error("expected go install NOT to be called")
		}
	})
}

// --- Requirement: version comparison ---

func TestVersionComparisonScenarios(t *testing.T) {
	t.Run("newer stable version on remote (stable channel)", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0", "v2.0.0-beta.4"}, "0.20.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		if state.installTag != "v0.21.0" {
			t.Errorf("install tag = %q, want v0.21.0", state.installTag)
		}
	})

	t.Run("equal versions", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.19.0"}, "0.20.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if state.installCalled {
			t.Error("expected go install NOT to be called")
		}
		if !strings.Contains(state.stdout.String(), "Already up to date") {
			t.Errorf("expected 'Already up to date', got: %s", state.stdout.String())
		}
	})

	t.Run("local version greater than remote", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.19.0"}, "1.0.0")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if state.installCalled {
			t.Error("expected go install NOT to be called")
		}
		if !strings.Contains(state.stdout.String(), "Already up to date") {
			t.Errorf("expected 'Already up to date', got: %s", state.stdout.String())
		}
	})

	t.Run("stable channel ignores betas", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v2.0.0-beta.4"}, "0.20.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.installCalled {
			t.Error("expected go install NOT to be called (stable ignores betas)")
		}
	})

	t.Run("beta channel sees betas", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v2.0.0-beta.4"}, "2.0.0-beta.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		if state.installTag != "v2.0.0-beta.4" {
			t.Errorf("install tag = %q, want v2.0.0-beta.4", state.installTag)
		}
	})

	t.Run("beta versus stable with same base", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v2.0.0-beta.4", "v2.0.0"}, "2.0.0-beta.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		if state.installTag != "v2.0.0" {
			t.Errorf("install tag = %q, want v2.0.0", state.installTag)
		}
	})
}

// --- Requirement: post-upgrade hint ---

func TestPostUpgradeHintScenarios(t *testing.T) {
	t.Run("upgrade succeeds", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()

		if err := cmdUpgrade([]string{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		state.flushStdout()
		if !state.installCalled {
			t.Fatal("expected go install to be called")
		}
		output := state.stdout.String()
		if !strings.Contains(output, "v0.21.0") {
			t.Errorf("expected new version in output, got: %s", output)
		}
		if !strings.Contains(output, "litespec update") {
			t.Errorf("expected 'litespec update' hint in output, got: %s", output)
		}
	})
}

// --- Requirement: background self-update gate ---

func TestBackgroundSelfUpdateGateScenarios(t *testing.T) {
	t.Run("check interval elapsed", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()

		stampDir := stampDirPath(t)
		if err := os.MkdirAll(stampDir, 0o755); err != nil {
			t.Fatal(err)
		}
		stampFile := stampFilePath(t)
		if err := os.WriteFile(stampFile, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		oldTime := time.Now().Add(-8 * 24 * time.Hour)
		if err := os.Chtimes(stampFile, oldTime, oldTime); err != nil {
			t.Fatal(err)
		}

		maybeBackgroundUpgrade()

		if !state.bgCalled {
			t.Error("expected background install to be called")
		}
		if state.bgTag != "v0.21.0" {
			t.Errorf("bg tag = %q, want v0.21.0", state.bgTag)
		}
	})

	t.Run("check interval not elapsed", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()

		stampDir := stampDirPath(t)
		if err := os.MkdirAll(stampDir, 0o755); err != nil {
			t.Fatal(err)
		}
		stampFile := stampFilePath(t)
		if err := os.WriteFile(stampFile, nil, 0o644); err != nil {
			t.Fatal(err)
		}

		maybeBackgroundUpgrade()

		if state.bgCalled {
			t.Error("expected background install NOT to be called (recent stamp)")
		}
	})

	t.Run("not a go install installation", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()

		otherDir := t.TempDir()
		exePath := filepath.Join(otherDir, "litespec")
		if err := os.WriteFile(exePath, nil, 0o755); err != nil {
			t.Fatal(err)
		}
		oldExeFn := executableFn
		defer func() { executableFn = oldExeFn }()
		executableFn = func() (string, error) { return exePath, nil }

		maybeBackgroundUpgrade()

		if state.bgCalled {
			t.Error("expected background install NOT to be called (not go install)")
		}
	})
}

// --- Requirement: cache directory ---

func TestCacheDirectoryScenarios(t *testing.T) {
	t.Run("cache directory does not exist", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()

		stampDir := stampDirPath(t)
		if _, err := os.Stat(stampDir); !os.IsNotExist(err) {
			t.Fatalf("expected stamp dir to not exist yet, got: %v", err)
		}

		maybeBackgroundUpgrade()

		info, err := os.Stat(stampDir)
		if err != nil {
			t.Errorf("expected stamp dir to be created: %v", err)
		} else if !info.IsDir() {
			t.Error("expected stamp dir to be a directory")
		}
		if !state.bgCalled {
			t.Error("expected background install to be called")
		}
	})

	t.Run("timestamp file does not exist", func(t *testing.T) {
		state, cleanup := setupUpgradeTest(t, []string{"v0.20.2", "v0.21.0"}, "0.20.2")
		defer cleanup()

		stampDir := stampDirPath(t)
		if err := os.MkdirAll(stampDir, 0o755); err != nil {
			t.Fatal(err)
		}
		stampFile := stampFilePath(t)
		if _, err := os.Stat(stampFile); !os.IsNotExist(err) {
			t.Fatalf("expected stamp file to not exist, got: %v", err)
		}

		maybeBackgroundUpgrade()

		if !state.bgCalled {
			t.Error("expected background install to be called (no stamp file)")
		}
		if _, err := os.Stat(stampFile); err != nil {
			t.Errorf("expected stamp file to be created: %v", err)
		}
	})
}
