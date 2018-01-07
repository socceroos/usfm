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

	// MarkerH represents '\h' marker for headings
	MarkerH

	// MarkerD represents '\d' marker for descriptive titles
	MarkerD

	// MarkerC represents '\c' marker
	MarkerC

	// MarkerV represents '\v' marker
	MarkerV

	// MarkerM represents '\m' marker for non-indented paragraphs
	MarkerM

	// MarkerP represents '\p' marker for paragraphs
	MarkerP

	// MarkerB represents '\b' marker for a line-break
	MarkerB

	// MarkerNB represents '\nb' marker for non-break continuing paragraphs
	MarkerNB

	// MarkerS represents '\s#' marker for section headings
	MarkerS

	// MarkerSP represents '\sp' marker for speaker identification
	MarkerSP

	// MarkerQ1 represents '\q1' marker for a poetry line
	MarkerQ1

	// MarkerQ2 represents '\q2' marker for a poetry line
	MarkerQ2

	// MarkerQS represents '\qs' marker for the word 'selah'
	MarkerQS

	// EndMarkerQS represents '\qs*' marker for the word 'selah'
	EndMarkerQS

	// MarkerW represents '\w' marker for wordlist/glossary/dictionary
	MarkerW

	// EndMarkerW represents '\w*' marker for wordlist/glossary/dictionary
	EndMarkerW

	// MarkerWJ represents '\wj' marker for Jesus' words
	MarkerWJ

	// EndMarkerWJ represents '\wj*' marker for Jesus' words
	EndMarkerWJ

	// MarkerF represents '\F' marker for footnotes
	MarkerF

	// EndMarkerF represents '\f*' marker for footnotes
	EndMarkerF

	// MarkerFR represents '\FR' marker for footnote reference
	MarkerFR

	// MarkerFT represents '\FT' marker for footnote text
	MarkerFT

	// MarkerX represents '\X' marker for cross-reference
	MarkerX

	// EndMarkerX represents '\f*' marker for cross-reference
	EndMarkerX

	// MarkerXO represents '\XO' marker for cross-reference origin reference
	MarkerXO

	// MarkerXT represents '\XT' marker for cross-reference text
	MarkerXT

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
