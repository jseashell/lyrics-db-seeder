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
)

func main() {
	start := time.Now()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	artistId, _ := strconv.Atoi(os.Getenv("GENIUS_ARTIST_ID"))
	pageNumber := 0

	for {
		songs, nextPage := genius.RequestPage(artistId, pageNumber)

		for i := 0; i < len(songs); i++ {
			song := songs[i]
			res := db.PutSong(song)
			if res == 0 {
				scraper.Run(song)
			}
		}

		if nextPage == nil {
			break
		} else {
			pageNumber = *nextPage
		}
	}

	t := time.Now()
	elapsed := t.Sub(start)
	slog.Info(fmt.Sprintf("Seeded in %f seconds", elapsed.Seconds()))
}
