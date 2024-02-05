package scraper

import (
	"strings"

	"github.com/gocolly/colly"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
	"github.com/microcosm-cc/bluemonday"
)

func Run(song genius.Song) *[]string {
	lyrics := &[]string{}
	selector := "div[data-lyrics-container=\"true\"]"

	c := colly.NewCollector()
	c.OnHTML(selector, func(e *colly.HTMLElement) {
		html, _ := e.DOM.Html()
		nextLyrics := Parse(html)
		*lyrics = append(*lyrics, nextLyrics...)
		// slog.Info(fmt.Sprintf("Scraped \"%s\". Received\n%v.", song.Title, strings.Join(*lyrics, "\n\t")))
	})
	c.Visit(song.URL)
	c.Wait()

	return lyrics
}

func Parse(html string) []string {
	html = strings.ReplaceAll(html, "<br/>", "\n")

	p := bluemonday.NewPolicy()
	html = p.Sanitize(html)
	html = strings.ReplaceAll(html, "&#39;", "'")

	lines := strings.Split(html, "\n")

	lyrics := []string{}
	for _, line := range lines {
		if line != "" && !strings.Contains(line, "[Intro") && !strings.Contains(line, "[Verse") && !strings.Contains(line, "[Chorus") && !strings.Contains(line, "[Hook") && !strings.Contains(line, "[Bridge") && !strings.Contains(line, "[Outro") {
			trimmed := strings.Trim(line, " ")
			trimmed = strings.TrimPrefix(trimmed, "[")
			trimmed = strings.TrimSuffix(trimmed, "]")
			trimmed = strings.ReplaceAll(trimmed, "’", "'")
			lyrics = append(lyrics, trimmed)
		}
	}
	return lyrics
}
