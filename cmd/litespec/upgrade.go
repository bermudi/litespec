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

	"github.com/bermudi/litespec/v2/internal"
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

	localVersion := version
	if localVersion == "dev" || localVersion == "" {
		localVersion = "0.0.0"
	}

	latestTag, err := fetchLatestVersionFor(localVersion)
	if err != nil {
		return err
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

type semver struct {
	major      int
	minor      int
	patch      int
	prerelease string
}

type githubTag struct {
	Name string `json:"name"`
}

func fetchLatestVersion() (string, error) {
	tags, err := fetchTagNames("https://api.github.com/repos/bermudi/litespec/tags")
	if err != nil {
		return "", err
	}
	return selectLatestStable(tags)
}

func fetchLatestVersionFor(localVersion string) (string, error) {
	tags, err := fetchTagNames("https://api.github.com/repos/bermudi/litespec/tags")
	if err != nil {
		return "", err
	}
	return selectLatestForChannel(tags, localVersion)
}

func fetchTagNames(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to check for updates: HTTP %d", resp.StatusCode)
	}

	var tags []githubTag
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("failed to parse tag info: %w", err)
	}
	names := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.Name != "" {
			names = append(names, t.Name)
		}
	}
	return names, nil
}

func isPrerelease(tag string) bool {
	v := strings.TrimPrefix(tag, "v")
	return strings.Contains(v, "-")
}

func selectLatestStable(tags []string) (string, error) {
	best := ""
	var bestV semver
	for _, t := range tags {
		if isPrerelease(t) {
			continue
		}
		v, err := parseSemver(t)
		if err != nil {
			continue
		}
		if best == "" || compareSemverVersions(v, bestV) > 0 {
			best = t
			bestV = v
		}
	}
	if best == "" {
		return "", fmt.Errorf("no stable release tag found")
	}
	return best, nil
}

func selectLatestOverall(tags []string) (string, error) {
	best := ""
	var bestV semver
	for _, t := range tags {
		v, err := parseSemver(t)
		if err != nil {
			continue
		}
		if best == "" || compareSemverVersions(v, bestV) > 0 {
			best = t
			bestV = v
		}
	}
	if best == "" {
		return "", fmt.Errorf("no release tag found")
	}
	return best, nil
}

func selectLatestForChannel(tags []string, localVersion string) (string, error) {
	if isPrerelease(localVersion) {
		return selectLatestOverall(tags)
	}
	return selectLatestStable(tags)
}

func parseSemver(tag string) (semver, error) {
	raw := strings.TrimPrefix(tag, "v")
	parts := strings.SplitN(raw, ".", 3)
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("invalid semver: %q", tag)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, fmt.Errorf("invalid semver major: %q", parts[0])
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, fmt.Errorf("invalid semver minor: %q", parts[1])
	}
	patchStr := parts[2]
	prerelease := ""
	if idx := strings.IndexByte(patchStr, '-'); idx >= 0 {
		prerelease = patchStr[idx+1:]
		patchStr = patchStr[:idx]
	}
	if idx := strings.IndexByte(patchStr, '+'); idx >= 0 {
		patchStr = patchStr[:idx]
	}
	patch, err := strconv.Atoi(patchStr)
	if err != nil {
		return semver{}, fmt.Errorf("invalid semver patch: %q", parts[2])
	}
	return semver{major, minor, patch, prerelease}, nil
}

func compareSemver(local, remote string) (int, error) {
	ls, err := parseSemver(local)
	if err != nil {
		return 0, err
	}
	rs, err := parseSemver(remote)
	if err != nil {
		return 0, err
	}
	return compareSemverVersions(ls, rs), nil
}

func compareSemverVersions(a, b semver) int {
	switch {
	case a.major != b.major:
		if a.major > b.major {
			return 1
		}
		return -1
	case a.minor != b.minor:
		if a.minor > b.minor {
			return 1
		}
		return -1
	case a.patch != b.patch:
		if a.patch > b.patch {
			return 1
		}
		return -1
	}
	switch {
	case a.prerelease == "" && b.prerelease == "":
		return 0
	case a.prerelease == "":
		return 1
	case b.prerelease == "":
		return -1
	default:
		return comparePrerelease(a.prerelease, b.prerelease)
	}
}

func comparePrerelease(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) < n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		ai, errA := strconv.Atoi(as[i])
		bi, errB := strconv.Atoi(bs[i])
		if errA == nil && errB == nil {
			if ai != bi {
				if ai > bi {
					return 1
				}
				return -1
			}
			continue
		}
		if errA == nil && errB != nil {
			return -1
		}
		if errA != nil && errB == nil {
			return 1
		}
		if as[i] != bs[i] {
			if as[i] > bs[i] {
				return 1
			}
			return -1
		}
	}
	if len(as) != len(bs) {
		if len(as) > len(bs) {
			return 1
		}
		return -1
	}
	return 0
}
