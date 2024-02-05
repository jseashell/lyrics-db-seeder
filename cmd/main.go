package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/jseashell/genius-lyrics-seed-service/internal/db"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
	"github.com/jseashell/genius-lyrics-seed-service/internal/scraper"
)

func main() {
	start := time.Now()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	artistId, _ := strconv.Atoi(os.Getenv("GENIUS_ARTIST_ID"))
	artistName := os.Getenv("GENIUS_ARTIST_NAME")
	pageNumber := 0
	songs := make(map[int]genius.Song)

	for {
		nextSongs, nextPage := genius.RequestPage(artistId, artistName, pageNumber)
		for _, song := range nextSongs {
			lyrics := scraper.Run(song)

			song.Lyrics = append(song.Lyrics, *lyrics...)
			song.ID = uuid.NewString()
			songs[song.SongID] = song

			slog.Info(fmt.Sprintf("Found %d lyrics for \"%s\".", len(song.Lyrics), song.Title))
		}

		if nextPage == nil { // || pageNumber > 2 {
			break
		} else {
			pageNumber = *nextPage
		}
	}

	slog.Info(fmt.Sprintf("Inserting %d songs", len(songs)))
	for i, song := range songs {
		db.PutSong(song)
		if i%50 == 0 {
			fmt.Print(".")
		}
	}
	fmt.Print("\n")

	t := time.Now()
	elapsed := t.Sub(start)
	slog.Info(fmt.Sprintf("Seeded in %f seconds", elapsed.Seconds()))
}
