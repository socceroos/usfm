package json

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/socceroos/usfm/parser"
)

// Index Types
const (
	Chapter = "chapter"
	Book    = "book"
	Verse   = "verse"
)

// NewJSON return an empty Json Index file
func NewJSON() *JSON {
	json := &JSON{Index: []indexItem{}}
	json.Index = append(json.Index, indexItem{})
	return json
}

// JSON renderer
type JSON struct {
	Translation translation `json:"translation"`
	Index       []indexItem `json:"index"`
}

type translation struct {
	ShortCode     string
	Name          string
	Revision      string
	DatePublished string
}

type indexItem struct {
	ID    int    `json:"rootID"` //id
	OSIS  string `json:"osis"`   //osis
	Start int64  `json:"start"`  //start
	End   int64  `json:"end"`    //end
	Type  string `json:"type"`   //type
}

// ReadTranslation read translation info from path
func (j *JSON) ReadTranslation(path string) {
	jsonFile, _ := os.Open(path)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &j.Translation)
	defer jsonFile.Close()
}

// WriteFile write json data to a specified path
func (j *JSON) WriteFile(path string) {
	jsonData, _ := json.MarshalIndent(j, " ", "  ")
	ioutil.WriteFile(path, jsonData, 0644)
}

// AppendUsfmIndex parse the usfm content and append to index json format
func (j *JSON) AppendUsfmIndex(path string, startKey int, startByte int64) (endKey int, fileSize int64) {
	// Init input stream
	in, _ := os.Open(path)

	defer in.Close()

	// Init parser from reader stream
	parser := parser.NewParser(in)
	content, err := parser.Parse()
	if err != nil {
		log.Printf("Error: %s", err)
	}

	j.Index, endKey = j.mapContent(content, startKey, startByte)

	fi, _ := in.Stat()
	fileSize = fi.Size()

	// Update last item ending byte index
	lastItem := j.Index[endKey]
	lastItem.End = fileSize + startByte
	j.Index[endKey] = lastItem

	return endKey, fileSize + startByte
}

func (j *JSON) mapContent(in *parser.Content, key int, byteStart int64) ([]indexItem, int) {
	out := j.Index

	verse := 0
	ch := indexItem{}
	book := indexItem{}
	for _, row := range in.Children {
		// fmt.Printf("row type %s %s\n", row.Type, row.Value)

		if row.Value == "\\c" {
			// Chapter
			key++
			var err error
			if err != nil {
				log.Printf("Error: %v", err)
			}
			ch = indexItem{
				Type:  Chapter,
				ID:    key,
				OSIS:  book.OSIS + "." + row.Children[0].Value,
				Start: int64(row.Position) + byteStart,
			}
			out = append(out, ch)
			// out[key] = ch
			prevItem := out[key-1]
			prevItem.End = ch.Start - 1
			out[key-1] = prevItem
			verse = 0
		} else if row.Value == "\\h" {
			// Header
			key++

			book = indexItem{
				Type:  Book,
				ID:    key,
				OSIS:  in.Value,
				Start: int64(row.Position) + byteStart,
			}
			out = append(out, book)
			// out[key] = book
		} else if row.Value == "\\v" {
			isSubVerse := false

			for _, vC := range row.Children {
				if vC.Type == "versenumber" {
					verse++
					// var err error
					// verse, err = strconv.Atoi(vC.Value)
					// if err != nil {
					// 	log.Println(row)
					// 	log.Printf("%s Error: %v", vC.Value, err)
					// }
				} else if vC.Type == "subverse" {
					isSubVerse = true
				}
			}

			if !isSubVerse {
				key++
			}

			for _, vC := range row.Children {
				if vC.Type == "marker" {
					if vC.Value == "\\c" {
						break
					}
				}
			}
			// log.Printf("Chapter %v Verse %v", chapter, verse)

			// Close the verse
			vC := indexItem{
				Type:  Verse,
				ID:    key,
				OSIS:  ch.OSIS + "." + strconv.Itoa(verse),
				Start: int64(row.Position) + byteStart,
			}

			out = append(out, vC)
			// out[key] = vC
			prevItem := out[key-1]
			prevItem.End = vC.Start - 1
			out[key-1] = prevItem
		}
	}
	return out, key
}
