package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// Github repo url
const (
	updateURL = "https://api.github.com/repos/devusSs/twitchspeak/releases/latest"
)

// Github release struct
type githubRelease struct {
	URL       string `json:"url"`
	AssetsURL string `json:"assets_url"`
	UploadURL string `json:"upload_url"`
	HTMLURL   string `json:"html_url"`
	ID        int    `json:"id"`
	Author    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"author"`
	NodeID          string    `json:"node_id"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	Assets          []struct {
		URL      string `json:"url"`
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		Label    string `json:"label"`
		Uploader struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"uploader"`
		ContentType        string    `json:"content_type"`
		State              string    `json:"state"`
		Size               int       `json:"size"`
		DownloadCount      int       `json:"download_count"`
		CreatedAt          time.Time `json:"created_at"`
		UpdatedAt          time.Time `json:"updated_at"`
		BrowserDownloadURL string    `json:"browser_download_url"`
	} `json:"assets"`
	TarballURL string `json:"tarball_url"`
	ZipballURL string `json:"zipball_url"`
	Body       string `json:"body"`
}

// Checks for updates and applies them if possible
func CheckForUpdatesAndApply(buildVersion string) error {
	updateURL, newVersion, changelog, err := findLatestReleaseURL()
	if err != nil {
		return fmt.Errorf("failed to find latest release: %w", err)
	}
	newVersionAvailable, err := newerVersionAvailable(newVersion, buildVersion)
	if err != nil {
		return fmt.Errorf("failed to compare versions: %w", err)
	}
	if newVersionAvailable {
		if err := doUpdate(updateURL); err != nil {
			return fmt.Errorf("failed to update: %w", err)
		}
		fmt.Printf("Update changelog (%s): %s\n", newVersion, changelog)
		fmt.Println("Update succeeded, please restart the app")
		os.Exit(0)
	}
	return nil
}

// Periodically checks for updates
//
// Blocking until context is canceled
func PeriodicUpdateCheck(
	ctx context.Context,
	buildVersion string,
	available chan bool,
	wg *sync.WaitGroup,
) {
	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			wg.Done()
			return
		case <-ticker.C:
			_, newVersion, _, err := findLatestReleaseURL()
			if err != nil {
				fmt.Printf("Failed to find latest release: %v\n", err)
				continue
			}
			newVersionAvailable, err := newerVersionAvailable(newVersion, buildVersion)
			if err != nil {
				fmt.Printf("Failed to compare versions: %v\n", err)
				continue
			}
			if newVersionAvailable {
				available <- true
			}
		}
	}
}

// Queries the latest release from Github repo.
func findLatestReleaseURL() (string, string, string, error) {
	resp, err := http.Get(updateURL)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}
	var release githubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", "", err
	}
	buildArch := runtime.GOARCH
	if buildArch == "amd64" {
		buildArch = "x86_64"
	}
	if buildArch == "386" {
		buildArch = "i386"
	}
	buildOS := runtime.GOOS
	for _, asset := range release.Assets {
		releaseName := strings.ToLower(asset.Name)
		if strings.Contains(releaseName, buildArch) && strings.Contains(releaseName, buildOS) {
			changeSplit := strings.Split(
				strings.ReplaceAll(strings.TrimSpace(release.Body), "## Changelog", ""),
				"\n",
			)
			for i, line := range changeSplit {
				changeSplit[i] = strings.ReplaceAll(fmt.Sprintf("\t\t\t%s", line), "*", "-")
			}
			changelog := strings.Join(changeSplit, "\n")
			return asset.BrowserDownloadURL, release.TagName, changelog, nil
		}
	}
	return "", "", "", errors.New("no matching release found")
}

// Compare current version with latest version
func newerVersionAvailable(newVersion string, buildVersion string) (bool, error) {
	vOld, err := semver.NewVersion(buildVersion)
	if err != nil {
		return false, err
	}
	vNew, err := semver.NewVersion(newVersion)
	if err != nil {
		return false, err
	}
	return vOld.LessThan(vNew), nil
}

// Perform the actual patch.
func doUpdate(url string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := selfupdate.UpdateTo(url, exe); err != nil {
		return err
	}
	return nil
}
