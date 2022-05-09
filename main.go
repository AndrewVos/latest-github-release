package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type Asset struct {
	Name               string
	ContentType        string `json:"content_type"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type Release struct {
	Assets []Asset
}

func findAsset(username string, repo string, format string, architecture string) (Asset, bool, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%v/%v/releases/latest", username, repo)
	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Asset{}, false, err
	}

	res, err := client.Do(req)
	if err != nil {
		return Asset{}, false, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return Asset{}, false, err
	}

	release := Release{}
	err = json.Unmarshal(body, &release)
	if err != nil {
		return Asset{}, false, err
	}

	for _, asset := range release.Assets {
		matchesContentType := format == "gzip" && asset.ContentType == "application/gzip"
		matchesArchitecture := architecture == "linux" && strings.Contains(strings.ToLower(asset.Name), "linux")
		if matchesContentType && matchesArchitecture {
			return asset, true, nil
		}
	}

	return Asset{}, false, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()
	r.GET("/:username/:repo/:format/:architecture", func(c *gin.Context) {
		username := c.Param("username")
		repo := c.Param("repo")
		format := c.Param("format")
		architecture := c.Param("architecture")

		asset, found, err := findAsset(username, repo, format, architecture)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %+v", err))
			return
		}

		if found {
			c.Redirect(http.StatusFound, asset.BrowserDownloadUrl)
		} else {
			c.String(http.StatusNotFound, "Couldn't find an asset")
		}
	})
	r.Run(":" + port)
}
