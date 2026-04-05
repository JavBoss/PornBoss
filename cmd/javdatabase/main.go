package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"pornboss/internal/jav"
)

func main() {
	var code string
	flag.StringVar(&code, "code", "", "movie code (e.g. IPX-004)")
	flag.Parse()

	if code == "" && flag.NArg() > 0 {
		code = flag.Arg(0)
	}
	if code == "" {
		log.Fatal("usage: go run ./cmd/javdatabase -code IPX-004")
	}

	info, err := jav.JavDatabaseProvider.LookupJavByCode(code)
	if err != nil {
		if errors.Is(err, jav.ResourceNotFonud) {
			fmt.Println("movie not found")
			return
		}
		log.Fatalf("lookup error: %v", err)
	}
	if info == nil {
		fmt.Println("movie not found")
		return
	}

	release := "unknown"
	if info.ReleaseUnix > 0 {
		release = time.Unix(info.ReleaseUnix, 0).UTC().Format("2006-01-02")
	}

	duration := "unknown"
	if info.DurationMin > 0 {
		duration = fmt.Sprintf("%d min", info.DurationMin)
	}

	tags := strings.Join(info.Tags, ", ")
	if tags == "" {
		tags = "None"
	}

	actors := strings.Join(info.Actors, ", ")
	if actors == "" {
		actors = "None"
	}

	fmt.Printf("Code: %s\n", info.Code)
	fmt.Printf("Title: %s\n", info.Title)
	fmt.Printf("Release: %s\n", release)
	fmt.Printf("Duration: %s\n", duration)
	fmt.Printf("Tags: %s\n", tags)
	fmt.Printf("Actors: %s\n", actors)
}
