package genius

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

type SearchResponse struct {
	Meta struct {
		Status int `json:"status"`
	} `json:"meta"`
	Response struct {
		Hits []SearchHit `json:"hits"`
	} `json:"response"`
}

type SearchHit struct {
	Result struct {
		ArtistNames     string   `json:"artist_names"`
		PrimaryArtist   Artist   `json:"primary_artist"`
		FeaturedArtists []Artist `json:"featured_artists"`
	} `json:"result"`
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
	ArtistNames              string   `json:"artist_names"`
	AppleMusicPlayerUrl      *string  `json:"apple_music_player_url,omitempty"`
	FullTitle                string   `json:"full_title"`
	HeaderImageThumbnailURL  string   `json:"header_image_thumbnail_url"`
	HeaderImageURL           string   `json:"header_image_url"`
	ID                       int      `json:"id"`
	Path                     string   `json:"path"`
	ReleaseDateForDisplay    string   `json:"release_date_for_display"`
	SongArtImageThumbnailURL string   `json:"song_art_image_thumbnail_url"`
	SongArtImageURL          string   `json:"song_art_image_url"`
	Title                    string   `json:"title"`
	URL                      string   `json:"url"`
	PrimaryArtist            Artist   `json:"primary_artist"`
	FeaturedArtists          []Artist `json:"featured_artists"`
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
	Album *Album `json:"album"`
	Media *Media `json:"media"`
}

type Artist struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ApiPath        string `json:"api_path"`
	HeaderImageUrl string `json:"header_image_url"`
	ImageUrl       string `json:"image_url"`
	URL            string `json:"url"`
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

func Search(searchTerm string) SearchResponse {
	url := "https://api.genius.com/search"
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	query := req.URL.Query()
	query.Add("q", searchTerm)
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
	return data
}

func Songs(artistId int, pageNumber int) ([]Song, *int) {
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
		song := SongById(song.ID)
		if song.AppleMusicPlayerUrl != nil {
			songs = append(songs, song)
		}
	}

	return songs, data.Response.NextPage
}

func SongById(id int) Song {
	path := fmt.Sprintf("/songs/%d", id)
	url := fmt.Sprintf("https://api.genius.com%s", path)
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
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

	return data.Response.Song
}
