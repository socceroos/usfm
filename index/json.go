package index

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"

	"github.com/socceroos/usfm/parser"
)

// NewRenderer returns a JSON renderer
func NewRenderer(r io.Reader) Renderer {
	json := &JSON{}
	json.usfmParser = parser.NewParser(r)
	return json
}

// JSON renderer
type JSON struct {
	usfmParser *parser.Parser
}

// Render JSON
func (h *JSON) Render(w io.Writer, startKey int, startByte int64) (endKey int, err error) {
	content, err := h.usfmParser.Parse()
	if err != nil {
		return startKey, err
	}

	//converted, endKey := convertV2(content, startKey)
	converted, endKey := convertToIndex(content, startKey, startByte)

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent(" ", "  ")
	err = jsonEncoder.Encode(converted)
	//err = jsonEncoder.Encode(content)
	if err != nil {
		return startKey, err
	}

	return endKey, nil
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

type Translation struct {
	ShortCode     string
	Name          string
	Revision      string
	DatePublished string
}

type IndexItem struct {
	ID     int    `json:"id"`     //id
	RootID int    `json:"rootID"` //rootID
	OSIS   string `json:"osis"`   //osis
	Start  int64  `json:"start"`  //start
	End    int64  `json:"end"`    //end
	Type   string `json:"type"`   //type
}

var Index map[string]IndexItem

type IndexFormat struct {
	Translation Translation       `json:"translation"`
	Index       map[int]IndexItem `json:"index"`
}

func convertToIndex(in *parser.Node, key int, byteStart int64) (interface{}, int) {
	log.Print("\n\n\n\n\n\n\n\nCreating Carry JSON Index file...\n\n")
	out := IndexFormat{}
	out.Translation = Translation{ShortCode: "web", Name: "World English Bible", Revision: "1", DatePublished: "1997"}
	out.Index = map[int]IndexItem{}

	chapter := 0
	verse := 0
	ch := IndexItem{}
	book := IndexItem{}
	for _, row := range in.Children {
		fmt.Printf("row type %s %s\n", row.Type, row.Value)
		if row.Value == "\\c" {
			key++
			var err error
			chapter, err = strconv.Atoi(row.Children[0].Value)
			if err != nil {
				log.Printf("Error: %v", err)
				chapter++
			}
			ch = IndexItem{
				Type:   "chapter",
				ID:     key,
				RootID: key,
				OSIS:   book.OSIS + "." + row.Children[0].Value,
				Start:  int64(row.Position) + byteStart,
			}
			out.Index[key] = ch
			prevItem := out.Index[key-1]
			prevItem.End = ch.Start - 1
			out.Index[key-1] = prevItem
			verse = 0
		} else if row.Value == "\\h" {
			key++

			book = IndexItem{
				Type:   "book",
				ID:     key,
				RootID: key,
				OSIS:   in.Value,
				Start:  int64(row.Position) + byteStart,
			}
			out.Index[key] = book
		} else if row.Value == "\\p" || row.Value == "\\nb" || row.Value == "\\m" {
			for _, v := range row.Children {
				if v.Type == "text" {
				} else if v.Value == "\\v" {
					verse++
					isSubVerse := false

					for _, vC := range v.Children {
						if vC.Type == "versenumber" {
							var err error
							verse, err = strconv.Atoi(vC.Value)
							if err != nil {
								log.Printf("Error: %v", err)
							}
						} else if vC.Type == "subverse" {
							isSubVerse = true
						}
					}

					if !isSubVerse {
						key++
					}

					for _, vC := range v.Children {
						if vC.Type == "marker" {
							if vC.Value == "\\c" {
								break
							}
						}
					}
					log.Printf("Chapter %v Verse %v", chapter, verse)

					// Close the verse
					vC := IndexItem{
						Type:   "verse",
						ID:     key,
						RootID: key,
						OSIS:   ch.OSIS + "." + strconv.Itoa(verse),
						Start:  int64(v.Position) + byteStart,
					}
					out.Index[key] = vC
					prevItem := out.Index[key-1]
					prevItem.End = vC.Start - 1
					out.Index[key-1] = prevItem
				}
			}
		}
	}

	// Output the key we got up to.
	log.Printf("Last key was %v", key)
	log.Printf("Last byte was %v", out.Index[key].End)

	return out, key
}
