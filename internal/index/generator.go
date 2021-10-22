package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"embed"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/blevesearch/bleve"
)

//go:embed assets
var fs embed.FS

type Location struct {
	Name string
	Body string
	TZ   string
	Lang string
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	mapping := bleve.NewIndexMapping()
	idx, err := bleve.New(filepath.Join(wd, "assets/cities.idx"), mapping)
	if err != nil {
		if errors.Is(err, bleve.ErrorIndexPathExists) {
			log.Fatalf("index exists: %v", err)
		}

		log.Fatal(err)
	}
	dir, err := fs.ReadDir("assets")
	if err != nil {
		log.Fatal(err)
	}
	var exist bool
	for _, entry := range dir {
		if entry.Name() == "cities15000.txt" {
			exist = true
		}
	}

	if !exist {
		log.Fatal("cities15000.txt not found")
	}

	f, err := fs.Open(filepath.Join("assets", "cities15000.txt"))
	if err != nil {
		log.Fatalf("open file: %v", err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		bb := bytes.Split(scanner.Bytes(), []byte{'\t'})
		if err = idx.Index(string(bb[0]), Location{
			Name: string(bb[2]),
			Body: string(bb[3]),
			TZ:   string(bb[17]),
			Lang: string(bb[8]),
		}); err != nil {
			log.Printf("build index: %v", err)
		}
	}

	file, err := os.Create(filepath.Join(wd, "assets", "cities.idx.tar.gz"))
	if err != nil {
		log.Fatalf("can not create tarball file: %v", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	citiesDir, err := os.ReadDir(filepath.Join(wd, "assets/cities.idx"))
	if err != nil {
		log.Fatalf("unable read dir: %v", err)
	}

	for _, entry := range citiesDir {
		func() {
			file, err := os.Open(filepath.Join(wd, "assets/cities.idx", entry.Name()))
			if err != nil {
				log.Fatalf("can not open db file: %v", err)
			}

			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				log.Fatalf("unable check stat fike: %v", err)
			}

			header := &tar.Header{
				Name:    filepath.Join("cities.idx", entry.Name()),
				Size:    stat.Size(),
				Mode:    int64(stat.Mode()),
				ModTime: stat.ModTime(),
			}

			if err = tarWriter.WriteHeader(header); err != nil {
				log.Fatalf("unable write tar header: %v", err)
			}

			if _, err = io.Copy(tarWriter, file); err != nil {
				log.Fatalf("unable copy file to tarball: %v", err)
			}
		}()
	}
}
