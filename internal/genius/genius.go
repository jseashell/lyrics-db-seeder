// Copyright 2024 John Schellinger.
// Use of this file is governed by the MIT license that can
// be found in the LICENSE.txt file in the project root.

// Package `genius` encapsulates functions for interacting with the Genius.com API.
// Requires the `GENIUS_ACCESS_TOKEN` environment variable for authorization.
package genius

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Response metadata
type GeniusMeta struct {
	// HTTP status code
	Status int `json:"status"`
}

// Response for a /search request
type SearchResponse struct {
	Meta     GeniusMeta `json:"meta"`
	Response struct {
		// List of search results
		Hits []SearchHit `json:"hits"`
	} `json:"response"`
}

// A search result
type SearchHit struct {
	Result struct {
		// All artists on the song
		ArtistNames string `json:"artist_names"`
		// Primary [Artist] credited to the song
		PrimaryArtist Artist `json:"primary_artist"`
		// List of featured [Artist]s credited to the song
		FeaturedArtists []Artist `json:"featured_artists"`
	} `json:"result"`
}

// Response for a /artist/:id/songs request
type SongsResponse struct {
	Meta     GeniusMeta `json:"meta"`
	Response struct {
		// List of songs for a given artist ID
		Songs []Song `json:"songs"`
		// Pagination parameter
		NextPage *int `json:"next_page"`
	} `json:"response"`
}

// A song
type Song struct {
	// All artists on the song
	ArtistNames string `json:"artist_names"`
	// URL for embedable Apple Music player
	AppleMusicPlayerUrl *string `json:"apple_music_player_url,omitempty"`
	// Song full title
	FullTitle string `json:"full_title"`
	// Genius.com header image, typically 1:1 aspect ratio
	HeaderImageThumbnailURL string `json:"header_image_thumbnail_url"`
	// Larger version of [HeaderImageThumbnailURL]
	HeaderImageURL string `json:"header_image_url"`
	// Unique identifier
	ID int `json:"id"`
	// Human-readable release date
	ReleaseDateForDisplay string `json:"release_date_for_display"`
	// Song/Album art, typically 1:1 aspect ratio
	SongArtImageThumbnailURL string `json:"song_art_image_thumbnail_url"`
	// Larger version of [SongArtImageThumbnailURL]
	SongArtImageURL string `json:"song_art_image_url"`
	// Song title
	Title string `json:"title"`
	// Absolute Genius.com URL
	URL string `json:"url"`
	// Path relative to [https://genius.com]
	Path string `json:"path"`
	// Primary [Artist] credited for the song
	PrimaryArtist Artist `json:"primary_artist"`
	// List of featured [Artist]s credited to the song
	FeaturedArtists []Artist `json:"featured_artists"`
}

// Response for a /songs/:id request
type SongByIdResponse struct {
	Meta     GeniusMeta `json:"meta"`
	Response struct {
		// Song identified by the given song ID
		Song SongWithExtras `json:"song"`
	} `json:"response"`
}

// A [Song] with [Album] and [Media] extras
type SongWithExtras struct {
	Song
	// The [Album] to which this song belongs
	Album *Album `json:"album"`
	// External [Media] for consuming this song
	Media *Media `json:"media"`
}

// Person responsible for composing, recording, and/or performing a [Song]
type Artist struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	ApiPath        string `json:"api_path"`
	HeaderImageUrl string `json:"header_image_url"`
	ImageUrl       string `json:"image_url"`
	URL            string `json:"url"`
}

// An entity representing a collection of [Song]s. However, this application
// does not attempt to group songs within this stucture. This information
// is purely metadata about an individual song.
type Album struct {
	APIPath               string `json:"api_path"`
	CoverArtURL           string `json:"cover_art_url"`
	FullTitle             string `json:"full_title"`
	ID                    int    `json:"id"`
	Name                  string `json:"name"`
	ReleaseDateForDisplay string `json:"release_date_for_display"`
	URL                   string `json:"url"`
}

// A consumable piece of digital media representing a [Song]
type Media struct {
	Provider string `json:"provider"`
	Start    int    `json:"start"`
	Type     string `json:"type"`
	URL      string `json:"url"`
}

// Searches the Genius.com API for the given search term
func Search(searchTerm string) SearchResponse {
	path := "/search"
	url := fmt.Sprintf("https://api.genius.com%s", path)
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	query := req.URL.Query()
	query.Add("q", searchTerm)
	req.URL.RawQuery = query.Encode()

	client := &http.Client{}
	slog.Debug("Search", "url", req.URL.String())
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

// Fetches a page of songs for a given artist via GET request to the Genius.com API.
// Subsequently loops over each song to fetch its metadata via go routines. Only songs
// that have the given artist present as a primary or featured artist are returned.
func Songs(artistId int, artistName string, pageNumber int) ([]SongWithExtras, *int) {
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
	slog.Debug("Songs", "url", req.URL.String())
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Reqeust failed.", "url", req.URL.String(), "error", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", "url", req.URL.String(), "error", err)
	}

	var data SongsResponse
	json.Unmarshal(body, &data)

	songs := []SongWithExtras{}

	var wg sync.WaitGroup
	mu := sync.Mutex{}
	for _, song := range data.Response.Songs {
		wg.Add(1)
		go func(songId int) {
			defer wg.Done()
			song := SongById(songId)

			if strings.Contains(song.PrimaryArtist.Name, artistName) {
				slog.Info("As primary artist", "song", song)

				mu.Lock()
				songs = append(songs, song)
				mu.Unlock()

				return
			} else {

				foundFeature := false
				for _, feature := range song.FeaturedArtists {
					if strings.Contains(feature.Name, artistName) {
						slog.Info("As featured artist", "song", song)

						mu.Lock()
						songs = append(songs, song)
						mu.Unlock()

						foundFeature = true
						break
					}
				}

				if !foundFeature {
					slog.Debug("Not a contributor", "song", song)
				}
			}
		}(song.ID)
	}
	wg.Wait()

	return songs, data.Response.NextPage
}

// Fetches a song identified by the given ID via GET request to the Genius.com API.
func SongById(id int) SongWithExtras {
	path := fmt.Sprintf("/songs/%d", id)
	url := fmt.Sprintf("https://api.genius.com%s", path)
	req, _ := http.NewRequest("GET", url, nil)

	accessToken := os.Getenv("GENIUS_ACCESS_TOKEN")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		slog.Error("Request failed.", "url", req.URL.String(), "error", err)
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Failed to read buffer.", "url", req.URL.String(), "error", err)
	}

	var data SongByIdResponse
	json.Unmarshal(body, &data)

	slog.Debug("SongById", "url", req.URL.String(), "res", data.Response.Song)

	return data.Response.Song
}
