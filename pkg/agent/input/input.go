package input

// Input represents a chunk of line.
type Input struct {
	Content []byte
	Origin  string
}

// NewInput returns a new ingestion.
func NewInput(content []byte, origin string) *Input {
	return &Input{
		Content: content,
		Origin:  origin,
	}
}
