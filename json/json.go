package json

import (
	"encoding/json"
	"io"

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

	jsonEncoder := json.NewEncoder(w)
	jsonEncoder.SetIndent(" ", "  ")
	err = jsonEncoder.Encode(content)
	if err != nil {
		return err
	}

	return nil
}
