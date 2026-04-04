package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
)

func cmdUpgrade(args []string) error {
	if hasHelpFlag(args) {
		printUpgradeHelp()
		return nil
	}

	if !isGoInstall() {
		return fmt.Errorf("auto-upgrade only supports installations via 'go install'")
	}

	modulePath, err := getModulePath()
	if err != nil {
		return err
	}

	latestTag, err := fetchLatestVersion()
	if err != nil {
		return err
	}

	cmp, err := compareSemver(version, latestTag)
	if err != nil {
		return err
	}
	if cmp >= 0 {
		fmt.Printf("Already up to date (v%s)\n", version)
		return nil
	}

	cmd := exec.Command("go", "install", modulePath+"/cmd/litespec@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Printf("\nUpgraded to %s\n", latestTag)
	fmt.Println("Run 'litespec update' in your projects to refresh generated artifacts")
	return nil
}

func isGoInstall() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}

	if gobin := os.Getenv("GOBIN"); gobin != "" {
		gobin, err = filepath.EvalSymlinks(gobin)
		if err != nil {
			return false
		}
		if strings.HasPrefix(exe, gobin+string(os.PathSeparator)) {
			return true
		}
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		gopath = filepath.Join(home, "go")
	}
	gopath, err = filepath.EvalSymlinks(gopath)
	if err != nil {
		return false
	}
	gobinDefault := filepath.Join(gopath, "bin")
	return strings.HasPrefix(exe, gobinDefault+string(os.PathSeparator))
}

func getModulePath() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Path == "" {
		return "", fmt.Errorf("could not determine module path from build info")
	}
	return info.Main.Path, nil
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func fetchLatestVersion() (string, error) {
	return fetchLatestVersionFromURL("https://api.github.com/repos/bermudi/litespec/releases/latest")
}

func fetchLatestVersionFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}
	if release.TagName == "" {
		return "", fmt.Errorf("release tag not found in response")
	}
	return release.TagName, nil
}

func parseSemver(tag string) (int, int, int, error) {
	tag = strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(tag, ".", 3)
	if len(parts) != 3 {
		return 0, 0, 0, fmt.Errorf("invalid semver: %q", tag)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid semver major: %q", parts[0])
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid semver minor: %q", parts[1])
	}
	patchStr := parts[2]
	if idx := strings.IndexByte(patchStr, '-'); idx >= 0 {
		patchStr = patchStr[:idx]
	}
	if idx := strings.IndexByte(patchStr, '+'); idx >= 0 {
		patchStr = patchStr[:idx]
	}
	if idx := strings.IndexByte(patchStr, '+'); idx >= 0 {
		patchStr = patchStr[:idx]
	}
	patch, err := strconv.Atoi(patchStr)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid semver patch: %q", parts[2])
	}
	return major, minor, patch, nil
}

func compareSemver(local, remote string) (int, error) {
	lm, ln, lp, err := parseSemver(local)
	if err != nil {
		return 0, err
	}
	rm, rn, rp, err := parseSemver(remote)
	if err != nil {
		return 0, err
	}
	switch {
	case lm > rm:
		return 1, nil
	case lm < rm:
		return -1, nil
	case ln > rn:
		return 1, nil
	case ln < rn:
		return -1, nil
	case lp > rp:
		return 1, nil
	case lp < rp:
		return -1, nil
	default:
		return 0, nil
	}
}
