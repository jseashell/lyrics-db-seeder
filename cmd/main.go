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
)

func main() {
	start := time.Now()

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	artistName := os.Getenv("GENIUS_ARTIST_NAME")
	artistAlbums := strings.Split(os.Getenv("GENIUS_ARTIST_ALBUMS"), ",")

	slog.Info(fmt.Sprintf("======== AWS Genius Lyrics ========\nArtist:\t\t%s\nAlbums (%d):\t%s", artistName, len(artistAlbums), strings.Join(artistAlbums, ", \n\t\t")))

	pageNumber := 0
	songs := make(map[int]genius.Song)

	artistId, err := genius.SearchArtistId(artistName)
	if err != nil {
		panic(err)
	}

	for {
		nextSongs, nextPage := genius.RequestPage(artistId, artistAlbums, pageNumber)
		for _, song := range nextSongs {
			lyrics := scraper.Run(artistName, song)

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
