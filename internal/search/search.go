package search

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
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

func Query(artistName string, affiliations []string) ([]int, error) {
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
	artistAndAffiliationMap := searchPrimaryArtist(fmt.Sprintf("%s and %s", artistName, affiliation))
	artistAndSymAffiliationMap := searchPrimaryArtist(fmt.Sprintf("%s & %s", artistName, affiliation))
	affiliationAndArtistMap := searchPrimaryArtist(fmt.Sprintf("%s and %s", affiliation, artistName))
	affiliationAndSymArtistMap := searchPrimaryArtist(fmt.Sprintf("%s & %s", affiliation, artistName))
	ftMap := searchPrimaryArtist(fmt.Sprintf("%s (Ft. %s)", affiliation, artistName))
	featMap := searchPrimaryArtist(fmt.Sprintf("%s (feat. %s)", affiliation, artistName))

	ret := make(map[int]interface{})
	maps.Copy(ret, affiliationMap)
	maps.Copy(ret, artistAndAffiliationMap)
	maps.Copy(ret, artistAndSymAffiliationMap)
	maps.Copy(ret, affiliationAndArtistMap)
	maps.Copy(ret, affiliationAndSymArtistMap)
	maps.Copy(ret, ftMap)
	maps.Copy(ret, featMap)

	return ret
}

func searchPrimaryArtist(artistName string) map[int]interface{} {
	searchResponse := genius.Search(artistName)

	slog.Info(fmt.Sprintf("Searching for \"%s\"...", artistName))
	artistIdMap := make(map[int]interface{})

	for _, hit := range searchResponse.Response.Hits {
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

func isFeaturedArtist(featuredArtists []genius.Artist, artistName string) bool {
	for _, featuredArtist := range featuredArtists {
		if featuredArtist.Name == artistName {
			return true
		}
	}

	return false
}
