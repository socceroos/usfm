package json

import (
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/socceroos/usfm/parser"
)

// NewRenderer returns a JSON renderer
func NewRenderer(o Options, r io.Reader) Renderer {
	json := &JSON{}
	json.usfmParser = parser.NewParser(r)
	return json
}

// JSON renderer
type JSON struct {
	usfmParser *parser.Parser
}

// Render JSON
func (h *JSON) Render(w io.Writer) error {
	content, err := h.usfmParser.Parse()
	if err != nil {
		return err
	}

	converted := convert(content)

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent(" ", "  ")
	err = jsonEncoder.Encode(converted)
	//err = jsonEncoder.Encode(content)
	if err != nil {
		return err
	}

	return nil
}

type Chapter struct {
	Verses map[int]*string
}

type Book struct {
	Title        string
	Abbreviation string
	Chapters     map[int]*Chapter
}

type Bible struct {
	Translation string
	Books       map[int]*Book
}

type OutputFormat struct {
	Books map[int]struct {
		Title    string
		Chapters map[int]struct {
			Verses map[int]struct {
				Text string
			}
		}
	}
}

func convert(in *parser.Content) interface{} {
	log.Print("\n\n\n\n\n\n\n\nConverting format to Carry JSON...\n\n")
	out := Bible{}
	out.Books = make(map[int]*Book)
	book := Book{}
	book.Title = in.Value
	book.Chapters = make(map[int]*Chapter)
	out.Books[0] = &book

	chapter := 0
	verse := 0
	ch := Chapter{}
	for _, row := range in.Children {
		log.Printf("GETTING NEW ROW: %v", row.Value)
		if row.Value == "\\c" {
			var err error
			chapter, err = strconv.Atoi(row.Children[0].Value)
			if err != nil {
				log.Printf("Error: %v", err)
				chapter++
			}
			log.Printf("Found Chapter %v", chapter)
			ch = Chapter{}
			ch.Verses = make(map[int]*string)
			out.Books[0].Chapters[chapter] = &ch
		} else if row.Value == "\\p" {
			log.Print("Found Paragraph.")
			// Should do something for paragraphs here...
		} else if row.Value == "\\v" {
			log.Print("Found Verse")
			verse++
			var verseText string
			for _, v := range row.Children {
				if v.Type == "marker" {
					if v.Value == "\\c" {
						break
					}
					for _, wl := range v.Children {
						if wl.Type == "text" {
							if !unicode.IsPunct([]rune(wl.Value)[0]) {
								verseText += " "
							}
							verseText += wl.Value
						}
					}
				} else if v.Type == "versenumber" {
					var err error
					verse, err = strconv.Atoi(v.Value)
					if err != nil {
						log.Printf("Error: %v", err)
					}
					log.Printf("Found Verse Number %v", verse)
				} else if v.Type == "text" {
					if !unicode.IsPunct([]rune(v.Value)[0]) {
						verseText += " "
					}
					verseText += v.Value
				}
			}
			log.Printf("Chapter: %v", chapter)
			verseText = strings.TrimSpace(verseText)
			out.Books[0].Chapters[chapter].Verses[verse] = &verseText
		}
	}

	return out
}
