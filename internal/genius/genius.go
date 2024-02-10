package genius

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/maps"
)

type SearchResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Hits []struct {
			Result struct {
				ArtistNames     string                 `json:"artist_names"`
				PrimaryArtist   SearchResponseArtist   `json:"primary_artist"`
				FeaturedArtists []SearchResponseArtist `json:"featured_artists"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
}

type SearchResponseArtist struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type BulkSongsResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Songs    []BulkSong `json:"songs"`
		NextPage *int       `json:"next_page"`
	} `json:"response"`
}

type BulkSong struct {
	ArtistNames              string  `json:"artist_names"`
	AppleMusicPlayerUrl      *string `json:"apple_music_player_url,omitempty"`
	FullTitle                string  `json:"full_title"`
	HeaderImageThumbnailURL  string  `json:"header_image_thumbnail_url"`
	HeaderImageURL           string  `json:"header_image_url"`
	SongID                   int     `json:"id"`   // "id" is from genius.com
	ID                       string  `json:"uuid"` // UUID is a reserved word in DynamoDB
	Path                     string  `json:"path"`
	ReleaseDateForDisplay    string  `json:"release_date_for_display"`
	SongArtImageThumbnailURL string  `json:"song_art_image_thumbnail_url"`
	SongArtImageURL          string  `json:"song_art_image_url"`
	Title                    string  `json:"title"`
	URL                      string  `json:"url"`
}

type SongResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Song Song `json:"song"`
	} `json:"response"`
}

type Song struct {
	BulkSong
	Album  *Album   `json:"album"`
	Media  *Media   `json:"media"`
	Lyrics []string `json:"lyrics"` // only present after scraping, does not come from genius.com
}

type Album struct {
	APIPath               string `json:"api_path"`
	CoverArtURL           string `json:"cover_art_url"`
	FullTitle             string `json:"full_title"`
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	ReleaseDateForDisplay string `json:"release_date_for_display"`
	URL                   string `json:"url"`
}

type Media struct {
	Provider string `json:"provider"`
	Start    int    `json:"start"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

func SearchArtistId(artistName string, affiliations []string) ([]int, error) {
	artistMap := make(map[int]interface{})

	primaryArtistMap := searchPrimaryArtist(artistName)
	maps.Copy(artistMap, primaryArtistMap)

	for _, affiliation := range affiliations {
		affiliationMap := searchWithAffiliation(artistName, affiliation)
		maps.Copy(artistMap, affiliationMap)
	}

	artistIds := maps.Keys(artistMap)
	if len(artistIds) != 0 {
		slog.Info("Unique artists", "artists", artistMap)
		return artistIds, nil
	} else {
		return artistIds, errors.New("No search results.")
	}
}

func searchWithAffiliation(artistName, affiliation string) map[int]interface{} {
	affiliationMap := searchPrimaryArtist(affiliation)
	artistAndAffiliationMap := searchPrimaryArtist(fmt.Sprintf("%s & %s", artistName, affiliation))
	affiliationAndArtistMap := searchPrimaryArtist(fmt.Sprintf("%s & %s", affiliation, artistName))
	ftMap := searchPrimaryArtist(fmt.Sprintf("%s (Ft. %s)", affiliation, artistName))
	featMap := searchPrimaryArtist(fmt.Sprintf("%s (feat. %s)", affiliation, artistName))

	ret := make(map[int]interface{})
	maps.Copy(ret, affiliationMap)
	maps.Copy(ret, artistAndAffiliationMap)
	maps.Copy(ret, affiliationAndArtistMap)
	maps.Copy(ret, ftMap)
	maps.Copy(ret, featMap)

	return ret
}

func searchPrimaryArtist(artistName string) map[int]interface{} {
	url := "https://api.genius.com/search"
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	query := req.URL.Query()
	query.Add("q", artistName)
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	slog.Info("Request", "url", req.URL.String())
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed request.", "error", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", "error", err)
	}

	var data SearchResponse
	json.Unmarshal(body, &data)

	slog.Info(fmt.Sprintf("Searching for \"%s\"...", artistName))
	artistIdMap := make(map[int]interface{})

	for _, hit := range data.Response.Hits {
		isInArtistNames := strings.Contains(hit.Result.ArtistNames, artistName)
		isPrimaryArtist := strings.Contains(hit.Result.PrimaryArtist.Name, artistName)
		isFeaturedArtist := isFeaturedArtist(hit.Result.FeaturedArtists, artistName)

		if isInArtistNames || isPrimaryArtist || isFeaturedArtist {
			artistId := hit.Result.PrimaryArtist.ID
			slog.Info("Search result", "match", hit.Result.ArtistNames, slog.Int("artistId", artistId))
			artistIdMap[artistId] = hit.Result.PrimaryArtist.Name
		}
	}

	return artistIdMap
}

func isFeaturedArtist(featuredArtists []SearchResponseArtist, artistName string) bool {
	for _, featuredArtist := range featuredArtists {
		if featuredArtist.Name == artistName {
			return true
		}
	}

	return false
}

func RequestPage(artistId int, pageNumber int) ([]Song, *int) {
	path := fmt.Sprintf("/artists/%s/songs", strconv.Itoa(artistId))
	url := fmt.Sprintf("https://api.genius.com%s", path)
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	query := req.URL.Query()
	if pageNumber > 0 {
		query.Add("page", strconv.Itoa(pageNumber))
	}
	maxPageSize := "50"
	query.Add("per_page", maxPageSize)
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	slog.Info("Request", "url", req.URL.String())
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Reqeust failed.", "url", req.URL.String(), "error", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", "url", req.URL.String(), "error", err)
	}

	var data BulkSongsResponse
	json.Unmarshal(body, &data)

	songs := []Song{}

	for _, song := range data.Response.Songs {
		path = fmt.Sprintf("/songs/%d", song.SongID)
		url := fmt.Sprintf("https://api.genius.com%s", path)
		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		slog.Info("Request", "url", req.URL.String())

		res, err := client.Do(req)
		if err != nil {
			slog.Error("Request failed.", "url", req.URL.String(), "error", err)
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("Failed to read buffer.", "url", req.URL.String(), "error", err)
		}

		var data SongResponse
		json.Unmarshal(body, &data)
		song := data.Response.Song

		// skip songs without playable apple music provider
		if song.AppleMusicPlayerUrl != nil {
			songs = append(songs, song)
		}
	}

	return songs, data.Response.NextPage
}
