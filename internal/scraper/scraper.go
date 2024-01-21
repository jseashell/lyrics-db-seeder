package scraper

import (
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"github.com/jseashell/genius-lyrics-seed-service/internal/db"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
)

func Run(song genius.Song) {
	c := colly.NewCollector()

	c.OnHTML("div[data-lyrics-container=\"true\"]", func(e *colly.HTMLElement) {
		html, _ := e.DOM.Html()
		lyrics := parse(html, song)

		for j := 0; j < len(lyrics); j++ {
			db.PutLyric(lyrics[j])
		}
	})

	c.Visit(song.URL)
}

func parse(html string, song genius.Song) []genius.Lyric {
	html = strings.ReplaceAll(html, "<br/>", "\n")

	html = regexp.MustCompile(`<a.*>.*</a>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<a.*><span.*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</a><span.*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`</span>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<span.*>.*</span>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<span.*><span.*>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<i.*>.*</i>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`<b.*>.*</b>`).ReplaceAllString(html, "")
	html = regexp.MustCompile(`\(.*\)`).ReplaceAllString(html, "")

	html = strings.ReplaceAll(html, "&#39;", "'")

	lines := strings.Split(html, "\n")

	var lyrics []genius.Lyric
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if line != "" && len(line) > 20 && !strings.Contains(line, "[Intro") && !strings.Contains(line, "[Verse") && !strings.Contains(line, "[Chorus") && !strings.Contains(line, "[Bridge") && !strings.Contains(line, "[Outro") {
			trimmed := strings.Trim(line, " ")

			id := uuid.NewString()

			lyrics = append(lyrics, genius.Lyric{
				ID:     id,
				SongID: song.ID,
				Value:  trimmed,
			})
		}
	}

	return lyrics
}
