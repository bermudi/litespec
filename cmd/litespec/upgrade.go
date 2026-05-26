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

	"github.com/bermudi/litespec/internal"
)

func cmdUpgrade(args []string) error {
	fs := newFlagSet("upgrade", printUpgradeHelp)
	var asJSON, asMinimal bool
	fs.BoolVar(&asJSON, "json", false, "output as JSON")
	fs.BoolVar(&asMinimal, "minimal", false, "minimal output")

	ok, err := parseFlagSet(fs, args)
	if !ok {
		return err
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

	localVersion := version
	if localVersion == "dev" || localVersion == "" {
		localVersion = "0.0.0"
	}

	cmp, err := compareSemver(localVersion, latestTag)
	if err != nil {
		return fmt.Errorf("cannot determine current version (%q): %w", version, err)
	}

	if cmp >= 0 {
		if asJSON {
			if asMinimal {
				type upgradeMinimalJSON struct {
					Upgraded       bool   `json:"upgraded"`
					CurrentVersion string `json:"currentVersion"`
				}
				data, err := internal.MarshalJSON(upgradeMinimalJSON{Upgraded: false, CurrentVersion: localVersion})
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}
			type upgradeResultJSON struct {
				Upgraded       bool   `json:"upgraded"`
				CurrentVersion string `json:"currentVersion"`
				Message        string `json:"message"`
			}
			data, err := internal.MarshalJSON(upgradeResultJSON{Upgraded: false, CurrentVersion: localVersion, Message: "Already up to date"})
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		if asMinimal {
			fmt.Printf("up-to-date\tv%s\n", localVersion)
			return nil
		}
		fmt.Printf("Already up to date (v%s)\n", localVersion)
		return nil
	}

	cmd := exec.Command("go", "install", modulePath+"/cmd/litespec@"+latestTag)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOPROXY=https://proxy.golang.org,direct")
	if err := cmd.Run(); err != nil {
		return err
	}

	if asJSON {
		if asMinimal {
			type upgradeMinimalJSON struct {
				Upgraded        bool   `json:"upgraded"`
				PreviousVersion string `json:"previousVersion"`
				NewVersion      string `json:"newVersion"`
			}
			data, err := internal.MarshalJSON(upgradeMinimalJSON{Upgraded: true, PreviousVersion: localVersion, NewVersion: latestTag})
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}
		type upgradeResultJSON struct {
			Upgraded        bool   `json:"upgraded"`
			PreviousVersion string `json:"previousVersion"`
			NewVersion      string `json:"newVersion"`
			Hint            string `json:"hint"`
		}
		data, err := internal.MarshalJSON(upgradeResultJSON{
			Upgraded:        true,
			PreviousVersion: localVersion,
			NewVersion:      latestTag,
			Hint:            "Run 'litespec update' in your projects to refresh generated artifacts",
		})
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	if asMinimal {
		fmt.Printf("upgraded\tv%s\n", latestTag)
		return nil
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
