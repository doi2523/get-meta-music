package music

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// GetSongMetadata auto-detects platform (Spotify / Apple Music) and fetches metadata
func GetSongMetadata(rawURL string) (*SongMetadata, error) {
	if strings.Contains(rawURL, "spotify.com") {
		return GetSpotifyMetadata(rawURL)
	} else if strings.Contains(rawURL, "music.apple.com") || strings.Contains(rawURL, "apple.com") {
		return GetAppleMusicMetadata(rawURL)
	}
	return nil, fmt.Errorf("unsupported URL domain (must be spotify.com or music.apple.com)")
}

// GetSpotifyMetadata extracts song metadata for Spotify links
func GetSpotifyMetadata(rawURL string) (*SongMetadata, error) {
	if !strings.Contains(rawURL, "spotify.com") {
		return nil, fmt.Errorf("URL is not a valid Spotify link")
	}

	rawURL = cleanSpotifyURL(rawURL)
	log.Printf("🎧 [GetSpotifyMetadata] Starting extraction for: %s", rawURL)

	meta := &SongMetadata{
		SpotifyURL: rawURL,
		Source:     "Spotify",
	}

	// 1. Extract Track ID & Fetch via Spotify Embed Widget (Most reliable metadata)
	reTrackID := regexp.MustCompile(`(?:track|embed/track)/([a-zA-Z0-9]{22})`)
	matchTrack := reTrackID.FindStringSubmatch(rawURL)

	if len(matchTrack) > 1 {
		trackID := matchTrack[1]
		meta.SpotifyURL = fmt.Sprintf("https://open.spotify.com/track/%s", trackID)
		log.Printf("   ├─ [Step 1] Extracted Spotify Track ID: %s", trackID)

		embedURL := fmt.Sprintf("https://open.spotify.com/embed/track/%s", trackID)
		if embedData, err := fetchSpotifyEmbedData(embedURL); err == nil && embedData != nil {
			meta.SongTitle = embedData.Title
			meta.Artist = embedData.Artist
			meta.PreviewURL = embedData.PreviewURL
			meta.ImageURL = embedData.ImageURL
			log.Printf("   ├─ [Step 1 Success] Spotify Embed Data: Title='%s', Artist='%s'", meta.SongTitle, meta.Artist)
		} else {
			log.Printf("   ├─ [Step 1 Fallback] Spotify Embed Widget failed: %v", err)
		}
	}

	// 2. Fallback to Spotify oEmbed API if essential data still missing
	if meta.SongTitle == "" || meta.ImageURL == "" {
		log.Printf("   ├─ [Step 2] Trying Spotify oEmbed API...")
		if oembedData, err := fetchSpotifyOEmbed(rawURL); err == nil && oembedData != nil {
			if meta.SongTitle == "" {
				meta.SongTitle = oembedData.Title
			}
			if meta.ImageURL == "" {
				meta.ImageURL = oembedData.ThumbnailURL
			}
			log.Printf("   ├─ [Step 2 Success] oEmbed Data: Title='%s'", meta.SongTitle)
		} else {
			log.Printf("   ├─ [Step 2 Fallback] oEmbed API failed: %v", err)
		}
	}

	// 3. Fallback to Direct HTML Page Scraping
	if meta.SongTitle == "" || meta.ImageURL == "" {
		log.Printf("   ├─ [Step 3] Trying direct HTML scraping...")
		if htmlContent, err := fetchHTMLPage(rawURL); err == nil {
			if meta.SongTitle == "" {
				meta.SongTitle = getMetaProperty(htmlContent, "og:title")
				if meta.SongTitle == "" {
					meta.SongTitle = getMetaName(htmlContent, "twitter:title")
				}
			}
			if meta.Artist == "" {
				desc := getMetaProperty(htmlContent, "og:description")
				if desc != "" {
					meta.Artist = strings.Split(desc, " · ")[0]
				}
			}
			if meta.ImageURL == "" {
				meta.ImageURL = getMetaProperty(htmlContent, "og:image")
			}
			if meta.PreviewURL == "" {
				meta.PreviewURL = getMetaProperty(htmlContent, "og:audio")
			}
			log.Printf("   ├─ [Step 3 Result] HTML Scraping -> Title='%s', Artist='%s'", meta.SongTitle, meta.Artist)
		} else {
			log.Printf("   ├─ [Step 3 Fallback] HTML Scraping failed: %v", err)
		}
	}

	// 4. Try ISRC Lookup (Deezer / Lynkify)
	log.Printf("   ├─ [Step 4] Querying ISRC lookup for Spotify URL...")
	lookupURL := rawURL
	if meta.SpotifyURL != "" {
		lookupURL = meta.SpotifyURL
	}
	isrcData, err := GetISRCFromURL(lookupURL)
	if err == nil && isrcData != nil {
		meta.ISRC = isrcData.ISRC
		if isrcData.Source != "" {
			meta.Source = isrcData.Source
		}
		if meta.SongTitle == "" && isrcData.Title != "" {
			meta.SongTitle = isrcData.Title
		}
		if meta.Artist == "" && isrcData.Artist != "" {
			meta.Artist = isrcData.Artist
		}
		if meta.PreviewURL == "" && isrcData.PreviewURL != "" {
			meta.PreviewURL = isrcData.PreviewURL
		}
		log.Printf("   └─ [Step 4 Success] Found ISRC: %s (Source: %s)", meta.ISRC, meta.Source)
	} else {
		log.Printf("   └─ [Step 4 Warning] ISRC Lookup returned no results/error: %v", err)
	}

	// Validate essential fields
	if meta.SongTitle == "" || meta.ImageURL == "" {
		return nil, fmt.Errorf("essential metadata (song_title or image_url) missing")
	}

	meta.SongTitle = html.UnescapeString(meta.SongTitle)
	meta.Artist = html.UnescapeString(meta.Artist)

	return meta, nil
}

// GetAppleMusicMetadata extracts song metadata for Apple Music links
func GetAppleMusicMetadata(rawURL string) (*SongMetadata, error) {
	if !strings.Contains(rawURL, "music.apple.com") && !strings.Contains(rawURL, "apple.com") {
		return nil, fmt.Errorf("URL is not a valid Apple Music link")
	}

	log.Printf("🍎 [GetAppleMusicMetadata] Starting extraction for: %s", rawURL)

	meta := &SongMetadata{
		AppleMusicURL: rawURL,
		Source:        "Apple Music",
	}

	// 1. Extract Apple Track/Album ID & query iTunes Lookup API
	appleID := extractAppleID(rawURL)
	if appleID != "" {
		log.Printf("   ├─ [Step 1] Parsed Apple ID: %s. Querying iTunes Lookup API...", appleID)
		if itunesData, err := fetchITunesLookup(appleID); err == nil && itunesData != nil {
			meta.SongTitle = itunesData.TrackName
			meta.Artist = itunesData.ArtistName
			meta.PreviewURL = itunesData.PreviewUrl
			meta.ImageURL = strings.Replace(itunesData.ArtworkUrl100, "100x100bb", "600x600bb", 1)
			if itunesData.TrackViewUrl != "" {
				meta.AppleMusicURL = itunesData.TrackViewUrl
			}
			if itunesData.ISRC != "" {
				meta.ISRC = itunesData.ISRC
			}
			log.Printf("   ├─ [Step 1 Success] iTunes Lookup Data: Title='%s', Artist='%s', ISRC='%s'", meta.SongTitle, meta.Artist, meta.ISRC)
		} else {
			log.Printf("   ├─ [Step 1 Fallback] iTunes Lookup API failed: %v", err)
		}
	} else {
		log.Printf("   ├─ [Step 1 Warning] Could not parse Apple ID from URL")
	}

	// 2. Fallback to Direct HTML Page Scraping
	if meta.SongTitle == "" || meta.ImageURL == "" {
		log.Printf("   ├─ [Step 2] Trying HTML Scraping / JSON-LD...")
		if htmlContent, err := fetchHTMLPage(rawURL); err == nil {
			jsonLd := getJsonLd(htmlContent)

			if meta.SongTitle == "" {
				if jsonLd != nil && jsonLd["name"] != nil {
					meta.SongTitle = fmt.Sprintf("%v", jsonLd["name"])
				} else {
					meta.SongTitle = getMetaName(htmlContent, "apple:title")
					if meta.SongTitle == "" {
						meta.SongTitle = getMetaProperty(htmlContent, "og:title")
					}
				}
			}

			if meta.Artist == "" && jsonLd != nil {
				meta.Artist = extractArtistFromJSONLD(jsonLd)
			}

			if meta.PreviewURL == "" {
				if jsonLd != nil {
					meta.PreviewURL = extractPreviewFromJSONLD(jsonLd)
				}
				if meta.PreviewURL == "" {
					meta.PreviewURL = getMetaProperty(htmlContent, "music:song:preview_url:secure_url")
				}
			}

			if meta.ImageURL == "" {
				if jsonLd != nil && jsonLd["image"] != nil {
					meta.ImageURL = extractImageFromJSONLD(jsonLd["image"])
				}
				if meta.ImageURL == "" {
					meta.ImageURL = getMetaProperty(htmlContent, "og:image")
				}
			}
			log.Printf("   ├─ [Step 2 Result] HTML Scraping -> Title='%s', Artist='%s'", meta.SongTitle, meta.Artist)
		} else {
			log.Printf("   ├─ [Step 2 Fallback] HTML Scraping failed: %v", err)
		}
	}

	// 3. Try ISRC Lookup (Deezer / Lynkify)
	log.Printf("   ├─ [Step 3] Querying ISRC lookup for Apple Music URL...")
	isrcData, err := GetISRCFromURL(rawURL)
	if err == nil && isrcData != nil {
		if meta.ISRC == "" {
			meta.ISRC = isrcData.ISRC
		}
		if isrcData.Source != "" {
			meta.Source = isrcData.Source
		}
		if meta.SongTitle == "" && isrcData.Title != "" {
			meta.SongTitle = isrcData.Title
		}
		if meta.Artist == "" && isrcData.Artist != "" {
			meta.Artist = isrcData.Artist
		}
		if meta.PreviewURL == "" && isrcData.PreviewURL != "" {
			meta.PreviewURL = isrcData.PreviewURL
		}
		log.Printf("   └─ [Step 3 Success] Found ISRC: %s (Source: %s)", meta.ISRC, meta.Source)
	} else {
		log.Printf("   └─ [Step 3 Warning] ISRC Lookup returned no results/error: %v", err)
	}

	if meta.SongTitle == "" || meta.ImageURL == "" {
		return nil, fmt.Errorf("essential metadata (song_title or image_url) missing")
	}

	meta.SongTitle = html.UnescapeString(meta.SongTitle)
	meta.Artist = html.UnescapeString(meta.Artist)

	return meta, nil
}

/* ---------------- Helper Data Fetchers ---------------- */

type spotifyEmbedData struct {
	Title      string
	Artist     string
	PreviewURL string
	ImageURL   string
}

func fetchSpotifyEmbedData(embedURL string) (*spotifyEmbedData, error) {
	req, err := http.NewRequest("GET", embedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reNextData := regexp.MustCompile(`(?s)<script id="__NEXT_DATA__"[^>]*>([\s\S]*?)</script>`)
	match := reNextData.FindStringSubmatch(string(bodyBytes))
	if len(match) <= 1 {
		return nil, fmt.Errorf("__NEXT_DATA__ script not found in spotify embed page")
	}

	var root struct {
		Props struct {
			PageProps struct {
				State struct {
					Data struct {
						Entity struct {
							Name    string `json:"name"`
							Title   string `json:"title"`
							Artists []struct {
								Name string `json:"name"`
							} `json:"artists"`
							AudioPreview struct {
								URL string `json:"url"`
							} `json:"audioPreview"`
							VisualIdentity struct {
								Image []struct {
									URL string `json:"url"`
								} `json:"image"`
							} `json:"visualIdentity"`
						} `json:"entity"`
					} `json:"data"`
				} `json:"state"`
			} `json:"pageProps"`
		} `json:"props"`
	}

	if err := json.Unmarshal([]byte(match[1]), &root); err != nil {
		return nil, err
	}

	entity := root.Props.PageProps.State.Data.Entity

	songTitle := entity.Name
	if songTitle == "" {
		songTitle = entity.Title
	}

	var artistNames []string
	for _, a := range entity.Artists {
		if a.Name != "" {
			artistNames = append(artistNames, a.Name)
		}
	}
	artistStr := strings.Join(artistNames, ", ")

	imgURL := ""
	if len(entity.VisualIdentity.Image) > 0 {
		imgURL = entity.VisualIdentity.Image[0].URL
	}

	return &spotifyEmbedData{
		Title:      songTitle,
		Artist:     artistStr,
		PreviewURL: entity.AudioPreview.URL,
		ImageURL:   imgURL,
	}, nil
}

type spotifyOEmbedResponse struct {
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnail_url"`
}

func fetchSpotifyOEmbed(rawURL string) (*spotifyOEmbedResponse, error) {
	oembedURL := "https://open.spotify.com/oembed?url=" + url.QueryEscape(rawURL)
	req, err := http.NewRequest("GET", oembedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("oembed returned status %d", resp.StatusCode)
	}

	var res spotifyOEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

type iTunesTrack struct {
	TrackName     string `json:"trackName"`
	ArtistName    string `json:"artistName"`
	PreviewUrl    string `json:"previewUrl"`
	ArtworkUrl100 string `json:"artworkUrl100"`
	TrackViewUrl  string `json:"trackViewUrl"`
	ISRC          string `json:"isrc"`
}

func fetchITunesLookup(appleID string) (*iTunesTrack, error) {
	apiURL := fmt.Sprintf("https://itunes.apple.com/lookup?id=%s&country=VN", appleID)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ResultCount int           `json:"resultCount"`
		Results     []iTunesTrack `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ResultCount > 0 && len(result.Results) > 0 {
		return &result.Results[0], nil
	}

	return nil, fmt.Errorf("no iTunes lookup results")
}

func fetchHTMLPage(rawURL string) (string, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

// extractAppleID parses an Apple Music URL and returns the track/album ID.
// This is the single source of truth — used by both metadata.go and isrc.go.
func extractAppleID(rawURL string) string {
	reI := regexp.MustCompile(`[?&]i=(\d+)`)
	if match := reI.FindStringSubmatch(rawURL); len(match) > 1 {
		return match[1]
	}

	reSong := regexp.MustCompile(`/song/.*/(\d+)`)
	if match := reSong.FindStringSubmatch(rawURL); len(match) > 1 {
		return match[1]
	}

	reAlbum := regexp.MustCompile(`/album/.*/(\d+)`)
	if match := reAlbum.FindStringSubmatch(rawURL); len(match) > 1 {
		return match[1]
	}

	return ""
}

/* ---------------- HTML & Meta Helpers ---------------- */

func getMetaProperty(htmlStr, property string) string {
	quoted := regexp.QuoteMeta(property)
	re1 := regexp.MustCompile(`(?i)<meta[^>]+property=["']` + quoted + `["'][^>]+content=["']([^"']*)["']`)
	if match := re1.FindStringSubmatch(htmlStr); len(match) > 1 {
		return match[1]
	}
	re2 := regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*)["'][^>]+property=["']` + quoted + `["']`)
	if match := re2.FindStringSubmatch(htmlStr); len(match) > 1 {
		return match[1]
	}
	return ""
}

func getMetaName(htmlStr, name string) string {
	quoted := regexp.QuoteMeta(name)
	re1 := regexp.MustCompile(`(?i)<meta[^>]+name=["']` + quoted + `["'][^>]+content=["']([^"']*)["']`)
	if match := re1.FindStringSubmatch(htmlStr); len(match) > 1 {
		return match[1]
	}
	re2 := regexp.MustCompile(`(?i)<meta[^>]+content=["']([^"']*)["'][^>]+name=["']` + quoted + `["']`)
	if match := re2.FindStringSubmatch(htmlStr); len(match) > 1 {
		return match[1]
	}
	return ""
}

func getJsonLd(htmlStr string) map[string]interface{} {
	re := regexp.MustCompile(`(?i)<script[^>]+type=["']application/ld\+json["'][^>]*>([\s\S]*?)</script>`)
	matches := re.FindAllStringSubmatch(htmlStr, -1)
	for _, match := range matches {
		if len(match) > 1 {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &result); err == nil {
				return result
			}
			var arrResult []map[string]interface{}
			if err := json.Unmarshal([]byte(match[1]), &arrResult); err == nil && len(arrResult) > 0 {
				return arrResult[0]
			}
		}
	}
	return nil
}

func extractArtistFromJSONLD(jsonLd map[string]interface{}) string {
	if audio, ok := jsonLd["audio"].(map[string]interface{}); ok {
		if byArtist, ok := audio["byArtist"].([]interface{}); ok && len(byArtist) > 0 {
			if artistMap, ok := byArtist[0].(map[string]interface{}); ok {
				if name, ok := artistMap["name"].(string); ok {
					return name
				}
			}
		}
	}
	if byArtist, ok := jsonLd["byArtist"].([]interface{}); ok && len(byArtist) > 0 {
		if artistMap, ok := byArtist[0].(map[string]interface{}); ok {
			if name, ok := artistMap["name"].(string); ok {
				return name
			}
		}
	}
	return ""
}

func extractPreviewFromJSONLD(jsonLd map[string]interface{}) string {
	if audio, ok := jsonLd["audio"].(map[string]interface{}); ok {
		if innerAudio, ok := audio["audio"].(map[string]interface{}); ok {
			if contentURL, ok := innerAudio["contentUrl"].(string); ok {
				return contentURL
			}
		}
	}
	return ""
}

func extractImageFromJSONLD(imgObj interface{}) string {
	switch v := imgObj.(type) {
	case string:
		return v
	case []interface{}:
		if len(v) > 0 {
			if str, ok := v[0].(string); ok {
				return str
			}
		}
	case map[string]interface{}:
		if urlStr, ok := v["url"].(string); ok {
			return urlStr
		}
	}
	return ""
}
