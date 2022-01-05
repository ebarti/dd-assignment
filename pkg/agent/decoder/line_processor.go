package decoder

import "github.com/ebarti/dd-assignment/pkg/common"

type LineProcessor interface {
	Handle(input *DecodedInput)
	common.Restartable
}

// SingleLineProcessor processes lines one by one
type SingleLineProcessor struct {
	inputChan  chan *DecodedInput
	outputChan chan *DecodedMessage
}

// NewSingleLineProcessor returns a new SingleLineProcessor.
func NewSingleLineProcessor(outputChan chan *DecodedMessage) *SingleLineProcessor {
	return &SingleLineProcessor{
		inputChan:  make(chan *DecodedInput),
		outputChan: outputChan,
	}
}

// Handle puts all new lines into a channel for later processing.
func (p *SingleLineProcessor) Handle(input *DecodedInput) {
	p.inputChan <- input
}

// Start starts the parser.
func (p *SingleLineProcessor) Start() error {
	go p.run()
	return nil
}

// Stop stops the parser.
func (p *SingleLineProcessor) Stop() {
	close(p.inputChan)
}

// run consumes new lines and processes them.
func (p *SingleLineProcessor) run() {
	for input := range p.inputChan {
		p.process(input)
	}
	close(p.outputChan)
}

func (p *SingleLineProcessor) process(input *DecodedInput) {
	// Forward to next step of the pipeline
	// Additional parsing/processing could be done at this step
	p.outputChan <- NewDecodedMessage(input.content, input.rawDataLen, input.origin)
}
