// VERBATIM COPY OF https://github.com/DataDog/datadog-agent/blob/7468a80b589b4f194db44f9e9926fd4936f31690/pkg/logs/decoder/matcher.go
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package decoder

var (
	// Utf16leEOL is the bytes sequence for UTF-16 Little-Endian end-of-line char
	Utf16leEOL = []byte{'\n', 0x00}
	// Utf16beEOL is the bytes sequence for UTF-16 Big-Endian end-of-line char
	Utf16beEOL = []byte{0x00, '\n'}
)

// EndLineMatcher defines the criterion to whether to end a line or not.
type EndLineMatcher interface {

	// Match takes the existing bytes (the entire slice) and the bytes to be
	// appended (appender[start:end+1]), and returns true if the combination
	// matches the end of line condition at the end.
	Match(exists []byte, appender []byte, start int, end int) bool
	SeparatorLen() int
}

// NewLineMatcher implements EndLineMatcher for line ending with '\n'
type NewLineMatcher struct {
}

// Match returns true whenever a '\n' (newline) is met.
func (n *NewLineMatcher) Match(exists []byte, appender []byte, start int, end int) bool {
	return appender[end] == '\n'
}

// SeparatorLen returns the length of the line separator
func (n *NewLineMatcher) SeparatorLen() int {
	return 1
}

// BytesSequenceMatcher defines the criterion to whether to end a line based on an arbitrary byte sequence
type BytesSequenceMatcher struct {
	sequence  []byte
	alignment int
}

// NewBytesSequenceMatcher Returns a new matcher based on custom bytes sequence.  Only matches
// that begin at a multiple of `alignment` are considered.
func NewBytesSequenceMatcher(sequence []byte, alignment int) *BytesSequenceMatcher {
	return &BytesSequenceMatcher{sequence, alignment}
}

// Match returns true whenever it finds a matching sequence at the end of append(exists, appender[start:end+1])
func (b *BytesSequenceMatcher) Match(exists []byte, appender []byte, start int, end int) bool {
	// Total read common is append(exists,appender[start:end]) and the Decoder just read appender[end]
	// Thus the separator sequence is checked against append(exists, appender[start:end+1]...)
	l := len(exists) + ((end + 1) - start)
	if l < len(b.sequence) {
		return false
	}
	if (l-len(b.sequence))%b.alignment != 0 {
		return false
	}
	seqIdx := len(b.sequence) - 1
	for ; seqIdx >= 0 && end >= start; seqIdx, end = seqIdx-1, end-1 {
		if appender[end] != b.sequence[seqIdx] {
			return false
		}
	}
	for i := len(exists) - 1; seqIdx >= 0 && i >= 0; seqIdx, i = seqIdx-1, i-1 {
		if exists[i] != b.sequence[seqIdx] {
			return false
		}
	}
	return true
}

// SeparatorLen return the number of byte to ignore
func (b *BytesSequenceMatcher) SeparatorLen() int {
	return len(b.sequence)
}