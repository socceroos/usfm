package parser_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/socceroos/usfm/parser"
)

// Ensure the parser can parse strings into Node ASTs.
func TestParser(t *testing.T) {
	var tests = []struct {
		s    string
		Node *parser.Node
		err  string
	}{
		{
			s: `\id RUT T1 T2`,
			Node: &parser.Node{
				Type:  "book",
				Value: "RUT",
				Children: []*parser.Node{
					&parser.Node{
						Type:  "marker",
						Value: "\\id",
						Children: []*parser.Node{
							&parser.Node{Type: "bookcode", Value: "RUT"},
							&parser.Node{Type: "text", Value: "T1"},
							&parser.Node{Type: "text", Value: "T2"},
						},
					},
				},
			},
		},
		{
			s: `\ide 65001 - Unicode (UTF-8)`,
			Node: &parser.Node{
				Type:  "book",
				Value: "",
				Children: []*parser.Node{
					&parser.Node{
						Type:  "marker",
						Value: "\\ide",
						Children: []*parser.Node{
							&parser.Node{Type: "text", Value: "65001"},
							&parser.Node{Type: "text", Value: "-"},
							&parser.Node{Type: "text", Value: "Unicode"},
							&parser.Node{Type: "text", Value: "(UTF-8)"},
						},
					},
				},
			},
		},
		{
			s: `\c 42`,
			Node: &parser.Node{
				Type:  "book",
				Value: "",
				Children: []*parser.Node{
					&parser.Node{
						Type:  "marker",
						Value: "\\c",
						Children: []*parser.Node{
							&parser.Node{Type: "chapternumber", Value: "42"},
						},
					},
				},
			},
		},

		{
			s: `\v 1 T1 200`,
			Node: &parser.Node{
				Type:  "book",
				Value: "",
				Children: []*parser.Node{
					&parser.Node{
						Type:  "marker",
						Value: "\\v",
						Children: []*parser.Node{
							&parser.Node{Type: "versenumber", Value: "1"},
							&parser.Node{Type: "text", Value: "T1"},
							&parser.Node{Type: "text", Value: "200"},
						},
					},
				},
			},
		},
		{
			s: `\id RUT T1 T2 \ide UTF-8 \c 1 \v 1 T3 200 \v 28 T4 T5`,
			Node: &parser.Node{
				Type:  "book",
				Value: "RUT",
				Children: []*parser.Node{
					&parser.Node{
						Type:  "marker",
						Value: "\\id",
						Children: []*parser.Node{
							&parser.Node{Type: "bookcode", Value: "RUT"},
							&parser.Node{Type: "text", Value: "T1"},
							&parser.Node{Type: "text", Value: "T2"},
						},
					},
					&parser.Node{
						Type:  "marker",
						Value: "\\ide",
						Children: []*parser.Node{
							&parser.Node{Type: "text", Value: "UTF-8"},
						},
					},
					&parser.Node{
						Type:  "marker",
						Value: "\\c",
						Children: []*parser.Node{
							&parser.Node{Type: "chapternumber", Value: "1"},
						},
					},
					&parser.Node{
						Type:  "marker",
						Value: "\\v",
						Children: []*parser.Node{
							&parser.Node{Type: "versenumber", Value: "1"},
							&parser.Node{Type: "text", Value: "T3"},
							&parser.Node{Type: "text", Value: "200"},
						},
					},
					&parser.Node{
						Type:  "marker",
						Value: "\\v",
						Children: []*parser.Node{
							&parser.Node{Type: "versenumber", Value: "28"},
							&parser.Node{Type: "text", Value: "T4"},
							&parser.Node{Type: "text", Value: "T5"},
						},
					},
				},
			},
		},

		// Errors
		{s: `\id X T1 200`, err: `found "X", expected book code`},
		{s: `\v X T1 200`, err: `found "X", expected verse number`},
	}

	for i, tt := range tests {
		// Init Parser
		in := strings.NewReader(tt.s)
		Node, err := parser.NewParser(in).Parse()

		if !reflect.DeepEqual(tt.err, errstring(err)) {
			t.Errorf("%d. %q: error mismatch:\n  exp=%s\n  got=%s\n\n", i, tt.s, tt.err, err)
		} else if tt.err == "" && !reflect.DeepEqual(tt.Node, Node) {
			t.Errorf("%d. %q\n\nNode mismatch:\n\nexp=%#v\n\ngot=%#v\n\n", i, tt.s, tt.Node, Node)
		}
	}
}

// errstring returns the string representation of an error.
func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
