package markdown

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"html/template"
	"io"
	"os"
	"strings"
	"time"
)

const DefaultDelimiter string = "---fm---"

type MDService struct {
	parser goldmark.Markdown
}
type MDAsync struct {
	fm  map[string]string
	md  []byte
	res chan Article
}

type Article struct {
	FM  FM
	Raw string
	MD  MD
	Err string
}

/*
Header used as metadata.
The header is composed of a multiline key-value
metadata encapsulated by a delimiter.

EX:

---fm---
author: victor hernandez
title: this is an article
desc: in this article we will talk about...
thumb: "link to article image"
---fm---
*/

// FM ...
type FM struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"desc"`
	Date        string `json:"date"`
	Category    string `json:"category"`
	Image       string `json:"image"`
}

// Holds a sanitized HTML version of your markdown rendered by BlackFriday.

// MD ...
type MD = template.HTML

func NewMarkdown() *MDService {
	return &MDService{
		parser: goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
				html.WithHardWraps(),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
				parser.WithHeadingAttribute(),
			),
			goldmark.WithExtensions(
				extension.TaskList,
				extension.DefinitionList,
				extension.GFM,
				extension.NewLinkify(
					extension.WithLinkifyAllowedProtocols([][]byte{
						[]byte("http:"),
						[]byte("https:"),
					}),
				),
			)),
	}
}

func (mds *MDService) ArticleAsync(md io.Reader) <-chan Article {
	w := MDAsync{
		fm:  make(map[string]string),
		res: make(chan Article),
	}

	go func(worker MDAsync) {
		defer close(worker.res)
		m, err := mds.Parse(md, worker.fm)
		if err != nil {
			worker.res <- Article{Err: err.Error()}
		}

		worker.res <- Article{
			FM: mds.FrontMatter(worker.fm),
			MD: mds.Markdown(m),
		}

	}(w)

	return w.res
}

// Article returns a fully rendered Article from a Markdown file and Front matter header.
func (mds *MDService) Article(md io.Reader) (Article, error) {
	fm := make(map[string]string)
	m, err := mds.Parse(md, fm)
	if err != nil {
		return Article{}, err
	}

	// Finalization of the Article's FrontMatter data.
	return Article{
		FM:  mds.FrontMatter(fm),
		MD:  mds.Markdown(m),
		Raw: string(m),
	}, nil
}

// Markdown Sanitizes a parsed markdown byte array and returns it as an MD (alias of template.HTML)
// Can be used in go's html templating engine.
func (mds *MDService) Markdown(md []byte) MD {
	// Full Post with Post data struct population
	var buf bytes.Buffer
	if err := mds.parser.Convert(md, &buf); err != nil {
		panic(err)
	}

	return template.HTML(buf.Bytes())
}

// FrontMatter Sets Article with FrontMatterData.
func (mds *MDService) FrontMatter(fm map[string]string) FM {
	pd := FM{}

	// Asserting data coming from front-matter
	if author, ok := fm["author"]; ok {
		pd.Author = author
	}
	if title, ok := fm["title"]; ok {
		pd.Title = title
	}
	if desc, ok := fm["desc"]; ok {
		pd.Description = desc
	}
	if thumb, ok := fm["thumb"]; ok {
		if _, err := os.Stat(thumb); os.IsExist(err) {
			pd.Image = thumb
		}
	}

	pd.Date = time.Now().Format("Mon, Jan 2 2006")

	return pd
}

// ParseDelimited parses the markdown file by feeding it the path of the markdown file.
// - If the file has a front matter header pass a map of 'map[string]string', else pass nil.
// - Pass a delimiter for your front matter header if any.
// - Returns un-sanitized byte array of markdown.
func (mds *MDService) ParseDelimited(md io.Reader, delimiter string, fm map[string]string) ([]byte, error) {
	return mds.parse(md, delimiter, fm)
}

// Parse parses the markdown file by feeding it the path of the markdown file.
// - If the file has a front matter header pass a map of 'map[string]string', else pass nil.
// - Uses default delimiter for front matter header if any.
// - Returns un-sanitized byte array of markdown.
func (mds *MDService) Parse(md io.Reader, fm map[string]string) ([]byte, error) {
	return mds.parse(md, DefaultDelimiter, fm)
}

func (_ *MDService) parse(md io.Reader, delimiter string, fm map[string]string) ([]byte, error) {
	// Ready to iterate with file scanner
	scanner := bufio.NewScanner(md)

	// markdown will be written here.
	buff := bytes.Buffer{}
	isFMBlockOpen := false
	for scanner.Scan() {
		text := scanner.Text()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		switch text {
		case delimiter: // Toggle front-matter parsing.
			isFMBlockOpen = !isFMBlockOpen

		default:
			if !isFMBlockOpen { // When not parsing front-matter
				buff.WriteString(text + "\n") // We parse the markdown.
				continue
			} else { // Here we parse the front-matter if the block is open.
				if fm != nil {
					splitValues := strings.Split(text, ":")
					if len(splitValues) < 2 || len(splitValues) > 2 {
						return nil, errors.New("check front-matter header for errors")
					}
					fm[splitValues[0]] = strings.TrimSpace(splitValues[1])
				}
			}
		}
	}

	return buff.Bytes(), nil
}
