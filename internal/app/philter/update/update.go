package update

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	update "github.com/inconshreveable/go-update"
	philterversion "github.com/liamg/philter/internal/app/philter/version"
	version "github.com/mcuadros/go-version"
	log "github.com/sirupsen/logrus"
)

type release struct {
	TagName string  `json:"tag_name"`
	Assets  []asset `json:"assets"`
}

type fileset struct {
	Version      string
	BinaryURL    string
	BlacklistURL string
}

type asset struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

const repo = "liamg/philter"

func getLatestRelease() (*release, error) {

	client := http.Client{
		Timeout: time.Second * 10,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	release := release{}
	body, err := ioutil.ReadAll(res.Body)

	log.Infof("Response: %s", string(body))

	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

func isUpdateAvailable() (*fileset, bool, error) {
	release, err := getLatestRelease()
	if err != nil {
		return nil, false, err
	}

	if philterversion.Version == "" { // dev version, don't auto update
		return nil, false, nil
	}

	if !strings.Contains(philterversion.Version, "-") && version.Compare(release.TagName, philterversion.Version, "<=") {
		return nil, false, nil
	}

	pkg := fileset{
		Version: release.TagName,
	}

	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, fmt.Sprintf("philter-%s-%s", runtime.GOOS, runtime.GOARCH)) {
			pkg.BinaryURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", repo, asset.ID)
		} else if asset.Name == "blacklist.txt" {
			pkg.BlacklistURL = fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", repo, asset.ID)
		}
	}

	if pkg.BinaryURL == "" || pkg.BlacklistURL == "" {
		return nil, false, nil
	}

	return &pkg, true, nil
}

// Update binary from github releases
func Update() (bool, error) {

	var pkg *fileset
	var available bool

	var err error
	pkg, available, err = isUpdateAvailable()
	if err != nil {
		return false, fmt.Errorf("unable to check for updates: %s", err)
	}

	if !available {
		return false, nil
	}

	client := http.Client{
		Timeout: time.Second * 120,
	}

	req, err := http.NewRequest(http.MethodGet, pkg.BlacklistURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Accept

	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	f, err := os.Create("/var/lib/philter/blacklist.txt")
	if err != nil {
		return false, err
	}
	defer f.Close()

	if _, err := io.Copy(f, res.Body); err != nil {
		return false, err
	}

	req2, err := http.NewRequest(http.MethodGet, pkg.BinaryURL, nil)
	if err != nil {
		return false, err
	}
	req2.Header.Add("Accept", "application/octet-stream")
	res2, err := client.Do(req2)
	if err != nil {
		return false, err
	}
	defer res2.Body.Close()

	err = update.Apply(res2.Body, update.Options{TargetPath: "/usr/bin/philter"})
	return err == nil, err

}
