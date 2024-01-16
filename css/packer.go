package main

import (
	"bytes"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	dir        = "."                              // Replace with your directory path
	outputFile = "/authentity/res/packed.min.css" // Name of the output file
)

func main() {

	// Create a minifier instance
	m := minify.New()
	m.AddFunc("text/css", css.Minify)

	var buffer bytes.Buffer

	// Read all files in the specified directory
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".css") {
			// Read the CSS file
			data, e := os.ReadFile(path)
			if e != nil {
				return e
			}

			// Minify the CSS and write to the buffer
			minifiedCSS, e := m.String("text/css", string(data))
			if e != nil {
				return e
			}

			buffer.WriteString(minifiedCSS)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	// Write the combined minified CSS to an output file
	err = os.WriteFile(outputFile, buffer.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("CSS packing complete:", outputFile)
}
