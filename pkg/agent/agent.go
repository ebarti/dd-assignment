package agent

import (
	"github.com/ebarti/dd-assignment/pkg/agent/decoder"
	"github.com/ebarti/dd-assignment/pkg/agent/input"
	"github.com/ebarti/dd-assignment/pkg/agent/output"
	"github.com/ebarti/dd-assignment/pkg/common"
)

type Agent struct {
	receiver common.Restartable
	decoder  *decoder.Decoder
	sender   common.Restartable
}

func NewFileReaderAgent(filePath string, outputChan chan *common.Message) *Agent {
	decoder := decoder.NewDecoderWithEndLineMatcher(&decoder.NewLineMatcher{})
	receiver := input.NewFileReader(filePath, decoder.InputChan)
	sender := output.NewStreamer(decoder.OutputChan, outputChan)
	return &Agent{
		receiver: receiver,
		sender:   sender,
		decoder:  decoder,
	}
}

func (a *Agent) Start() error {
	// start services backwards first
	if err := a.sender.Start(); err != nil {
		return err
	}
	if err := a.decoder.Start(); err != nil {
		return err
	}
	return a.receiver.Start()
}

func (a *Agent) Stop() {
	// Stop propagates down the line
	a.receiver.Stop()
}
