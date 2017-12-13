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
func (h *JSON) Render(w io.Writer, startKey int) error {
	content, err := h.usfmParser.Parse()
	if err != nil {
		return err
	}

	converted := convertV2(content, startKey)

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

func convertV1(in *parser.Content) interface{} {
	log.Print("\n\n\n\n\n\n\n\nConverting format to Carry JSON v1...\n\n")
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
					if v.Value == "\\wj" || v.Value == `\wj` {
						verseText += `<span class="jesus-words">`
					}
					for _, wl := range v.Children {
						if wl.Type == "text" {
							if !unicode.IsPunct([]rune(wl.Value)[0]) {
								verseText += " "
							}
							verseText += wl.Value
						}
					}
					if v.Value == "\\wj" || v.Value == `\wj` {
						verseText += `</span>`
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

type Translation struct {
	ShortCode     string
	Name          string
	Revision      string
	DatePublished string
}

type Item struct {
	Type     string
	Key      int
	RootMap  []int
	BCV      string
	Text     string
	Children []Item
}

type CarryFormat struct {
	Translation Translation
	BibleStream []Item
}

func convertV2(in *parser.Content, key int) interface{} {
	log.Print("\n\n\n\n\n\n\n\nConverting format to Carry JSON v2...\n\n")
	out := CarryFormat{}
	out.Translation = Translation{ShortCode: "web", Name: "World English Bible", Revision: "1", DatePublished: "1997"}

	chapter := 0
	verse := 0
	ch := Item{}
	book := Item{}
	bookName := ""
	for _, row := range in.Children {
		if row.Value == "\\c" {
			key++
			var err error
			chapter, err = strconv.Atoi(row.Children[0].Value)
			if err != nil {
				log.Printf("Error: %v", err)
				chapter++
			}
			chText := `<span class="bible-chapter">` + bookName + " " + row.Children[0].Value + `</span>`
			ch = Item{Type: "chapter", Key: key, BCV: book.BCV + "." + row.Children[0].Value, Text: chText}
			ch.RootMap = append(ch.RootMap, key)
			out.BibleStream = append(out.BibleStream, ch)
			verse = 0
		} else if row.Value == "\\h" {
			key++
			cHead := `<span class="bible-book">`
			for _, v := range row.Children {
				if v.Type == "heading" {
					cHead += v.Value
					bookName = v.Value
				}
			}
			cHead += "</span>"
			book = Item{Type: "book", Key: key, BCV: in.Value, Text: cHead}
			book.RootMap = append(book.RootMap, key)
			out.BibleStream = append(out.BibleStream, book)
		} else if row.Value == "\\p" {
			hasQ1Marker := false
			q1Count := 0
			hasQ2Marker := false
			q2Count := 0
			pText := "<p>"
			p := Item{Type: "paragraph", Key: 0, Text: "", Children: []Item{}}
			for _, v := range row.Children {
				if v.Type == "text" {
					if !unicode.IsPunct([]rune(v.Value)[0]) || []rune(v.Value)[0] == 0x201C {
						pText += " "
					}

					pText += v.Value
				} else if v.Value == "\\q1" {
					hasQ1Marker = true
					q1Count++
				} else if v.Value == "\\q2" {
					hasQ2Marker = true
					q2Count++
				} else if v.Value == "\\v" {
					verse++
					isSubVerse := false
					var verseText string

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
						verseText += "<span class='bible-verse-number r" + strconv.Itoa(key) + " v" + strconv.Itoa(verse) + "'>" + strconv.Itoa(verse) + "</span>"
					}
					verseText += "<span class='bible-verse r" + strconv.Itoa(key) + " v" + strconv.Itoa(verse) + "'>"

					// If we have a poetic marker then add the span
					if hasQ1Marker && q1Count == 1 && q2Count == 0 {
						verseText += "<span class='poetic-1'>"
					} else if hasQ1Marker && q1Count > 1 {
						verseText += "</span><span class='poetic-1'>"
					}
					if hasQ2Marker && q2Count == 1 && q1Count == 0 {
						verseText += "<span class='poetic-2'>"
					} else if hasQ2Marker && (q2Count > 1 || q1Count >= 1) {
						verseText += "</span><span class='poetic-2'>"
					}
					for _, vC := range v.Children {
						if vC.Type == "marker" {
							if vC.Value == "\\c" {
								break
							} else if vC.Value == "\\q1" {
								hasQ1Marker = true
								q1Count++

								// If we have a poetic marker then add the span
								if hasQ1Marker && q1Count == 1 && q2Count == 0 {
									verseText += "<span class='poetic-1'>"
								} else if hasQ1Marker && q1Count > 1 {
									verseText += "</span><span class='poetic-1'>"
								}
							} else if vC.Value == "\\q2" {
								hasQ2Marker = true
								q2Count++

								// If we have a poetic marker then add the span
								if hasQ2Marker && q2Count == 1 && q1Count == 0 {
									verseText += "<span class='poetic-2'>"
								} else if hasQ2Marker && (q2Count > 1 || q1Count >= 1) {
									verseText += "</span><span class='poetic-2'>"
								}
							} else if vC.Value == "\\wj" {
								verseText += `<span class='jesus-words'>`
							}
							for _, wl := range vC.Children {
								if wl.Type == "text" {
									if !unicode.IsPunct([]rune(wl.Value)[0]) {
										verseText += " "
									}
									verseText += wl.Value
								}
							}
							if vC.Value == "\\wj" {
								verseText += `</span>`
							}
						} else if vC.Type == "text" {
							if !unicode.IsPunct([]rune(vC.Value)[0]) {
								verseText += " "
							}
							verseText += vC.Value
						}
					}
					log.Printf("Chapter %v Verse %v", chapter, verse)
					// Close our poetic lines
					if hasQ1Marker || hasQ2Marker {
						verseText += "</span>"
					}
					// Close the verse
					verseText += "</span>"
					pText += strings.TrimSpace(verseText)
					vC := Item{Type: "verse", Key: key, BCV: ch.BCV + "." + strconv.Itoa(verse), Text: verseText}
					p.Children = append(p.Children, vC)
				}
			}
			if len(row.Children) > 0 {
				pText += "</p>"
			}
			//p := Item{Type: "paragraph", Key: key, Text: pText}
			// Find the range of RootMap keys we're supporting in this.
			var pKeys []int
			for _, v := range p.Children {
				if v.Key > 0 {
					pKeys = append(pKeys, v.Key)
				}
			}
			p.RootMap = pKeys
			out.BibleStream = append(out.BibleStream, p)
		} /*else if row.Value == "\\v" {
			verse++
			var verseText string
			for _, v := range row.Children {
				if v.Type == "marker" {
					if v.Value == "\\c" {
						break
					}
					if v.Value == "\\wj" {
						verseText += `<span class='jesus-words'>`
					}
					for _, wl := range v.Children {
						if wl.Type == "text" {
							if !unicode.IsPunct([]rune(wl.Value)[0]) {
								verseText += " "
							}
							verseText += wl.Value
						}
					}
					if v.Value == "\\wj" {
						verseText += `</span>`
					}
				} else if v.Type == "versenumber" {
					var err error
					verse, err = strconv.Atoi(v.Value)
					if err != nil {
						log.Printf("Error: %v", err)
					}
				} else if v.Type == "text" {
					if !unicode.IsPunct([]rune(v.Value)[0]) {
						verseText += " "
					}
					verseText += v.Value
				}
			}
			log.Printf("Chapter %v Verse %v", chapter, verse)
			verseText = strings.TrimSpace(verseText)
			v := Item{Type: "verse", Key: i, BCV: ch.BCV + "." + strconv.Itoa(verse), Text: verseText}
			out.BibleStream = append(out.BibleStream, v)
		}*/
	}

	return out
}
