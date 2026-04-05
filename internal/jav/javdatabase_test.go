package jav

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/net/html"
)

func TestParseJavDatabaseMovieInfo(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(`
<!doctype html>
<html>
<head><title>IPX-004 - Tsumugi Akari - JAV Database</title></head>
<body>
  <div class="movietable" style="padding-top: 5px;">
    <div class="row">
      <div class="col-md-2 col-lg-2 col-xxl-2 col-4"></div>
      <div class="col-md-10 col-lg-10 col-xxl-10 col-8">
        <p class="mb-1"><b>Title: </b>Together With A Miraculous Beautiful Girl</p>
        <p class="mb-1"><b>DVD ID: </b>IPX-004</p>
        <p class="mb-1"><b>Release Date: </b>2017-09-09</p>
        <p class="mb-1"><b>Runtime: </b>159  (HD: 159) min.</p>
        <p class="mb-1"><b>Genre(s): </b><span><a href="/genres/a">Beautiful Girl</a></span> <span><a href="/genres/b">Hi-Def</a></span></p>
        <p class="mb-1"><b>Idol(s)/Actress(es): </b><span><a href="/idols/tsumugi-akari/">Tsumugi Akari</a></span></p>
      </div>
    </div>
  </div>
</body>
</html>`))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	info := parseJavDatabaseMovieInfo(doc)
	if info == nil {
		t.Fatal("expected info, got nil")
	}

	if info.Title != "Together With A Miraculous Beautiful Girl" {
		t.Fatalf("unexpected title: %q", info.Title)
	}
	if info.Code != "IPX-004" {
		t.Fatalf("unexpected code: %q", info.Code)
	}

	wantRelease := time.Date(2017, 9, 9, 0, 0, 0, 0, time.UTC).Unix()
	if info.ReleaseUnix != wantRelease {
		t.Fatalf("unexpected release unix: got %d want %d", info.ReleaseUnix, wantRelease)
	}
	if info.DurationMin != 159 {
		t.Fatalf("unexpected duration: %d", info.DurationMin)
	}

	wantTags := []string{"Beautiful Girl", "Hi-Def"}
	if len(info.Tags) != len(wantTags) {
		t.Fatalf("unexpected tags length: got %d want %d", len(info.Tags), len(wantTags))
	}
	for i, tag := range wantTags {
		if info.Tags[i] != tag {
			t.Fatalf("unexpected tag at %d: got %q want %q", i, info.Tags[i], tag)
		}
	}

	wantActors := []string{"Tsumugi Akari"}
	if len(info.Actors) != len(wantActors) {
		t.Fatalf("unexpected actors length: got %d want %d", len(info.Actors), len(wantActors))
	}
	for i, actor := range wantActors {
		if info.Actors[i] != actor {
			t.Fatalf("unexpected actor at %d: got %q want %q", i, info.Actors[i], actor)
		}
	}
}
