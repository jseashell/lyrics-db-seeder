package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jseashell/genius-lyrics-seed-service/internal/db"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
	"github.com/jseashell/genius-lyrics-seed-service/internal/scraper"
	"github.com/jseashell/genius-lyrics-seed-service/internal/search"
)

func main() {
	start := time.Now()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	artistName := os.Getenv("GENIUS_PRIMARY_ARTIST")
	affiliations := strings.Split(os.Getenv("GENIUS_AFFILIATIONS"), ",")

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	artistIds, err := search.Query(artistName, affiliations)
	if err != nil {
		panic(err)
	}

	for _, id := range artistIds {
		processArtistId(artistName, id)
	}

	t := time.Now()
	elapsed := t.Sub(start)
	slog.Info("Seeded", slog.Float64("seconds", elapsed.Seconds()))
}

func processArtistId(artistName string, artistId int) {
	pageNumber := 0
	songs := make(map[int]scraper.ScrapedSong)

	for {
		nextSongs, nextPage := genius.Songs(artistId, pageNumber)
		for _, song := range nextSongs {
			lyrics := scraper.Run(artistName, song)

			scrapedSong := scraper.ScrapedSong{
				Song:   song,
				ID:     uuid.NewString(), // uuid is for fetching random song from AWS DynamoDB
				Lyrics: *lyrics,
			}

			songs[scrapedSong.Song.ID] = scrapedSong
		}

		if nextPage == nil {
			break
		} else {
			pageNumber = *nextPage
		}
	}

	slog.Info("Inserting songs", slog.Int("length", len(songs)))
	for i, song := range songs {
		db.PutSong(song)
		if i%50 == 0 {
			fmt.Print(".")
		}
	}
	fmt.Print("\n")
}
