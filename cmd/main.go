// Copyright 2024 John Schellinger.
// Use of this file is governed by the MIT license that can
// be found in the LICENSE.txt file in the project root.

// Package `main` integrates the Genius.com API and web scraping
// to capture all lyrics for a given artist.
package main

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jseashell/lyrics-db-seeder/internal/db"
	"github.com/jseashell/lyrics-db-seeder/internal/genius"
	"github.com/jseashell/lyrics-db-seeder/internal/logger"
	"github.com/jseashell/lyrics-db-seeder/internal/scraper"
	"github.com/jseashell/lyrics-db-seeder/internal/search"
)

func main() {
	start := time.Now()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	artistName := os.Getenv("GENIUS_PRIMARY_ARTIST")
	affiliations := strings.Split(os.Getenv("GENIUS_AFFILIATIONS"), ",")

	logger := logger.New()
	slog.SetDefault(logger)

	artistIds, err := search.Query(artistName, affiliations)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, id := range artistIds {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processArtistId(artistName, id)
		}(id)
	}
	wg.Wait()

	t := time.Now()
	elapsed := t.Sub(start)
	slog.Info("Seed complete", slog.Float64("seconds", elapsed.Seconds()))
}

func processArtistId(artistName string, artistId int) {
	pageNumber := 0

	songs := sync.Map{}
	var wg sync.WaitGroup
	for {
		nextSongs, nextPage := genius.Songs(artistId, artistName, pageNumber)
		wg.Add(1)
		go func(toProcess []genius.SongWithExtras) {
			defer wg.Done()
			scrapedSongs := processPage(artistName, toProcess)
			for _, scrapedSong := range scrapedSongs {
				songs.Store(scrapedSong.Song.ID, scrapedSong)
			}
		}(nextSongs)

		if nextPage == nil {
			break
		} else {
			pageNumber = *nextPage
		}
	}
	wg.Wait()

	songs.Range(func(key any, value any) bool {
		db.PutSong(value.(scraper.ScrapedSong))
		return true
	})

}

func processPage(artistName string, nextSongs []genius.SongWithExtras) []scraper.ScrapedSong {
	songs := []scraper.ScrapedSong{}

	var wg sync.WaitGroup
	mu := &sync.Mutex{}
	for _, nextSong := range nextSongs {
		wg.Add(1)
		go func(artistName string, nextSong genius.SongWithExtras) {
			defer wg.Done()

			lyrics := scraper.Run(artistName, nextSong)
			if len(lyrics) > 0 {
				scrapedSong := scraper.ScrapedSong{
					Song:   nextSong,
					ID:     uuid.NewString(), // uuid is for fetching random song from AWS DynamoDB
					Lyrics: lyrics,
				}

				mu.Lock()
				songs = append(songs, scrapedSong)
				mu.Unlock()
			}

		}(artistName, nextSong)
	}
	wg.Wait()

	return songs
}
