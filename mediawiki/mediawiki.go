package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/mediawiki"
)

const wikidataTestDump = "https://gitlab.com/tozd/go/mediawiki/-/raw/main/testdata/wikidata-testdata-all.json.bz2"

var (
	acceptedSites = map[string]bool{
		"enwiki": true,
		"ruwiki": true,
		"dewiki": true,
		"fawiki": true,
		"eswiki": true,
	}
	articlesToExtract = map[string]map[string]bool{
		"enwiki": {},
		"ruwiki": {},
		"dewiki": {},
		"fawiki": {},
		"eswiki": {},
	}
)

func latestURL() string {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		log.Printf("Request: %s %s Retry: %v", req.Method, req.URL, retry)
	}

	latestURL, err := mediawiki.LatestWikidataEntitiesRun(context.Background(), client)
	if err != nil {
		log.Fatal(err)
	}

	return latestURL
}

func isInterestedFeature(feature string) bool {
	// to be determined from .osm2ft file
	return feature == "Q31"
}

func parseMediaWikiJSONDump() map[string]map[string]bool {
	client := retryablehttp.NewClient()
	client.RequestLogHook = func(logger retryablehttp.Logger, req *http.Request, retry int) {
		log.Printf("Request: %s %s Retry: %v", req.Method, req.URL, retry)
	}

	cacheDir := os.TempDir()
	dumpPath := filepath.Join(cacheDir, path.Base(wikidataTestDump))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := mediawiki.ProcessWikidataDump(
		ctx,
		&mediawiki.ProcessDumpConfig{
			URL:    wikidataTestDump,
			Path:   dumpPath,
			Client: client,
		},
		func(_ context.Context, a mediawiki.Entity) errors.E {
			if !isInterestedFeature(a.ID) {
				return nil
			}

			for siteLink, info := range a.SiteLinks {
				if !acceptedSites[siteLink] {
					continue
				}

				articlesToExtract[siteLink][info.Title] = true
			}
			return nil
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	return articlesToExtract
}

func main() {
	t := time.Now()
	fmt.Printf("Latest URL: %s\n", latestURL())
	articlesToExtract := parseMediaWikiJSONDump()
	fmt.Printf("Time elapsed: %v\n", time.Since(t))
	for site, articles := range articlesToExtract {
		fmt.Printf("%s: %v\n", site, articles)
	}
}
