package main

import (
	"bufio"
	"compress/bzip2"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dustin/go-wikiparse"
)

type IndexedParseSource struct {
	articleFile *os.File
	indexFile   *os.File
}

type ArticleID int64

const NEW_INDEX_FILE = "new-index.txt"

var (
	indexWikiFile   string
	articleWikiFile string
	outputPath      string
	workerNum       int

	// In operation, this may be best using a key-value store like Redis or LevelDB.
	articles = map[ArticleID]bool{
		41296: true, // 601:41296:Jitter
		42433: true, // 4507917:42433:Cambridge, England
		42230: true, // 3520081:42230:Hotel Chelsea
	}
)

func init() {
	// This has been tested using the following files (my internet is too slow to push them...):
	// https://dumps.wikimedia.org/enwiki/20230320/enwiki-20230320-pages-articles-multistream-index2.txt-p41243p151573.bz2
	// https://dumps.wikimedia.org/enwiki/20230320/enwiki-20230320-pages-articles-multistream2.xml-p41243p151573.bz2

	flag.StringVar(&indexWikiFile, "index", "enwiki-20230320-pages-articles-multistream-index2.txt-p41243p151573.bz2", "index wiki file")
	flag.StringVar(&articleWikiFile, "article", "enwiki-20230320-pages-articles-multistream2.xml-p41243p151573.bz2", "article wiki file")
	flag.StringVar(&outputPath, "output", "", "output path")
	flag.IntVar(&workerNum, "worker", 1, "number of workers")
}

func parseFlags() error {
	flag.Parse()
	if indexWikiFile == "" || articleWikiFile == "" {
		return errors.New("index and article wiki file must be specified")
	}

	if outputPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil
		}
		outputPath = wd
	}

	return nil
}

func (ips *IndexedParseSource) OpenIndex() (io.ReadCloser, error) {
	// Create a closure that wraps the bzip2 reader and also closes the index file.
	rc := func() io.ReadCloser {
		return struct {
			io.Reader
			io.Closer
		}{
			Reader: bufio.NewReader(ips.indexFile),
			Closer: ips.indexFile,
		}
	}()

	return rc, nil
}

// go-wikiparse prematurely closes the articleFile, so temporarily override the Close method.
type NoopCloser struct{}

func (n NoopCloser) Close() error {
	return nil
}

func (ips *IndexedParseSource) OpenData() (wikiparse.ReadSeekCloser, error) {
	// Create a closure that wraps the bzip2 reader and includes seek and close methods.
	rc := func() wikiparse.ReadSeekCloser {
		return struct {
			io.ReadSeeker
			io.Closer
		}{
			ReadSeeker: func() io.ReadSeeker {
				return struct {
					io.Reader
					io.Seeker
				}{
					Reader: bufio.NewReader(ips.articleFile),
					Seeker: ips.articleFile,
				}
			}(),
			Closer: NoopCloser{},
		}
	}()

	return rc, nil
}

func pBZipCompress(f *os.File) error {
	// run the cli pbzip2 command to compress the file
	cmd := exec.Command("pbzip2", "-f", f.Name())
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Create a new index file that contains only the articles we want to parse. This is determined by the articles map.
func generateNewIndex(indexFile, newIndexFile string) (*os.File, error) {
	oldIndex, err := os.Open(indexFile)
	if err != nil {
		return nil, err
	}
	defer oldIndex.Close()

	newIndex, err := os.OpenFile(newIndexFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	z := bzip2.NewReader(oldIndex)
	indexParser := wikiparse.NewIndexReader(z)

	for {
		indexEntry, err := indexParser.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if _, ok := articles[ArticleID(indexEntry.PageOffset)]; ok {
			_, err := newIndex.WriteString(indexEntry.String() + "\n")
			if err != nil {
				return nil, err
			}
		}
	}

	// check newIndex isn't empty
	if _, err := newIndex.Stat(); err != nil {
		os.Remove(newIndex.Name())
		return nil, err
	}

	// go-wikiparse assumes the index file is bzip2 compressed, so we need to compress it.
	if err := pBZipCompress(newIndex); err != nil {
		return nil, err
	}

	newCompressedIndex, err := os.Open(newIndex.Name() + ".bz2")
	if err != nil {
		return nil, err
	}

	return newCompressedIndex, nil
}

// Parse the articles we want from the wiki dump using the new index file to identify the correct chunks to process.
func parse(indexF, articleF, outputF *os.File, workerNum int) error {
	indexedParseSrc := &IndexedParseSource{
		indexFile:   indexF,
		articleFile: articleF,
	}

	parser, err := wikiparse.NewIndexedParserFromSrc(
		indexedParseSrc,
		workerNum,
	)
	if err != nil {
		return err
	}

	for {
		// parser.Next() will return each page of the current chunk - currently, it doesn't provide the ability to seek a decompressed chunk
		// for the specific page offset(s) we want (this information is made available by the IndexReader) so we have to check the page ID against the articles map.
		// See https://pkg.go.dev/github.com/dustin/go-wikiparse#IndexReader
		page, err := parser.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fmt.Printf("%+v\n", page)

		// go-wikiparse is currently only returning the first page of each chunk - it seems to be unimplemented, but this would be a quick fix.
		if _, ok := articles[ArticleID(page.ID)]; ok {
			// write the page to a file
			for _, revision := range page.Revisions {
				// TODO: Investigate how to convert the revision text format to HTML.
				_, err := outputF.WriteString(revision.Text + "\n")
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func main() {
	if err := parseFlags(); err != nil {
		log.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	newIndex, err := generateNewIndex(indexWikiFile, filepath.Join(wd, NEW_INDEX_FILE))
	if err != nil {
		log.Fatal(err)
	}

	articleF, err := os.Open(articleWikiFile)
	if err != nil {
		log.Fatal(err)
	}

	outputF, err := os.OpenFile(filepath.Join(outputPath, "output.txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	if err := parse(newIndex, articleF, outputF, workerNum); err != nil {
		log.Fatal(err)
	}
}
