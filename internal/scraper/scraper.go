package scraper

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/gocolly/colly"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
	"github.com/microcosm-cc/bluemonday"
)

func Run(artistName string, song genius.Song) *[]string {
	lyrics := &[]string{}
	selector := "div[data-lyrics-container=\"true\"]"

	c := colly.NewCollector()
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		html, _ := e.DOM.Html()
		nextLyrics := Parse(artistName, song, html)
		*lyrics = append(*lyrics, nextLyrics...)
	})
	c.Visit(song.URL)
	c.Wait()

	return lyrics
}

func Parse(artistName string, song genius.Song, html string) []string {
	html = strings.ReplaceAll(html, "<br/>", "\n")

	p := bluemonday.NewPolicy()
	html = p.Sanitize(html)

	lines := strings.Split(html, "\n")

	lyrics := []string{}
	featurePart := false
	for _, line := range lines {
		// Skip blank lines
		if line == "" {
			continue
		}

		// Skip featured artist verses
		if isMetaLine(line) && strings.Contains(line, ":") {
			featurePart = !strings.Contains(line, artistName)
			continue
		}

		// Parse lyrics from parts of the song by the defined artist
		if !isMetaLine(line) && !featurePart {
			trimmed := strings.Trim(line, " ")
			trimmed = strings.ReplaceAll(trimmed, "[?]", "___")
			trimmed = strings.TrimPrefix(trimmed, "[")
			trimmed = strings.TrimSuffix(trimmed, "]")
			trimmed = strings.ReplaceAll(trimmed, "’", "'")
			trimmed = strings.ReplaceAll(trimmed, "&#39;", "'")
			trimmed = strings.ReplaceAll(trimmed, "&#34;", "\"")
			lyrics = append(lyrics, trimmed)
			slog.Info(fmt.Sprintf("Scraped from %s (%s)\n%s", song.Title, song.Album.Name, trimmed))
		}
	}

	return lyrics
}

func isMetaLine(line string) bool {
	return strings.Contains(line, "[Intro") || strings.Contains(line, "[Verse") || strings.Contains(line, "[Chorus") || strings.Contains(line, "[Hook") || strings.Contains(line, "[Bridge") || strings.Contains(line, "[Outro")
}
