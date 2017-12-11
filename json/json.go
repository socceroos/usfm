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

	converted := convertV2(content)

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
					if v.Value == "\\wj" {
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
					if v.Value == "\\wj" {
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
	Type    string
	Key     int
	RootMap []int
	BCV     string
	Text    string
}

type CarryFormat struct {
	Translation Translation
	BibleStream []Item
}

func convertV2(in *parser.Content) interface{} {
	log.Print("\n\n\n\n\n\n\n\nConverting format to Carry JSON v2...\n\n")
	out := CarryFormat{}
	out.Translation = Translation{ShortCode: "web", Name: "World English Bible", Revision: "1", DatePublished: "1997"}

	chapter := 0
	verse := 0
	ch := Item{}
	book := Item{}
	for i, row := range in.Children {
		if row.Value == "\\c" {
			var err error
			chapter, err = strconv.Atoi(row.Children[0].Value)
			if err != nil {
				log.Printf("Error: %v", err)
				chapter++
			}
			chText := `<span class="chapter">` + row.Children[0].Value + `</span>`
			ch = Item{Type: "chapter", Key: i, BCV: book.BCV + "." + row.Children[0].Value, Text: chText}
			out.BibleStream = append(out.BibleStream, ch)
			verse = 0
		} else if row.Value == "\\h" {
			cHead := `<span class="book">`
			for _, v := range row.Children {
				if v.Type == "heading" {
					cHead += v.Value
				}
			}
			cHead += "</span>"
			book = Item{Type: "book", Key: i, BCV: in.Value, Text: cHead}
			out.BibleStream = append(out.BibleStream, book)
		} else if row.Value == "\\p" {
			pText := "<p>"
			for _, v := range row.Children {
				/*jsonEncoder := json.NewEncoder(os.Stdout)
				jsonEncoder.SetIndent(" ", "  ")
				err := jsonEncoder.Encode(v)
				if err != nil {
					log.Printf("%v", err)
				}*/
				if v.Type == "text" {
					if !unicode.IsPunct([]rune(v.Value)[0]) || []rune(v.Value)[0] == 0x201C {
						pText += " "
					}

					pText += v.Value
				} else if v.Value == "\\v" {
					verse++
					var verseText string
					verseText += "<span class='bible-verse-number r" + strconv.Itoa(i) + " v" + strconv.Itoa(verse) + "'>" + strconv.Itoa(verse) + "</span><span class='bible-verse r" + strconv.Itoa(i) + " v" + strconv.Itoa(verse) + "'>"
					for _, vC := range v.Children {
						if vC.Type == "marker" {
							if vC.Value == "\\c" {
								break
							}
							if vC.Value == "\\wj" {
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
						} else if vC.Type == "versenumber" {
							var err error
							verse, err = strconv.Atoi(vC.Value)
							if err != nil {
								log.Printf("Error: %v", err)
							}
						} else if vC.Type == "text" {
							if !unicode.IsPunct([]rune(vC.Value)[0]) {
								verseText += " "
							}
							verseText += vC.Value
						}
					}
					log.Printf("Chapter %v Verse %v", chapter, verse)
					verseText += "</span>"
					pText += strings.TrimSpace(verseText)
					//vC := Item{Type: "verse", Key: i, BCV: ch.BCV + "." + strconv.Itoa(verse), Text: verseText}
					//out.BibleStream = append(out.BibleStream, vC)
				}
			}
			if len(row.Children) > 0 {
				pText += "</p>"
			}
			p := Item{Type: "paragraph", Key: i, Text: pText}
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
