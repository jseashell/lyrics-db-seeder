package genius

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
)

type SearchResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Hits []struct {
			Result struct {
				PrimaryArtist struct {
					ID   int    `json:"id"`
					Name string `json:"name"`
				} `json:"primary_artist"`
			} `json:"result"`
		} `json:"hits"`
	} `json:"response"`
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
	ArtistNames              string `json:"artist_names"`
	AppleMusicPlayerUrl      string `json:"apple_music_player_url"`
	FullTitle                string `json:"full_title"`
	HeaderImageThumbnailURL  string `json:"header_image_thumbnail_url"`
	HeaderImageURL           string `json:"header_image_url"`
	SongID                   int    `json:"id"`   // "id" is from genius.com
	ID                       string `json:"uuid"` // UUID is a reserved word in DynamoDB
	Path                     string `json:"path"`
	ReleaseDateForDisplay    string `json:"release_date_for_display"`
	SongArtImageThumbnailURL string `json:"song_art_image_thumbnail_url"`
	SongArtImageURL          string `json:"song_art_image_url"`
	Title                    string `json:"title"`
	URL                      string `json:"url"`
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

func SearchArtistId(artistName string) (int, error) {
	url := "https://api.genius.com/search"
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	query := req.URL.Query()
	query.Add("q", artistName)
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	slog.Info(fmt.Sprintf("Requesting %s", req.URL.String()))
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed request.", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", err)
	}

	var data SearchResponse
	json.Unmarshal(body, &data)

	slog.Info(fmt.Sprintf("Searching for \"%s\" artist ID...", artistName))
	for _, hit := range data.Response.Hits {
		if artistName == hit.Result.PrimaryArtist.Name {
			artistId := hit.Result.PrimaryArtist.ID
			slog.Info(fmt.Sprintf("Found! %d", artistId))
			return artistId, nil
		}
	}

	return -1, errors.New(fmt.Sprintf("Failed to find ID for artist \"%s\".", artistName))
}

func RequestPage(artistId int, artistAlbums []string, pageNumber int) ([]Song, *int) {
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
	slog.Info(fmt.Sprintf("Requesting %s", req.URL.String()))
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed request.", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", err)
	}

	var data BulkSongsResponse
	json.Unmarshal(body, &data)

	songs := []Song{}

	for _, song := range data.Response.Songs {
		path = fmt.Sprintf("/songs/%d", song.SongID)
		url := fmt.Sprintf("https://api.genius.com%s", path)
		req, _ = http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		slog.Info(fmt.Sprintf("Requesting %s", req.URL.String()))

		res, err := client.Do(req)
		if err != nil {
			slog.Error("Failed request.", err)
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			slog.Error("Failed to read buffer.", err)
		}

		var data SongResponse
		json.Unmarshal(body, &data)
		song := data.Response.Song

		if song.Album == nil {
			// skip singles
			slog.Warn(fmt.Sprintf("No album for song \"%s\"", song.Title))
			continue
		}

		albumName := (*song.Album).Name
		if slices.Contains(artistAlbums, albumName) {
			slog.Info(fmt.Sprintf("Found song \"%s\" from album \"%s\"", song.Title, albumName))
			songs = append(songs, song)
		} else {
			slog.Warn(fmt.Sprintf("Invalid album for song \"%s\" (%s)", song.Title, albumName))
		}

	}

	return songs, data.Response.NextPage
}
