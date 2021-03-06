package parser

import (
	"fmt"
	"io"
	"log"
)

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok Token  // last read token
		lit string // last read literal
		n   int    // buffer size (max=1)
		pos int    // scanner position (byte offset)
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// Parse parses a USFM formatted book content
func (p *Parser) Parse() (*Content, error) {
	log.Printf("Scanning for book...")
	book := &Content{}
	book.Type = "book"
	markerV := &Content{}
	for {
		// Read a field.
		tok, lit, pos := p.scanIgnoreWhitespace()
		if tok == MarkerID {
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			tok, lit, pos = p.scanIgnoreWhitespace()
			if tok == Text && len([]rune(lit)) == 3 {
				child := &Content{}
				child.Type = "bookcode"
				child.Value = lit
				child.Position = pos
				book.Value = lit
				book.Position = pos
				marker.Children = append(marker.Children, child)
				for {
					tok, lit, pos = p.scanIgnoreWhitespace()
					if !(tok == Text || tok == Number) {
						p.unscan()
						break
					} else {
						child := &Content{}
						child.Type = "text"
						child.Value = lit
						child.Position = pos
						marker.Children = append(marker.Children, child)
					}
				}
			} else {
				return nil, fmt.Errorf("found %q, expected book code", lit)
			}
		} else if tok == MarkerIde {
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			for {
				tok, lit, pos = p.scanIgnoreWhitespace()
				if !(tok == Text || tok == Number) {
					p.unscan()
					break
				} else {
					child := &Content{}
					child.Type = "text"
					child.Value = lit
					child.Position = pos
					marker.Children = append(marker.Children, child)
				}
			}
		} else if tok == MarkerC {
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			tok, lit, pos = p.scanIgnoreWhitespace()
			if tok == Number {
				child := &Content{}
				child.Type = "chapternumber"
				child.Value = lit
				child.Position = pos
				marker.Children = append(marker.Children, child)
			} else {
				return nil, fmt.Errorf("found %q, expected chapter number", lit)
			}
		} else if tok == MarkerH {
			log.Print("Found Heading marker.")
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			for {
				tok, lit, pos = p.scanIgnoreWhitespace()
				if !(tok == Text || tok == Number) {
					p.unscan()
					break
				} else {
					child := &Content{}
					child.Type = "heading"
					child.Value = lit
					child.Position = pos
					marker.Children = append(marker.Children, child)
				}
			}
		} else if tok == MarkerD {
			log.Print("Found Descriptive Title marker.")
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			for {
				tok, lit, pos = p.scanIgnoreWhitespace()
				if !(tok == Text) {
					p.unscan()
					break
				} else {
					child := &Content{}
					child.Type = "description"
					child.Value = lit
					child.Position = pos
					marker.Children = append(marker.Children, child)
				}
			}
		} else if tok == MarkerP {
			haveQ1Carryover := false
			q1Carryover := &Content{}
			log.Print("\n\n\n\nFound Paragraph marker.")
			markerP := &Content{}
			markerP.Type = "marker"
			markerP.Value = lit
			markerP.Position = pos
			book.Children = append(book.Children, markerP)
			for {
				tok, lit, pos = p.scanIgnoreWhitespace()
				/*if tok == MarkerP || tok == MarkerC {*/
				if tok == EOF || tok == MarkerC || tok == MarkerP || tok == MarkerS {
					p.unscan()
					break
				} else if tok == MarkerQ1 {
					log.Print("Found Q1 marker.")
					child := &Content{}
					child.Type = "marker"
					child.Value = lit
					child.Position = pos
					tok, lit, pos = p.scanIgnoreWhitespace()
					if tok == MarkerV {
						q1Carryover = child
						haveQ1Carryover = true
						p.unscan()
						continue
					}
				} else if tok == MarkerD {
					log.Print("Found Descriptive Title marker.")
					marker := &Content{}
					marker.Type = "marker"
					marker.Value = lit
					marker.Position = pos
					markerP.Children = append(markerP.Children, marker)
					for {
						tok, lit, pos = p.scanIgnoreWhitespace()
						if !(tok == Text) {
							p.unscan()
							break
						} else {
							child := &Content{}
							child.Type = "description"
							child.Value = lit
							child.Position = pos
							marker.Children = append(marker.Children, child)
						}
					}
				} else if tok == MarkerV {
					log.Print("\n\nFound Verse marker.")
					markerV = &Content{}
					markerV.Type = "marker"
					markerV.Value = lit
					markerV.Position = pos
					markerP.Children = append(markerP.Children, markerV)
					tok, lit, pos = p.scanIgnoreWhitespace()
					if tok == Number {
						child := &Content{}
						child.Type = "versenumber"
						child.Value = lit
						child.Position = pos
						markerV.Children = append(markerV.Children, child)
						log.Printf("Verse Number is %v", child.Value)

						// Add the q1Carryover if there is one
						if haveQ1Carryover {
							markerV.Children = append(markerV.Children, q1Carryover)
							haveQ1Carryover = false
							q1Carryover = &Content{}
						}
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							//log.Printf("Token: %v Lit: %v", tok, lit)
							if tok == 0x0085 || tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerP || tok == MarkerS || tok == MarkerD {
								p.unscan()
								break
							} else if tok == MarkerQ1 {
								log.Print("Found Q1 marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == MarkerV {
									q1Carryover = childA
									haveQ1Carryover = true
									p.unscan()
									break
								} else {
									p.unscan()
									markerV.Children = append(markerV.Children, childA)
								}
							} else if tok == MarkerQ2 {
								log.Print("Found Q2 marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
							} else if tok == MarkerWJ {
								log.Print("Found Jesus' Words markerV.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerWJ {
										log.Print("Found Jesus' Words end marker.\n\n")
										//p.unscan()
										break
									} else {
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childA.Children = append(childA.Children, childT)
									}
								}
							} else if tok == MarkerAdd {
								log.Print("Found Add marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerAdd {
										log.Print("Found Add end marker.\n\n")
										//p.unscan()
										break
									} else {
										log.Print("Found Add subject text.")
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childA.Children = append(childA.Children, childT)
									}
								}
							} else if tok == MarkerW {
								log.Print("Found Wordlist marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerW {
										log.Print("Found Wordlist end marker.\n\n")
										break
									} else if tok == Citation {
										log.Print("Found Citation metadata.")
										childC := &Content{}
										childC.Type = "citation"
										childC.Value = lit
										childC.Position = pos
										childW.Children = append(childW.Children, childC)
									} else {
										log.Print("Found Citation subject text.")
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childW.Children = append(childW.Children, childT)
									}

								}
							} else if tok == MarkerSP {
								log.Print("Found Speaker Identification marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok != Text {
										p.unscan()
										break
									} else {
										childT := &Content{}
										childT.Type = "speaker"
										childT.Value = lit
										childT.Position = pos
										childW.Children = append(childW.Children, childT)
									}
								}
							} else if tok == MarkerQS {
								log.Print("Found Selah marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerQS {
										break
									} else {
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childW.Children = append(childW.Children, childT)
									}
								}
							} else if tok == MarkerF {
								log.Print("Found Footnote marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerF {
										log.Print("Found Footnote end marker.\n\n")
										//p.unscan()
										break
									}
								}
							} else if tok == MarkerX {
								log.Print("Found Cross-Reference marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerX {
										log.Print("Found Cross-Reference end marker.\n\n")
										//p.unscan()
										break
									}
								}
							} else if tok == MarkerB {
							} else {
								child := &Content{}
								child.Type = "text"
								child.Value = lit
								child.Position = pos
								markerV.Children = append(markerV.Children, child)
							}
						}

					} else {
						return nil, fmt.Errorf("found %q, expected verse number", lit)
					}
				} else if tok == Text {
					// OK we've found a paragraph that
					// continues a previous verse
					log.Print("\n\n\nWe're in a Paragraph with Text now:\n\n")

					p.unscan()
					var verseNum *Content
					for _, c := range markerV.Children {
						if c.Type == "versenumber" {
							verseNum = c
							break
						}
					}
					newVerseNum := &Content{Type: "versenumber", Value: verseNum.Value, Children: verseNum.Children}
					markerPV := &Content{}
					markerPV.Type = "marker"
					markerPV.Value = "\\v"
					markerPV.Children = append(markerPV.Children, newVerseNum)
					// Add a new "sub-verse" marker
					markerSV := &Content{Type: "subverse", Value: "Sub-verse paragraph", Children: nil}
					markerPV.Children = append(markerPV.Children, markerSV)
					for {
						tok, lit, pos = p.scanIgnoreWhitespace()
						//log.Printf("Token: %v Lit: %v", tok, lit)
						if tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerP || tok == MarkerS {
							log.Printf("We're breaking because we hit %v:%v", tok, lit)
							p.unscan()
							break
						} else if tok == MarkerB {
						} else if tok == MarkerD {
							p.unscan()
							break
							log.Print("Found Descriptive Title marker.")
							marker := &Content{}
							marker.Type = "marker"
							marker.Value = lit
							marker.Position = pos
							markerPV.Children = append(markerPV.Children, marker)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if !(tok == Text) {
									p.unscan()
									break
								} else {
									child := &Content{}
									child.Type = "description"
									child.Value = lit
									child.Position = pos
									marker.Children = append(marker.Children, child)
								}
							}
						} else if tok == MarkerSP {
							log.Print("Found Speaker Identification marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok != Text {
									p.unscan()
									break
								} else {
									childT := &Content{}
									childT.Type = "speaker"
									childT.Value = lit
									childT.Position = pos
									childA.Children = append(childA.Children, childT)
								}
							}
						} else if tok == MarkerQ1 {
							log.Print("Found Q1 marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
						} else if tok == MarkerQ2 {
							log.Print("Found Q2 marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
						} else if tok == MarkerWJ {
							log.Print("Found Jesus' Words marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerWJ {
									log.Print("Found Jesus' Words end markerPV.\n\n")
									//p.unscan()
									break
								} else {
									childT := &Content{}
									childT.Type = "text"
									childT.Value = lit
									childT.Position = pos
									childA.Children = append(childA.Children, childT)
								}
							}
						} else if tok == MarkerAdd {
							log.Print("Found Add marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerAdd {
									log.Print("Found Add end marker.\n\n")
									//p.unscan()
									break
								} else {
									log.Print("Found Add subject text.")
									childT := &Content{}
									childT.Type = "text"
									childT.Value = lit
									childT.Position = pos
									childA.Children = append(childA.Children, childT)
								}
							}
						} else if tok == MarkerW {
							log.Print("Found Wordlist marker.")
							childW := &Content{}
							childW.Type = "marker"
							childW.Value = lit
							childW.Position = pos
							markerPV.Children = append(markerPV.Children, childW)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerW {
									log.Print("Found Wordlist end marker.\n\n")
									break
								} else if tok == Citation {
									log.Print("Found Citation metadata.")
									childC := &Content{}
									childC.Type = "citation"
									childC.Value = lit
									childC.Position = pos
									childW.Children = append(childW.Children, childC)
								} else {
									log.Print("Found Citation subject text.")
									childT := &Content{}
									childT.Type = "text"
									childT.Value = lit
									childT.Position = pos
									childW.Children = append(childW.Children, childT)
								}

							}
						} else if tok == MarkerF {
							log.Print("Found Footnote marker.")
							childW := &Content{}
							childW.Type = "marker"
							childW.Value = lit
							childW.Position = pos
							markerPV.Children = append(markerPV.Children, childW)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerF {
									log.Print("Found Footnote end marker.\n\n")
									//p.unscan()
									break
								}
							}
						} else if tok == MarkerX {
							log.Print("Found Cross-Reference marker.")
							childW := &Content{}
							childW.Type = "marker"
							childW.Value = lit
							childW.Position = pos
							markerPV.Children = append(markerPV.Children, childW)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerX {
									log.Print("Found Cross-Reference end marker.\n\n")
									//p.unscan()
									break
								}
							}
						} else {
							childT := &Content{}
							childT.Type = "text"
							childT.Value = lit
							childT.Position = pos
							markerPV.Children = append(markerPV.Children, childT)
						}
					}
					markerP.Children = append(markerP.Children, markerPV)
					//break
				}
			}
		} else if tok == MarkerV || tok == MarkerQ1 {
			haveQ1Carryover := false
			q1Carryover := &Content{}
			log.Print("\n\n\n\nCreating fake Paragraph marker.")
			markerP := &Content{}
			markerP.Type = "marker"
			markerP.Value = "\\p"
			book.Children = append(book.Children, markerP)
			p.unscan()
			for {
				tok, lit, pos = p.scanIgnoreWhitespace()
				/*if tok == MarkerP || tok == MarkerC {*/
				if tok == EOF || tok == MarkerC || tok == MarkerP || tok == MarkerS {
					p.unscan()
					break
				} else if tok == MarkerD {
					log.Print("Found Descriptive Title marker.")
					marker := &Content{}
					marker.Type = "marker"
					marker.Value = lit
					marker.Position = pos
					markerP.Children = append(markerP.Children, marker)
					for {
						tok, lit, pos = p.scanIgnoreWhitespace()
						if !(tok == Text) {
							p.unscan()
							break
						} else {
							child := &Content{}
							child.Type = "description"
							child.Value = lit
							child.Position = pos
							marker.Children = append(marker.Children, child)
						}
					}
				} else if tok == MarkerSP {
					log.Print("Found Speaker Identification marker.")
					child := &Content{}
					child.Type = "marker"
					child.Value = lit
					child.Position = pos
					markerP.Children = append(markerP.Children, child)
					for {
						tok, lit, pos = p.scanIgnoreWhitespace()
						if tok != Text {
							p.unscan()
							break
						} else {
							childT := &Content{}
							childT.Type = "speaker"
							childT.Value = lit
							childT.Position = pos
							child.Children = append(child.Children, childT)
						}
					}
				} else if tok == MarkerQS {
					log.Print("Found Selah marker.")
					childW := &Content{}
					childW.Type = "marker"
					childW.Value = lit
					childW.Position = pos
					markerP.Children = append(markerP.Children, childW)
					for {
						tok, lit, pos = p.scanIgnoreWhitespace()
						if tok == EndMarkerQS {
							break
						} else {
							childT := &Content{}
							childT.Type = "text"
							childT.Value = lit
							childT.Position = pos
							childW.Children = append(childW.Children, childT)
						}
					}
				} else if tok == MarkerQ1 {
					log.Print("Found Q1 marker.")
					child := &Content{}
					child.Type = "marker"
					child.Value = lit
					child.Position = pos
					tok, lit, pos = p.scanIgnoreWhitespace()
					if tok == MarkerV {
						q1Carryover = child
						haveQ1Carryover = true
						p.unscan()
						continue
					}
				} else if tok == MarkerV {
					log.Print("\n\nFound Verse marker.")
					markerV = &Content{}
					markerV.Type = "marker"
					markerV.Value = lit
					markerV.Position = pos
					markerP.Children = append(markerP.Children, markerV)
					tok, lit, pos = p.scanIgnoreWhitespace()
					if tok == Number {
						child := &Content{}
						child.Type = "versenumber"
						child.Value = lit
						child.Position = pos
						markerV.Children = append(markerV.Children, child)
						log.Printf("Verse Number is %v", child.Value)

						// Add the q1Carryover if there is one
						if haveQ1Carryover {
							markerV.Children = append(markerV.Children, q1Carryover)
							haveQ1Carryover = false
							q1Carryover = &Content{}
						}
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							//log.Printf("Token: %v Lit: %v", tok, lit)
							if tok == 0x0085 || tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerP || tok == MarkerS || tok == MarkerD {
								p.unscan()
								break
							} else if tok == MarkerQS {
								log.Print("Found Selah marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerQS {
										break
									} else {
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childW.Children = append(childW.Children, childT)
									}
								}
							} else if tok == MarkerSP {
								log.Print("Found Speaker Identification marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok != Text {
										p.unscan()
										break
									} else {
										childT := &Content{}
										childT.Type = "speaker"
										childT.Value = lit
										childT.Position = pos
										childA.Children = append(childA.Children, childT)
									}
								}
							} else if tok == MarkerQ1 {
								log.Print("Found Q1 marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == MarkerV {
									q1Carryover = childA
									haveQ1Carryover = true
									p.unscan()
									break
								} else {
									p.unscan()
									markerV.Children = append(markerV.Children, childA)
								}
							} else if tok == MarkerQ2 {
								log.Print("Found Q2 marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
							} else if tok == MarkerWJ {
								log.Print("Found Jesus' Words markerV.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerWJ {
										log.Print("Found Jesus' Words end marker.\n\n")
										//p.unscan()
										break
									} else {
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childA.Children = append(childA.Children, childT)
									}
								}
							} else if tok == MarkerAdd {
								log.Print("Found Add marker.")
								childA := &Content{}
								childA.Type = "marker"
								childA.Value = lit
								childA.Position = pos
								markerV.Children = append(markerV.Children, childA)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerAdd {
										log.Print("Found Add end marker.\n\n")
										//p.unscan()
										break
									} else {
										log.Print("Found Add subject text.")
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childA.Children = append(childA.Children, childT)
									}
								}
							} else if tok == MarkerW {
								log.Print("Found Wordlist marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerW {
										log.Print("Found Wordlist end marker.\n\n")
										break
									} else if tok == Citation {
										log.Print("Found Citation metadata.")
										childC := &Content{}
										childC.Type = "citation"
										childC.Value = lit
										childC.Position = pos
										childW.Children = append(childW.Children, childC)
									} else {
										log.Print("Found Citation subject text.")
										childT := &Content{}
										childT.Type = "text"
										childT.Value = lit
										childT.Position = pos
										childW.Children = append(childW.Children, childT)
									}

								}
							} else if tok == MarkerF {
								log.Print("Found Footnote marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerF {
										log.Print("Found Footnote end marker.\n\n")
										//p.unscan()
										break
									}
								}
							} else if tok == MarkerX {
								log.Print("Found Cross-Reference marker.")
								childW := &Content{}
								childW.Type = "marker"
								childW.Value = lit
								childW.Position = pos
								markerV.Children = append(markerV.Children, childW)
								for {
									tok, lit, pos = p.scanIgnoreWhitespace()
									if tok == EndMarkerX {
										log.Print("Found Cross-Reference end marker.\n\n")
										//p.unscan()
										break
									}
								}
							} else {
								child := &Content{}
								child.Type = "text"
								child.Value = lit
								child.Position = pos
								markerV.Children = append(markerV.Children, child)
							}
						}

					} else {
						return nil, fmt.Errorf("found %q, expected verse number", lit)
					}
				} else if tok == Text {
					// OK we've found a paragraph that
					// continues a previous verse
					log.Print("\n\n\nWe're in a Paragraph with Text now:\n\n")

					p.unscan()
					var verseNum *Content
					for _, c := range markerV.Children {
						if c.Type == "versenumber" {
							verseNum = c
							break
						}
					}
					newVerseNum := &Content{Type: "versenumber", Value: verseNum.Value, Children: verseNum.Children}
					markerPV := &Content{}
					markerPV.Type = "marker"
					markerPV.Value = "\\v"
					markerPV.Children = append(markerPV.Children, newVerseNum)
					// Add a new "sub-verse" marker
					markerSV := &Content{Type: "subverse", Value: "Sub-verse paragraph", Children: nil}
					markerPV.Children = append(markerPV.Children, markerSV)
					for {
						tok, lit, pos = p.scanIgnoreWhitespace()
						//log.Printf("Token: %v Lit: %v", tok, lit)
						if tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerP || tok == MarkerS {
							log.Printf("We're breaking because we hit %v:%v", tok, lit)
							p.unscan()
							break
						} else if tok == MarkerD {
							log.Print("Found Descriptive Title marker.")
							marker := &Content{}
							marker.Type = "marker"
							marker.Value = lit
							marker.Position = pos
							markerP.Children = append(markerP.Children, marker)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if !(tok == Text) {
									p.unscan()
									break
								} else {
									child := &Content{}
									child.Type = "description"
									child.Value = lit
									child.Position = pos
									marker.Children = append(marker.Children, child)
								}
							}
						} else if tok == MarkerQ1 {
							log.Print("Found Q1 marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
						} else if tok == MarkerQ2 {
							log.Print("Found Q2 marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
						} else if tok == MarkerWJ {
							log.Print("Found Jesus' Words marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerWJ {
									log.Print("Found Jesus' Words end markerPV.\n\n")
									//p.unscan()
									break
								} else {
									childT := &Content{}
									childT.Type = "text"
									childT.Value = lit
									childT.Position = pos
									childA.Children = append(childA.Children, childT)
								}
							}
						} else if tok == MarkerAdd {
							log.Print("Found Add marker.")
							childA := &Content{}
							childA.Type = "marker"
							childA.Value = lit
							childA.Position = pos
							markerPV.Children = append(markerPV.Children, childA)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerAdd {
									log.Print("Found Add end marker.\n\n")
									//p.unscan()
									break
								} else {
									log.Print("Found Add subject text.")
									childT := &Content{}
									childT.Type = "text"
									childT.Value = lit
									childT.Position = pos
									childA.Children = append(childA.Children, childT)
								}
							}
						} else if tok == MarkerW {
							log.Print("Found Wordlist marker.")
							childW := &Content{}
							childW.Type = "marker"
							childW.Value = lit
							childW.Position = pos
							markerPV.Children = append(markerPV.Children, childW)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerW {
									log.Print("Found Wordlist end marker.\n\n")
									break
								} else if tok == Citation {
									log.Print("Found Citation metadata.")
									childC := &Content{}
									childC.Type = "citation"
									childC.Value = lit
									childC.Position = pos
									childW.Children = append(childW.Children, childC)
								} else {
									log.Print("Found Citation subject text.")
									childT := &Content{}
									childT.Type = "text"
									childT.Value = lit
									childT.Position = pos
									childW.Children = append(childW.Children, childT)
								}

							}
						} else if tok == MarkerF {
							log.Print("Found Footnote marker.")
							childW := &Content{}
							childW.Type = "marker"
							childW.Value = lit
							childW.Position = pos
							markerPV.Children = append(markerPV.Children, childW)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerF {
									log.Print("Found Footnote end marker.\n\n")
									//p.unscan()
									break
								}
							}
						} else if tok == MarkerX {
							log.Print("Found Cross-Reference marker.")
							childW := &Content{}
							childW.Type = "marker"
							childW.Value = lit
							childW.Position = pos
							markerPV.Children = append(markerPV.Children, childW)
							for {
								tok, lit, pos = p.scanIgnoreWhitespace()
								if tok == EndMarkerX {
									log.Print("Found Cross-Reference end marker.\n\n")
									//p.unscan()
									break
								}
							}
						} else {
							childT := &Content{}
							childT.Type = "text"
							childT.Value = lit
							childT.Position = pos
							markerPV.Children = append(markerPV.Children, childT)
						}
					}
					markerP.Children = append(markerP.Children, markerPV)
					//break
				}
			}
		} else if tok == MarkerS {
			log.Print("Found Section Heading marker.")
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			/*} else if tok == MarkerV {
			log.Print("Found Verse marker.")
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			marker.Position = pos
			book.Children = append(book.Children, marker)
			tok, lit, pos = p.scanIgnoreWhitespace()
			if tok == Number {
				child := &Content{}
				child.Type = "versenumber"
				child.Value = lit
				child.Position = pos
				marker.Children = append(marker.Children, child)
				for {
					tok, lit, pos = p.scanIgnoreWhitespace()
					//if !(tok == Text || tok == Number || tok == MarkerW) {
					//	log.Printf("Invalid child token: %v", tok)
					//	p.unscan()
					//	break
					//} else if tok == MarkerW {
					if tok == 0x0085 || tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerP || tok == MarkerS {
						p.unscan()
						break
					} else if tok == MarkerWJ {
						log.Print("Found Jesus' Words marker.")
						childA := &Content{}
						childA.Type = "marker"
						childA.Value = lit
						childA.Position = pos
						marker.Children = append(marker.Children, childA)
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							if tok == EndMarkerWJ {
								log.Print("Found Jesus' Words end marker.\n\n")
								//p.unscan()
								break
							} else {
								childT := &Content{}
								childT.Type = "text"
								childT.Value = lit
								childT.Position = pos
								childA.Children = append(childA.Children, childT)
							}
						}
					} else if tok == MarkerAdd {
						log.Print("Found Add marker.")
						childA := &Content{}
						childA.Type = "marker"
						childA.Value = lit
						childA.Position = pos
						marker.Children = append(marker.Children, childA)
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							if tok == EndMarkerAdd {
								log.Print("Found Add end marker.\n\n")
								//p.unscan()
								break
							} else {
								log.Print("Found Add subject text.")
								childT := &Content{}
								childT.Type = "text"
								childT.Value = lit
								childT.Position = pos
								childA.Children = append(childA.Children, childT)
							}
						}
					} else if tok == MarkerW {
						log.Print("Found Wordlist marker.")
						markerEnd := false
						childW := &Content{}
						childW.Type = "marker"
						childW.Value = lit
						childW.Position = pos
						marker.Children = append(marker.Children, childW)
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							if tok == EndMarkerW {
								log.Print("Found Wordlist end marker.\n\n")
								markerEnd = true
								//p.unscan()
								break
							} else if tok == Citation {
								log.Print("Found Citation metadata.")
								childC := &Content{}
								childC.Type = "citation"
								childC.Value = lit
								childC.Position = pos
								childW.Children = append(childW.Children, childC)
							} else {
								log.Print("Found Citation subject text.")
								childT := &Content{}
								childT.Type = "text"
								childT.Value = lit
								childT.Position = pos
								childW.Children = append(childW.Children, childT)
							}

						}
						if markerEnd {
							//break
						}
					} else if tok == MarkerF {
						log.Print("Found Footnote marker.")
						childW := &Content{}
						childW.Type = "marker"
						childW.Value = lit
						childW.Position = pos
						marker.Children = append(marker.Children, childW)
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							if tok == EndMarkerF {
								log.Print("Found Footnote end marker.\n\n")
								//p.unscan()
								break
							}
						}
					} else if tok == MarkerX {
						log.Print("Found Cross-Reference marker.")
						childW := &Content{}
						childW.Type = "marker"
						childW.Value = lit
						childW.Position = pos
						marker.Children = append(marker.Children, childW)
						for {
							tok, lit, pos = p.scanIgnoreWhitespace()
							if tok == EndMarkerX {
								log.Print("Found Cross-Reference end marker.\n\n")
								//p.unscan()
								break
							}
						}
					} else {
						child := &Content{}
						child.Type = "text"
						child.Value = lit
						child.Position = pos
						marker.Children = append(marker.Children, child)
					}
				}

			} else {
				return nil, fmt.Errorf("found %q, expected verse number", lit)
			}*/
		} else if tok == EOF {
			break
		}
	}
	// Return the successfully parsed statement.
	return book, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string, pos int) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit, p.buf.pos
	}

	// Otherwise read the next token from the scanner.
	tok, lit, pos = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit, p.buf.pos = tok, lit, pos

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string, pos int) {
	tok, lit, pos = p.scan()
	if tok == Whitespace {
		tok, lit, pos = p.scan()
	}
	return
}

// scanAlnumAndIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanAlnumAndIgnoreWhitespace() (tok Token, lit string, pos int) {
	tok, lit, pos = p.scan()
	if tok == Whitespace {
		tok, lit, pos = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
