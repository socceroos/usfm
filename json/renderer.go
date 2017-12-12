package json

import "io"

// Renderer render the parsed content
type Renderer interface {
	Render(w io.Writer, startKey int) error
}

// Options for rendering
type Options struct {
	Title string
}
