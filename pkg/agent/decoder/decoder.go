package decoder

import (
	"bytes"
	"github.com/ebarti/dd-assignment/pkg/agent/input"
)

// Inspired by Datadog-agent's Decoder
const defaultContentLenLimit = 256 * 1000

// Decoder is the structure that decodes the incoming data
// No interface for Decoder as it has only one implementation
type Decoder struct {
	lineBuffer      *bytes.Buffer
	lineProcessor   LineProcessor
	InputChan       chan *input.Input
	OutputChan      chan *DecodedMessage
	matcher         EndLineMatcher
	contentLenLimit int
	rawDataLen      int
	origin          string
}

// NewDecoderWithEndLineMatcher initialize a Decoder with given EndLineMatcher strategy.
func NewDecoderWithEndLineMatcher(matcher EndLineMatcher) *Decoder {
	inputChan := make(chan *input.Input)
	outputChan := make(chan *DecodedMessage)
	lineLimit := defaultContentLenLimit
	lineProcessor := NewSingleLineProcessor(outputChan)

	return New(inputChan, outputChan, lineProcessor, lineLimit, matcher)
}

// New returns an initialized Decoder
func New(InputChan chan *input.Input, OutputChan chan *DecodedMessage, lineProcessor LineProcessor, contentLenLimit int, matcher EndLineMatcher) *Decoder {
	var lineBuffer bytes.Buffer
	return &Decoder{
		InputChan:       InputChan,
		OutputChan:      OutputChan,
		lineBuffer:      &lineBuffer,
		lineProcessor:   lineProcessor,
		contentLenLimit: contentLenLimit,
		matcher:         matcher,
	}
}

// Start starts the Decoder
func (d *Decoder) Start() error {
	if err := d.lineProcessor.Start(); err != nil {
		return err
	}
	go d.run()
	return nil
}

// Stop stops the Decoder
func (d *Decoder) Stop() {
	close(d.InputChan)
}

// run lets the Decoder handle data coming from InputChan
// the line processor will then feed the processed data to OutputChan
func (d *Decoder) run() {
	for data := range d.InputChan {
		d.origin = data.Origin
		d.decodeIncomingData(data.Content)
	}
	// Forward the stop down the chain
	d.lineProcessor.Stop()
}

func (d *Decoder) decodeIncomingData(inBuf []byte) {
	i, j := 0, 0
	n := len(inBuf)
	maxj := d.contentLenLimit - d.lineBuffer.Len()

	for ; j < n; j++ {
		if j == maxj {
			// send line because it is too long
			d.lineBuffer.Write(inBuf[i:j])
			d.rawDataLen += (j - i)
			d.sendLine()
			i = j
			maxj = i + d.contentLenLimit
		} else if d.matcher.Match(d.lineBuffer.Bytes(), inBuf, i, j) {
			d.lineBuffer.Write(inBuf[i:j])
			d.rawDataLen += (j - i)
			d.rawDataLen++ // account for the matching byte
			d.sendLine()
			i = j + 1 // skip the last bytes of the matched sequence
			maxj = i + d.contentLenLimit
		}
	}
	d.lineBuffer.Write(inBuf[i:j])
	d.rawDataLen += (j - i)
}

// sendLine copies content from lineBuffer which is passed to lineProcessor
func (d *Decoder) sendLine() {
	// Account for longer-than-1-byte line separator
	content := make([]byte, d.lineBuffer.Len()-(d.matcher.SeparatorLen()-1))
	copy(content, d.lineBuffer.Bytes())
	d.lineBuffer.Reset()
	d.lineProcessor.Handle(NewDecodedInput(content, d.rawDataLen, d.origin))
	d.rawDataLen = 0
}
