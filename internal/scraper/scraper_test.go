package scraper

import (
	"reflect"
	"testing"

	"github.com/jseashell/genius-lyrics-seed-service/internal/genius"
)

func TestParseAnchorLink(t *testing.T) {
	artistName := "foo"
	song := genius.Song{
		Album: &genius.Album{
			Name: "bar",
		},
	}
	html := "foo <a href=\"https://example.com\">example</a> bar"
	want := "foo example bar"
	got := Parse(artistName, song, html)[0]
	if want != got {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestParseLineBreak(t *testing.T) {
	artistName := "foo"
	song := genius.Song{
		Album: &genius.Album{
			Name: "bar",
		},
	}
	html := "foo<br/><a href=\"https://example.com\">bar"
	want := []string{"foo", "bar"}
	got := Parse(artistName, song, html)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v got %v", want, got)
	}
}

func TestSkipFeaturePart(t *testing.T) {
	artistName := "foo"
	featureArtist := "bar"
	song := genius.Song{
		Album: &genius.Album{
			Name: "bar",
		},
	}

	toSkip := "feature words"
	toParse := "artist lyric"

	html := "[Verse: " + featureArtist + "]<br/>" + toSkip + "<br/>[Chorus: " + artistName + "]<br/>" + toParse
	want := []string{toParse}
	got := Parse(artistName, song, html)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v got %v", want, got)
	}
}
