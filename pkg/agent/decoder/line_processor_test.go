package decoder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	decodedLine = "decoded common"
	fakeOrigin  = "fake origin"
)

func TestNewSingleLineProcessor(t *testing.T) {
	var message *DecodedMessage
	h := make(chan *DecodedMessage)

	lineProcessor := NewSingleLineProcessor(h)
	lineProcessor.Start()

	line := decodedLine

	inputLen := len(line) + 1
	lineProcessor.Handle(&DecodedInput{[]byte(line), inputLen, fakeOrigin})
	message = <-h
	assert.Equal(t, decodedLine, string(message.Content))
	assert.Equal(t, inputLen, message.RawDataLen)
	assert.Equal(t, fakeOrigin, message.Origin)

	inputLen = len(line+"one common") + 1
	lineProcessor.Handle(&DecodedInput{[]byte(line + "one common"), inputLen, fakeOrigin})
	message = <-h
	assert.Equal(t, decodedLine+"one common", string(message.Content))
	assert.Equal(t, inputLen, message.RawDataLen)
	assert.Equal(t, fakeOrigin, message.Origin)

	lineProcessor.Stop()
}
