package decoder

import "time"

// DecodedMessage represents a structured line.
type DecodedMessage struct {
	Content            []byte
	RawDataLen         int
	IngestionTimestamp int64
	Origin             string
}

// NewDecodedMessage returns a new output.
func NewDecodedMessage(content []byte, rawDataLen int, origin string) *DecodedMessage {
	return &DecodedMessage{
		Content:            content,
		RawDataLen:         rawDataLen,
		IngestionTimestamp: time.Now().UnixNano(),
		Origin:             origin,
	}
}

type DecodedInput struct {
	content    []byte
	rawDataLen int
	origin     string
}

// NewDecodedInput returns a new decoded ingestion.
func NewDecodedInput(content []byte, rawDataLen int, origin string) *DecodedInput {
	return &DecodedInput{
		content:    content,
		rawDataLen: rawDataLen,
		origin:     origin,
	}
}
