package json

import "io"

// Renderer render the parsed content
type Renderer interface {
	Render(w io.Writer, startKey int) (endKey int, err error)
}

// Options for rendering
type Options struct {
	Title string
}
