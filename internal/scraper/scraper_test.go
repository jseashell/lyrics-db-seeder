package scraper

import (
	"reflect"
	"testing"
)

func TestParseAnchorLink(t *testing.T) {
	html := "foo <a href=\"https://example.com\">example</a> bar"
	want := "foo example bar"
	got := Parse(html)[0]
	if want != got {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestParseLineBreak(t *testing.T) {
	html := "foo<br/><a href=\"https://example.com\">bar"
	want := []string{"foo", "bar"}
	got := Parse(html)

	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %v got %v", want, got)
	}
}
