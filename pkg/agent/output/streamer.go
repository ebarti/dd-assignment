package output

import (
	"github.com/ebarti/dd-assignment/pkg/agent/decoder"
	"github.com/ebarti/dd-assignment/pkg/common"
)

// NewStreamer creates a Streamer with Sender signature
func NewStreamer(inputChan chan *decoder.DecodedMessage, outputChan chan *common.Message) common.Restartable {
	return &Streamer{
		inputChan:  inputChan,
		outputChan: outputChan,
	}
}

// Streamer is a Sender that forwards messages with the correct format to the given channel
// as they arrive (one by one)
type Streamer struct {
	inputChan  chan *decoder.DecodedMessage
	outputChan chan *common.Message
}

// Start starts the output.
func (s *Streamer) Start() error {
	go s.run()
	return nil
}

// Stop stops the output - blocking until inputChan is empty.
// It propagates the stop to the receiver channel by closing its input.
func (s *Streamer) Stop() {
	close(s.inputChan)
}

func (s *Streamer) run() {
	for input := range s.inputChan {
		s.outputChan <- &common.Message{
			Content:            input.Content,
			Origin:             input.Origin,
			IngestionTimestamp: input.IngestionTimestamp,
		}
	}
	close(s.outputChan)
}
