package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/walle/targz"
)

const (
	enterpriseTestDump = "adywiki-NS14-20230220-ENTERPRISE-HTML.json.tar.gz"
	// Need to determine how/if there is a package which mimics the functionality of the "--extract-file" flag from the tar command
	extractedName = "adywiki_0.ndjson"
)

type Article struct {
	Name         string    `json:"name"`
	Identifier   int       `json:"identifier"`
	DateModified time.Time `json:"date_modified"`
	Version      struct {
		Identifier  int    `json:"identifier"`
		Comment     string `json:"comment"`
		IsMinorEdit bool   `json:"is_minor_edit"`
		Editor      struct {
			Identifier int    `json:"identifier"`
			Name       string `json:"name"`
		} `json:"editor"`
	} `json:"version"`
	URL       string `json:"url"`
	Namespace struct {
		Name       string `json:"name"`
		Identifier int    `json:"identifier"`
	} `json:"namespace"`
	InLanguage struct {
		Name       string `json:"name"`
		Identifier string `json:"identifier"`
	} `json:"in_language"`
	MainEntity struct {
		Identifier string `json:"identifier"`
		URL        string `json:"url"`
	} `json:"main_entity"`
	AdditionalEntities []struct {
		Identifier string   `json:"identifier"`
		URL        string   `json:"url"`
		Aspects    []string `json:"aspects"`
	} `json:"additional_entities"`
	Categories []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"categories"`
	IsPartOf struct {
		Name       string `json:"name"`
		Identifier string `json:"identifier"`
	} `json:"is_part_of"`
	ArticleBody struct {
		HTML     time.Time `json:"html"`
		Wikitext string    `json:"wikitext"`
	} `json:"article_body"`
	License []struct {
		Name       string `json:"name"`
		Identifier string `json:"identifier"`
		URL        string `json:"url"`
	} `json:"license,omitempty"`
}

func main() {
	if err := targz.Extract(enterpriseTestDump, "."); err != nil {
		panic(err)
	}

	f, err := os.Open(extractedName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// read and marshal the ndjson file
	decoder := json.NewDecoder(f)

	for decoder.More() {
		var article Article
		decoder.Decode(&article)

		if article.Name == "" {
			log.Println("empty article")
			continue
		}

		fmt.Println(article.ArticleBody.HTML)
	}

}
