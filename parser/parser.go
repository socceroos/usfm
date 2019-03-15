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

// Parse parses a USFM formatted book Content
func (p *Parser) Parse() (*Content, error) {
	log.Printf("Scanning for book...")
	// Init book content
	book := &Content{}
	book.Type = "book"

	markerV := &Content{}
	for {
		// Read a field.
		tok, lit, pos := p.scanIgnoreWhitespace()

		// \id marker
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
			// \c Chapter
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
		} else if tok == MarkerV {
			bug := false
			if pos == 81913 {
				bug = true
				log.Print(lit)
			}
			markerV = &Content{}
			markerV.Type = "marker"
			markerV.Value = lit
			markerV.Position = pos
			book.Children = append(book.Children, markerV)
			tok, lit, pos = p.scanIgnoreWhitespace()

			if tok == Number {
				if bug == true {
					log.Print(lit)
				}

				child := &Content{}
				child.Type = "versenumber"
				child.Value = lit
				child.Position = pos

				markerV.Children = append(markerV.Children, child)
				for {
					tok, lit, pos = p.scanIgnoreWhitespace()
					//log.Printf("Token: %v Lit: %v", tok, lit)
					if tok == 0x0085 || tok == EOF || tok == MarkerV || tok == MarkerC || tok == MarkerID {
						p.unscan()
						break
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

	for {
		tok, lit, pos = p.scan()
		if tok != Whitespace {
			break
		}
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
