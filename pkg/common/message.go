package common

// Message represents a log line sent to the backend, with its metadata
// See note.md for more information
type Message struct {
	Content            []byte
	Origin             string
	IngestionTimestamp int64
}
