package music

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

type LynkifyActions struct {
	GetToken string
	Search   string
	GetISRC  string
}

var (
	actionsMutex       sync.RWMutex
	cachedActions      *LynkifyActions
	cachedSpotifyToken string
	tokenMutex         sync.RWMutex
	httpClient         = &http.Client{Timeout: 15 * time.Second}
)

func clearLynkifyCache() {
	actionsMutex.Lock()
	cachedActions = nil
	actionsMutex.Unlock()

	tokenMutex.Lock()
	cachedSpotifyToken = ""
	tokenMutex.Unlock()
}

func cleanSpotifyURL(rawURL string) string {
	re := regexp.MustCompile(`open\.spotify\.com/intl-[^/]+/`)
	return re.ReplaceAllString(rawURL, "open.spotify.com/")
}

// GetISRCFromURL attempts to retrieve ISRC data using song.link + Deezer, with fallbacks to Lynkify & Apple Music.
func GetISRCFromURL(rawURL string) (*ISRCData, error) {
	if strings.Contains(rawURL, "spotify.com") {
		rawURL = cleanSpotifyURL(rawURL)
	}

	log.Printf("🎵 Fetching ISRC for: %s", rawURL)

	// METHOD 1: song.link + Deezer
	isrcData, err := getISRCFromSongLink(rawURL)
	if err == nil && isrcData != nil && isrcData.ISRC != "" {
		log.Printf("✅ Found ISRC from Deezer: %s", isrcData.ISRC)
		return isrcData, nil
	}
	log.Println("⚠️ song.link + Deezer failed, proceeding to fallback search...")

	// METHOD 2: FALLBACK SEARCH
	if strings.Contains(rawURL, "spotify.com") {
		spotifyRes, err := searchBySpotify(rawURL, false)
		if err != nil || (spotifyRes != nil && spotifyRes.ISRC == "") {
			log.Printf("⚠️ Lynkify search failed: %v. Clearing cache and retrying...", err)
			clearLynkifyCache()
			spotifyRes, err = searchBySpotify(rawURL, false)
		}
		if err == nil && spotifyRes != nil && spotifyRes.ISRC != "" {
			return spotifyRes, nil
		}
	}

	if strings.Contains(rawURL, "apple.com") {
		appleRes, err := searchByAppleURL(rawURL)
		if err == nil && appleRes != nil && appleRes.ISRC != "" {
			return appleRes, nil
		}
	}

	return nil, fmt.Errorf("could not retrieve ISRC for url: %s", rawURL)
}

// getISRCFromSongLink queries song.link API and Deezer API
func getISRCFromSongLink(rawURL string) (*ISRCData, error) {
	songLinkURL := "https://api.song.link/v1-alpha.1/links?url=" + url.QueryEscape(rawURL)
	req, err := http.NewRequest("GET", songLinkURL, nil)
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
		return nil, fmt.Errorf("song.link returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var songLinkResult struct {
		EntitiesByUniqueID map[string]struct {
			ID string `json:"id"`
		} `json:"entitiesByUniqueId"`
	}

	if err := json.Unmarshal(body, &songLinkResult); err != nil {
		return nil, err
	}

	var deezerID string
	for key := range songLinkResult.EntitiesByUniqueID {
		if strings.HasPrefix(key, "DEEZER_SONG::") {
			parts := strings.Split(key, "::")
			if len(parts) >= 2 {
				deezerID = parts[1]
				break
			}
		}
	}

	if deezerID == "" {
		return nil, fmt.Errorf("deezer song ID not found in song.link response")
	}

	log.Printf("🆔 Found Deezer ID: %s", deezerID)

	deezerURL := fmt.Sprintf("https://api.deezer.com/track/%s", deezerID)
	dReq, err := http.NewRequest("GET", deezerURL, nil)
	if err != nil {
		return nil, err
	}
	dReq.Header.Set("User-Agent", UserAgent)

	dResp, err := httpClient.Do(dReq)
	if err != nil {
		return nil, err
	}
	defer dResp.Body.Close()

	if dResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("deezer returned status %d", dResp.StatusCode)
	}

	dBody, err := io.ReadAll(dResp.Body)
	if err != nil {
		return nil, err
	}

	var deezerResult struct {
		ISRC    string `json:"isrc"`
		Title   string `json:"title"`
		Preview string `json:"preview"`
		Artist  struct {
			Name string `json:"name"`
		} `json:"artist"`
	}

	if err := json.Unmarshal(dBody, &deezerResult); err != nil {
		return nil, err
	}

	if deezerResult.ISRC != "" {
		return &ISRCData{
			ISRC:       deezerResult.ISRC,
			Title:      deezerResult.Title,
			Artist:     deezerResult.Artist.Name,
			PreviewURL: deezerResult.Preview,
			Source:     "Deezer",
		}, nil
	}

	return nil, fmt.Errorf("isrc empty from deezer")
}

// discoverActions finds next-action hashes from lynkify.in JS bundle
func discoverActions() (*LynkifyActions, error) {
	actionsMutex.RLock()
	if cachedActions != nil {
		defer actionsMutex.RUnlock()
		return cachedActions, nil
	}
	actionsMutex.RUnlock()

	actionsMutex.Lock()
	defer actionsMutex.Unlock()

	if cachedActions != nil {
		return cachedActions, nil
	}

	req, err := http.NewRequest("GET", "https://lynkify.in/isrc-search", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	htmlBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reJS := regexp.MustCompile(`/_next/static/chunks/app/\(landing\)/isrc-search/page-[a-f0-9]+\.js`)
	jsMatches := reJS.FindString(string(htmlBody))
	if jsMatches == "" {
		return nil, fmt.Errorf("no lynkify js chunk found")
	}

	jsURL := "https://lynkify.in" + jsMatches
	jsReq, err := http.NewRequest("GET", jsURL, nil)
	if err != nil {
		return nil, err
	}
	jsReq.Header.Set("User-Agent", UserAgent)

	jsResp, err := httpClient.Do(jsReq)
	if err != nil {
		return nil, err
	}
	defer jsResp.Body.Close()

	jsContent, err := io.ReadAll(jsResp.Body)
	if err != nil {
		return nil, err
	}

	reHash := regexp.MustCompile(`"([a-f0-9]{40})"`)
	hashMatches := reHash.FindAllStringSubmatch(string(jsContent), -1)

	var hashes []string
	for _, m := range hashMatches {
		if len(m) > 1 {
			hashes = append(hashes, m[1])
		}
	}

	if len(hashes) >= 6 {
		cachedActions = &LynkifyActions{
			GetToken: hashes[1],
			Search:   hashes[2],
			GetISRC:  hashes[5],
		}
		log.Printf("✅ ACTIONS DISCOVERED: %+v", cachedActions)
		return cachedActions, nil
	}

	return nil, fmt.Errorf("failed to extract sufficient hashes from js (found %d)", len(hashes))
}

// getSpotifyToken obtains a temporary Spotify search token from lynkify.in
func getSpotifyToken() (string, error) {
	tokenMutex.RLock()
	if cachedSpotifyToken != "" {
		defer tokenMutex.RUnlock()
		return cachedSpotifyToken, nil
	}
	tokenMutex.RUnlock()

	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	if cachedSpotifyToken != "" {
		return cachedSpotifyToken, nil
	}

	actions, err := discoverActions()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://lynkify.in/isrc-search", bytes.NewBufferString("[]"))
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("next-action", actions.GetToken)
	req.Header.Set("content-type", "text/plain;charset=UTF-8")
	req.Header.Set("accept", "text/x-component")
	req.Header.Set("origin", "https://lynkify.in")
	req.Header.Set("referer", "https://lynkify.in/isrc-search")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	reToken := regexp.MustCompile(`1:"([^"]+)"`)
	match := reToken.FindStringSubmatch(string(respBody))
	if len(match) > 1 {
		cachedSpotifyToken = match[1]
		log.Printf("✅ TOKEN OBTAINED: %s", cachedSpotifyToken)
		return cachedSpotifyToken, nil
	}

	return "", fmt.Errorf("token not found in response: %s", string(respBody))
}

// searchBySpotify fetches track ISRC from track URL or track ID via Lynkify
func searchBySpotify(inputVal string, isIDOnly bool) (*ISRCData, error) {
	actions, err := discoverActions()
	if err != nil {
		return nil, err
	}

	token, err := getSpotifyToken()
	if err != nil {
		return nil, err
	}

	var trackID string
	if isIDOnly {
		trackID = inputVal
	} else {
		parts := strings.Split(inputVal, "/")
		last := parts[len(parts)-1]
		trackID = strings.Split(last, "?")[0]
	}

	if trackID == "" {
		return nil, fmt.Errorf("invalid track ID")
	}

	payload, err := json.Marshal([]string{token, trackID})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://lynkify.in/isrc-search", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("accept", "text/x-component")
	req.Header.Set("next-action", actions.GetISRC)
	req.Header.Set("content-type", "text/plain;charset=UTF-8")
	req.Header.Set("origin", "https://lynkify.in")
	req.Header.Set("referer", "https://lynkify.in/isrc-search")
	req.Header.Set("next-router-state-tree", `%5B%22%22%2C%7B%22children%22%3A%5B%22(landing)%22%2C%7B%22children%22%3A%5B%22isrc-search%22%2C%7B%22children%22%3A%5B%22__PAGE__%22%2C%7B%7D%2C%22%2Fisrc-search%22%2C%22refresh%22%5D%7D%5D%7D%5D%7D%2Cnull%2Cnull%2Ctrue%5D`)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		ISRC       string `json:"isrc"`
		Name       string `json:"name"`
		Artist     string `json:"artist"`
		CoverImage string `json:"coverImage"`
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		if idx := strings.Index(line, ":{"); idx != -1 {
			jsonStr := line[idx+1:]
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil && result.ISRC != "" {
				break
			}
		}
	}

	if result.ISRC == "" {
		return nil, fmt.Errorf("isrc is empty in spotify fallback result")
	}

	return &ISRCData{
		ISRC:   result.ISRC,
		Title:  result.Name,
		Artist: result.Artist,
		Source: "Spotify Fallback",
	}, nil
}

// searchSpotifyByName searches Spotify by track and artist name via Lynkify
func searchSpotifyByName(trackName, artistName string) (*ISRCData, error) {
	actions, err := discoverActions()
	if err != nil {
		return nil, err
	}

	token, err := getSpotifyToken()
	if err != nil {
		return nil, err
	}

	reFeat := regexp.MustCompile(`(?i)\(feat\..*?\)`)
	cleanName := strings.TrimSpace(reFeat.ReplaceAllString(trackName, ""))
	query := fmt.Sprintf("%s %s", cleanName, artistName)

	payload, err := json.Marshal([]string{token, query})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://lynkify.in/isrc-search", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("accept", "text/x-component")
	req.Header.Set("next-action", actions.Search)
	req.Header.Set("content-type", "text/plain;charset=UTF-8")
	req.Header.Set("origin", "https://lynkify.in")
	req.Header.Set("referer", "https://lynkify.in/isrc-search")
	req.Header.Set("next-router-state-tree", `%5B%22%22%2C%7B%22children%22%3A%5B%22(landing)%22%2C%7B%22children%22%3A%5B%22isrc-search%22%2C%7B%22children%22%3A%5B%22__PAGE__%22%2C%7B%7D%2C%22%2Fisrc-search%22%2C%22refresh%22%5D%7D%5D%7D%5D%7D%2Cnull%2Cnull%2Ctrue%5D`)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	reIDs := regexp.MustCompile(`(?s)"id":"([a-zA-Z0-9]{22})"`)
	idMatches := reIDs.FindAllStringSubmatch(string(body), -1)

	if len(idMatches) == 0 {
		return nil, fmt.Errorf("no spotify track IDs found for query: %s", query)
	}

	firstID := idMatches[0][1]
	return searchBySpotify(firstID, true)
}

// searchByAppleURL queries iTunes API by Apple track/album ID
func searchByAppleURL(rawURL string) (*ISRCData, error) {
	var appleID string

	// Try matching ?i=
	reI := regexp.MustCompile(`[?&]i=(\d+)`)
	if match := reI.FindStringSubmatch(rawURL); len(match) > 1 {
		appleID = match[1]
	}

	// Try matching /song/.../{id}
	if appleID == "" {
		reSong := regexp.MustCompile(`/song/.*/(\d+)`)
		if match := reSong.FindStringSubmatch(rawURL); len(match) > 1 {
			appleID = match[1]
		}
	}

	// Try matching /album/.../{id}
	if appleID == "" {
		reAlbum := regexp.MustCompile(`/album/.*/(\d+)`)
		if match := reAlbum.FindStringSubmatch(rawURL); len(match) > 1 {
			appleID = match[1]
		}
	}

	if appleID == "" {
		return nil, fmt.Errorf("could not parse Apple ID from URL")
	}

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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var itunesResult struct {
		ResultCount int `json:"resultCount"`
		Results     []struct {
			ISRC       string `json:"isrc"`
			TrackName  string `json:"trackName"`
			ArtistName string `json:"artistName"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &itunesResult); err != nil {
		return nil, err
	}

	if itunesResult.ResultCount == 0 || len(itunesResult.Results) == 0 {
		return nil, fmt.Errorf("no iTunes results found for Apple ID %s", appleID)
	}

	track := itunesResult.Results[0]
	if track.ISRC != "" {
		return &ISRCData{
			ISRC:   track.ISRC,
			Title:  track.TrackName,
			Artist: track.ArtistName,
			Source: "Apple Music",
		}, nil
	}

	log.Printf("⚠️ Apple hidden ISRC for %s. Searching Spotify...", track.TrackName)
	spotifyRes, err := searchSpotifyByName(track.TrackName, track.ArtistName)
	if err != nil || (spotifyRes != nil && spotifyRes.ISRC == "") {
		log.Printf("⚠️ Lynkify search failed: %v. Clearing cache and retrying...", err)
		clearLynkifyCache()
		spotifyRes, err = searchSpotifyByName(track.TrackName, track.ArtistName)
	}
	return spotifyRes, err
}
