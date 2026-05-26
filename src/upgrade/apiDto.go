package upgrade

import "time"

type apiResponse struct {
	TagName     string     `json:"tag_name"`
	PublishedAt time.Time  `json:"published_at"`
	Assets      []apiAsset `json:"assets"`
}

type apiAsset struct {
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}
