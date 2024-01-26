package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/jseashell/genius-lyrics-seed-service/internal/db"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
	"github.com/jseashell/genius-lyrics-seed-service/internal/scraper"
	"golang.org/x/exp/maps"
)

func main() {
	start := time.Now()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	artistId, _ := strconv.Atoi(os.Getenv("GENIUS_ARTIST_ID"))
	pageNumber := 0

	lyrics := make(map[int]genius.Lyric)

	songs := []genius.Song{}

	for {
		nextSongs, nextPage := genius.RequestPage(artistId, pageNumber)
		songs = append(songs, nextSongs...)

		for i := 0; i < len(nextSongs); i++ {
			nextSong := nextSongs[i]
			slog.Info(fmt.Sprintf("%d lyrics accumulated. Scraping \"%s\" next.", len(maps.Values(lyrics)), nextSong.Title))
			scraper.Run(nextSong, lyrics)
		}

		if nextPage == nil {
			break
		} else {
			pageNumber = *nextPage
		}
	}

	slog.Info(fmt.Sprintf("Inserting %d songs", len(songs)))
	for i := 0; i < len(songs); i++ {
		db.PutSong(songs[i])
		if i%100 == 0 {
			fmt.Print(".")
		}
	}

	lyricsArray := maps.Values(lyrics)
	slog.Info(fmt.Sprintf("Inserting %d lyrics", len(songs)))
	for i := 0; i < len(lyricsArray); i++ {
		db.PutLyric(lyricsArray[i])
		if i%100 == 0 {
			fmt.Print(".")
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)
	slog.Info(fmt.Sprintf("Seeded in %f seconds", elapsed.Seconds()))
}
