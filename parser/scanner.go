package parser

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Scanner represents a lexical scanner.
type Scanner struct {
	r        *bufio.Reader
	Pos      int
	LastSize int
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, bytes, err := s.r.ReadRune()
	if err != nil {
		return eof
	}

	// add the byte count to the position counter
	s.Pos = s.Pos + bytes
	s.LastSize = bytes

	return ch
}

// peek peeks the next byte - this doesn't advance the reader.
func (s *Scanner) peek(i int) rune {
	ch, err := s.r.Peek(i)
	if err != nil {
		return eof
	}
	r, _ := utf8.DecodeRune(ch)
	return r
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() {
	err := s.r.UnreadRune()
	if err == nil {
		s.Pos = s.Pos - s.LastSize
	}
}

// Scan returns the next token, literal value and position.
func (s *Scanner) Scan() (tok Token, lit string, pos int) {
	// Read the next rune.
	ch := s.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter then consume as an ident or reserved word.
	if unicode.IsSpace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if isBackslash(ch) || ch == 0x00B6 {
		s.unread()
		return s.scanMarker()
	} else if isLetter(ch) {
		s.unread()
		return s.scanText()
	} else if unicode.IsDigit(ch) {
		ch2 := s.peek(1)
		s.unread()
		if isLetter(ch2) {
			return s.scanText()
		}
		return s.scanNumber()
	}

	switch ch {
	case eof:
		return EOF, "", s.Pos - s.LastSize
	}

	return Illegal, string(ch), s.Pos - s.LastSize
}

// scanMarker consumes the current rune and read whole marker
func (s *Scanner) scanMarker() (tok Token, lit string, pos int) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent non-whitespace character into the buffer.
	// Whitespace character and EOF will cause the loop to exit.
	for i := 0; i <= 6; i++ {
		// FIXME: illegal?
		if ch := s.read(); ch == eof {
			break
		} else if unicode.IsSpace(ch) {
			s.unread()
			break
		} else if ch == 0x002A {
			buf.WriteRune(ch)
			break
		} else {
			buf.WriteRune(ch)
		}
		// Handle largest marker like \imte1
		// anything beyond that is illegal
		if i == 6 {
			return Illegal, buf.String(), s.Pos - s.LastSize
		}
	}

	//fmt.Printf("\nPosition: %v    Last Read Size: %v    Marker Buffer Length: %v    Marker: %v", s.Pos, s.LastSize, buf.Len(), buf.String())
	//fmt.Printf("\nMarker %v was scanned to %v byte position and we're going to calculate that it starts at %v\n", buf.String(), s.Pos, (s.Pos - (buf.Len() - 1)))

	size := buf.Len()
	position := s.Pos - size

	switch strings.ToUpper(buf.String()) {
	case `\ID`:
		return MarkerID, buf.String(), position
	case `\IDE`:
		return MarkerIde, buf.String(), position
	case `\IMTE`, `\IMTE1`:
		return MarkerImte1, buf.String(), position
	case `\H`:
		return MarkerH, buf.String(), position
	case `\C`:
		return MarkerC, buf.String(), position
	case `\V`:
		return MarkerV, buf.String(), position
	}

	return Illegal, buf.String(), position

}

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, lit string, pos int) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !unicode.IsSpace(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return Whitespace, buf.String(), s.Pos - buf.Len()
}

// scanText consumes the current rune and all contiguous ident runes.
func (s *Scanner) scanText() (tok Token, lit string, pos int) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	// Read every subsequent runes part of scripture into the buffer.
	// Non-letter, non-digit characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if isBackslash(ch) {
			s.unread()
			break
		} else if !unicode.IsLetter(ch) && !unicode.IsPunct(ch) && !unicode.IsDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return Text, buf.String(), s.Pos - buf.Len()
}

// scanNumber consumes the current rune and all contiguous number runes.
func (s *Scanner) scanNumber() (tok Token, lit string, pos int) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	buf.WriteRune(s.read())
	// Read every subsequent ident character into the buffer.
	// Non-ident characters and EOF will cause the loop to exit.
	for {
		if ch := s.read(); ch == eof {
			break
		} else if !unicode.IsDigit(ch) {
			s.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return Number, buf.String(), s.Pos - buf.Len()
}

// isLetter returns true if the rune is backslash (\)
func isLetter(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsPunct(ch)
}

// isBackslash returns true if the rune is backslash (\)
func isBackslash(ch rune) bool { return ch == '\\' }

// isPipe returns true if the rune is a Vertical Line or Pipe (|)
func isPipe(ch rune) bool { return ch == '|' }

// eof represents a marker rune for the end of the reader.
var eof = rune(0)
