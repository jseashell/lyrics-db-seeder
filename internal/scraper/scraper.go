package scraper

import (
	"hash/fnv"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
)

func Run(song genius.Song, lyricsMap map[int]genius.Lyric) {
	c := colly.NewCollector()

	c.OnHTML("div[data-lyrics-container=\"true\"]", func(e *colly.HTMLElement) {
		html, _ := e.DOM.Html()
		lyrics := parse(html, song)

		for j := 0; j < len(lyrics); j++ {
			lyric := lyrics[j]

			h := fnv.New32a()
			h.Write([]byte(lyric.Value))
			hash := int(h.Sum32())

			// lyrics with the same hash are not duplicated
			_, ok := lyricsMap[hash]
			if !ok {
				lyricsMap[hash] = lyric
			}
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
			trimmed = strings.TrimPrefix(trimmed, "[")
			trimmed = strings.TrimSuffix(trimmed, "]")
			trimmed = strings.ReplaceAll(trimmed, "’", "'")

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
