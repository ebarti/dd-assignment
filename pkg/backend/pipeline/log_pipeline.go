package pipeline

import (
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/common"
	"sync"
)

type LogProcessorFunc func(*common.Message) (*logs.ProcessedLog, error)

type LogPipeline struct {
	inputChan        chan *common.Message
	Monitors         []chan *logs.ProcessedLog
	OutputChan       chan *logs.ProcessedLog
	logProcessorFunc LogProcessorFunc
}

func NewLogPipeline(
	inputChan chan *common.Message,
	outputChan chan *logs.ProcessedLog,
	logProcessorFunc LogProcessorFunc,
) *LogPipeline {

	return &LogPipeline{
		inputChan:        inputChan,
		OutputChan:       outputChan,
		logProcessorFunc: logProcessorFunc,
	}
}

func (i *LogPipeline) AddMonitor(logMonitor *monitors.LogMonitor) {
	i.Monitors = append(i.Monitors, logMonitor.InputChan)
}

func (i *LogPipeline) Start() error {
	go i.run()
	return nil
}

func (i *LogPipeline) Stop() {
	close(i.inputChan)
}

func (i *LogPipeline) run() {
	defer i.cleanUp()
	for input := range i.inputChan {
		i.process(input)
	}
}

func (i *LogPipeline) cleanUp() {
	close(i.OutputChan)
	for _, output := range i.Monitors {
		close(output)
	}
}

func (i *LogPipeline) process(msg *common.Message) {
	log, err := i.logProcessorFunc(msg)
	if err != nil {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(i.Monitors) + 1)
	for _, output := range i.Monitors {
		go func(output chan *logs.ProcessedLog) {
			defer wg.Done()
			output <- log
		}(output)
	}
	go func() {
		defer wg.Done()
		i.OutputChan <- log
	}()
	wg.Wait()
}
