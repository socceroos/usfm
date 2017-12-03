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
	for {
		// Read a field.
		tok, lit := p.scanIgnoreWhitespace()
		if tok == MarkerID {
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			book.Children = append(book.Children, marker)
			tok, lit = p.scanIgnoreWhitespace()
			if tok == Text && len([]rune(lit)) == 3 {
				child := &Content{}
				child.Type = "bookcode"
				child.Value = lit
				book.Value = lit
				marker.Children = append(marker.Children, child)
				for {
					tok, lit = p.scanIgnoreWhitespace()
					if !(tok == Text || tok == Number) {
						p.unscan()
						break
					} else {
						child := &Content{}
						child.Type = "text"
						child.Value = lit
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
			book.Children = append(book.Children, marker)
			for {
				tok, lit = p.scanIgnoreWhitespace()
				if !(tok == Text || tok == Number) {
					p.unscan()
					break
				} else {
					child := &Content{}
					child.Type = "text"
					child.Value = lit
					marker.Children = append(marker.Children, child)
				}
			}
		} else if tok == MarkerC {
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			book.Children = append(book.Children, marker)
			tok, lit = p.scanIgnoreWhitespace()
			if tok == Number {
				child := &Content{}
				child.Type = "chapternumber"
				child.Value = lit
				marker.Children = append(marker.Children, child)
			} else {
				return nil, fmt.Errorf("found %q, expected chapter number", lit)
			}
		} else if tok == MarkerP {
			log.Print("Found Paragraph marker.")
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			book.Children = append(book.Children, marker)
		} else if tok == MarkerS {
			log.Print("Found Section Heading marker.")
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			book.Children = append(book.Children, marker)
		} else if tok == MarkerV {
			marker := &Content{}
			marker.Type = "marker"
			marker.Value = lit
			book.Children = append(book.Children, marker)
			tok, lit = p.scanIgnoreWhitespace()
			if tok == Number {
				child := &Content{}
				child.Type = "versenumber"
				child.Value = lit
				marker.Children = append(marker.Children, child)
				for {
					log.Printf("Scanning for Verse children...")
					tok, lit = p.scanIgnoreWhitespace()
					//if !(tok == Text || tok == Number || tok == MarkerW) {
					//	log.Printf("Invalid child token: %v", tok)
					//	p.unscan()
					//	break
					//} else if tok == MarkerW {
					log.Printf("Token: %v", tok)
					log.Printf("Lit: %v", lit)
					//if !(tok == Text || tok == MarkerAdd || tok == EndMarkerAdd || tok == Number || tok == MarkerW || tok == EndMarkerW) {
					if tok == 0x0085 || tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerP || tok == MarkerS {
						p.unscan()
						break
					} else if tok == MarkerAdd {
						log.Print("Found Add marker.")
						childA := &Content{}
						childA.Type = "marker"
						childA.Value = lit
						marker.Children = append(marker.Children, childA)
						for {
							tok, lit = p.scanIgnoreWhitespace()
							log.Printf("Add Token: %v", tok)
							log.Printf("Add Lit: %v", lit)
							if tok == EndMarkerAdd {
								log.Print("Found Add end marker.\n\n")
								//p.unscan()
								break
							} else {
								log.Print("Found Add subject text.")
								childT := &Content{}
								childT.Type = "text"
								childT.Value = lit
								childA.Children = append(childA.Children, childT)
							}
						}
					} else if tok == MarkerW {
						log.Print("Found Wordlist marker.")
						markerEnd := false
						childW := &Content{}
						childW.Type = "marker"
						childW.Value = lit
						marker.Children = append(marker.Children, childW)
						for {
							tok, lit = p.scanIgnoreWhitespace()
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
								childW.Children = append(childW.Children, childC)
							} else {
								log.Print("Found Citation subject text.")
								childT := &Content{}
								childT.Type = "text"
								childT.Value = lit
								childW.Children = append(childW.Children, childT)
							}

						}
						if markerEnd {
							//break
						}
					} else {
						log.Print("Found verse text.")
						child := &Content{}
						child.Type = "text"
						child.Value = lit
						marker.Children = append(marker.Children, child)
					}
				}

			} else {
				return nil, fmt.Errorf("found %q, expected verse number", lit)
			}
		} else if tok == EOF {
			break
		}
	}
	// Return the successfully parsed statement.
	return book, nil
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.n != 0 {
		p.buf.n = 0
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == Whitespace {
		tok, lit = p.scan()
	}
	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.n = 1 }
