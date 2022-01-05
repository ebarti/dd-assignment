package output

import (
	"fmt"
	"github.com/ebarti/dd-assignment/pkg/agent/decoder"
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
)

func TestNewStreamer(t *testing.T) {
	// as we deal with goroutines, ensure there are no unexpected goroutines at the end of the test
	defer goleak.VerifyNone(t)

	inputChan := make(chan *decoder.DecodedMessage, 5)
	outputChan := make(chan *common.Message, 1)

	streamer := NewStreamer(inputChan, outputChan)
	if streamer == nil {
		t.Error("NewStreamer() returned nil")
	}
	// enqueue 5 messages
	for ii := 0; ii < 5; ii++ {
		message := []byte(fmt.Sprintf("test message %d", ii))
		inputChan <- &decoder.DecodedMessage{
			Content:            message,
			RawDataLen:         len(message),
			IngestionTimestamp: int64(ii),
			Origin:             "test",
		}
	}

	assert.NoError(t, streamer.Start(), "Start() returned an error")

	// verify the outputchan returns 5 messages
	for ii := 0; ii < 5; ii++ {
		msg := <-outputChan
		assert.Equal(t, fmt.Sprintf("test message %d", ii), string(msg.Content), "message content mismatch")
		assert.Equal(t, int64(ii), msg.IngestionTimestamp, "message ingestion timestamp mismatch")
		assert.Equal(t, "test", msg.Origin, "message origin mismatch")
	}
	// stop and verify outputchan is closed
	streamer.Stop()
	_, ok := <-outputChan
	assert.False(t, ok, "outputChan is not closed")
}
