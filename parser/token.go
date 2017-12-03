package parser

// Token represents a lexical token.
type Token int

const (
	// Illegal represents an illegal/invalid character
	Illegal Token = iota

	// Whitespace represents a white space (" ", \t, \r, \n) character
	Whitespace

	// EOF represents end of file
	EOF

	// MarkerID represents '\id' or '\id1' marker
	MarkerID

	// MarkerIde represents '\ide' marker
	MarkerIde

	// MarkerImte1 represents '\imte' or '\imte1' marker
	MarkerImte1

	// MarkerC represents '\c' marker
	MarkerC

	// MarkerV represents '\v' marker
	MarkerV

	// MarkerP represents '\p' marker for paragraphs
	MarkerP

	// MarkerS represents '\s#' marker for section headings
	MarkerS

	// MarkerW represents '\w' marker for wordlist/glossary/dictionary
	MarkerW

	// EndMarkerW represents '\w*' marker for wordlist/glossary/dictionary
	EndMarkerW

	// MarkerWJ represents '\wj' marker for Jesus' words
	MarkerWJ

	// EndMarkerWJ represents '\wj*' marker for Jesus' words
	EndMarkerWJ

	// MarkerAdd represents '\add' marker for words added by the translator for clarity
	MarkerAdd

	// EndMarkerAdd represents '\add*' marker for words added by the translator for clarity
	EndMarkerAdd

	// Citation represents the citation/dict/thesaurus definitions in the \w marker
	Citation

	// Number represents a number (verse, chapter)
	Number

	// Text represents actual text
	Text
)
