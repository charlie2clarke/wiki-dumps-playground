package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/paulmach/osm"
	"github.com/paulmach/osm/osmpbf"
)

const (
	WikipediaTag = "wikipedia"
	WikidataTag  = "wikidata"
)

var (
	wikipediaTags = map[string]bool{}
	wikidataTags  = map[string]bool{}
)

func parseTags(tags osm.Tags) {
	for _, t := range tags {
		if t.Key == WikipediaTag {
			wikipediaTags[t.Value] = true
		} else if t.Key == WikidataTag {
			wikidataTags[t.Value] = true
		}
	}
}

func main() {
	t := time.Now()

	f, err := os.Open("./yukon-latest.osm.pbf")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := osmpbf.New(context.Background(), f, 3)
	defer scanner.Close()

	for scanner.Scan() {
		o := scanner.Object()
		switch vtype := o.(type) {
		case *osm.Node:
			parseTags(o.(*osm.Node).Tags)
		case *osm.Way:
			parseTags(o.(*osm.Way).Tags)
		case *osm.Relation:
			parseTags(o.(*osm.Relation).Tags)
		default:
			log.Fatalf("unknown type %T", vtype)
		}
	}

	scanErr := scanner.Err()
	if scanErr != nil {
		log.Fatal(scanErr)
	}

	log.Printf("Time: %v\n", time.Since(t))
	log.Printf("Wikipedia tags: %v\n", wikipediaTags)
	log.Printf("Wikidata tags: %v\n", wikidataTags)
}
