// Copyright 2024 John Schellinger.
// Use of this file is governed by the MIT license that can
// be found in the LICENSE.txt file in the project root.

// Package `search` contains logic to search for an artist and all their affiliations.
package search

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jseashell/lyrics-db-seeder/internal/genius"
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

func Query(artistName string, affiliations []string, includeFeatured bool, includeAnded bool) ([]int, error) {
	artistMap := make(map[int]interface{})

	primaryArtistMap := search(artistName, artistName)
	maps.Copy(artistMap, primaryArtistMap)

	for _, affiliation := range affiliations {
		affiliationMap := searchWithAffiliation(artistName, affiliation, includeFeatured, includeFeatured)
		maps.Copy(artistMap, affiliationMap)
	}

	artistIds := maps.Keys(artistMap)
	if len(artistIds) != 0 {
		slog.Info(fmt.Sprintf("Found %d unique artists", len(artistMap)), "artists", artistMap)
		return artistIds, nil
	} else {
		return artistIds, errors.New("no search results")
	}
}

// Runs a [search] for an affiliated contributor, e.g. "Other Artist (Ft. My Artist)"
func searchWithAffiliation(artistName, affiliation string, includeFeatured bool, includeAnded bool) map[int]interface{} {
	ret := make(map[int]interface{})

	affiliationMap := search(artistName, affiliation)
	maps.Copy(ret, affiliationMap)

	if includeAnded {
		artistAndAffiliationMap := search(artistName, fmt.Sprintf("%s and %s", artistName, affiliation))
		maps.Copy(ret, artistAndAffiliationMap)

		artistAndSymAffiliationMap := search(artistName, fmt.Sprintf("%s & %s", artistName, affiliation))
		maps.Copy(ret, artistAndSymAffiliationMap)

		affiliationAndArtistMap := search(artistName, fmt.Sprintf("%s and %s", affiliation, artistName))
		maps.Copy(ret, affiliationAndArtistMap)

		affiliationAndSymArtistMap := search(artistName, fmt.Sprintf("%s & %s", affiliation, artistName))
		maps.Copy(ret, affiliationAndSymArtistMap)
	}

	if includeFeatured {
		ftMap := search(artistName, fmt.Sprintf("%s (Ft. %s)", affiliation, artistName))
		maps.Copy(ret, ftMap)

		ft2Map := search(artistName, fmt.Sprintf("%s (ft. %s)", affiliation, artistName))
		maps.Copy(ret, ft2Map)

		featMap := search(artistName, fmt.Sprintf("%s (feat. %s)", affiliation, artistName))
		maps.Copy(ret, featMap)
	}

	return ret
}

// Searches Genius.com for the given search string. Attempts to match results to the given artist name
func search(artistName string, search string) map[int]interface{} {
	searchResponse := genius.Search(search)

	artistIdMap := make(map[int]interface{})

	for _, hit := range searchResponse.Response.Hits {
		isInArtistNames := strings.Contains(hit.Result.ArtistNames, artistName)
		isPrimaryArtist := strings.Contains(hit.Result.PrimaryArtist.Name, artistName)
		isFeaturedArtist := isFeaturedArtist(hit.Result.FeaturedArtists, artistName)

		artistId := hit.Result.PrimaryArtist.ID
		if isInArtistNames || isPrimaryArtist || isFeaturedArtist {
			slog.Info("Match", "artist_names", hit.Result.ArtistNames, slog.Int("artist_id", artistId))
			artistIdMap[artistId] = hit.Result.PrimaryArtist.Name
		} else {
			slog.Debug("No match", "searching", fmt.Sprintf("%s (%d)", search, artistId), "artist_names", hit.Result.ArtistNames, "primary_artist", hit.Result.PrimaryArtist.Name, "featured_artists", hit.Result.FeaturedArtists)
		}
	}

	return artistIdMap
}

func isFeaturedArtist(featuredArtists []genius.Artist, artistName string) bool {
	for _, featuredArtist := range featuredArtists {
		if featuredArtist.Name == artistName {
			return true
		}
	}

	return false
}
