package pipeline

import (
	"github.com/ebarti/dd-assignment/pkg/backend/logs"
	"github.com/ebarti/dd-assignment/pkg/backend/monitors"
	"github.com/ebarti/dd-assignment/pkg/common"
	"sync"
	"sync/atomic"
)

type LogProcessorFunc func(*common.Message) (*logs.ProcessedLog, error)

type LogPipeline struct {
	inputChan        chan *common.Message
	monitors         []chan *logs.ProcessedLog
	OutputChan       chan *logs.ProcessedLog
	logProcessorFunc LogProcessorFunc
	done             chan struct{}
	isDone           uint32
}

func NewLogPipeline(
	logProcessorFunc LogProcessorFunc,
) *LogPipeline {

	return &LogPipeline{
		inputChan:        make(chan *common.Message),
		OutputChan:       make(chan *logs.ProcessedLog),
		logProcessorFunc: logProcessorFunc,
		done:             make(chan struct{}),
	}
}

func (i *LogPipeline) From(inputChan chan *common.Message) {
	i.inputChan = inputChan
}

func (i *LogPipeline) AddMonitors(logMonitor []*monitors.LogMonitor) {
	for _, monitor := range logMonitor {
		i.addMonitoredChannel(monitor.InputChan)
	}
}

func (i *LogPipeline) addMonitoredChannel(c chan *logs.ProcessedLog) {
	i.monitors = append(i.monitors, c)
}

func (i *LogPipeline) Start() error {
	go i.run()
	return nil
}

func (i *LogPipeline) Stop() {
	if atomic.CompareAndSwapUint32(&i.isDone, 0, 1) {
		close(i.inputChan)
		<-i.done
	}
}

func (i *LogPipeline) run() {
	defer i.cleanUp()
	for input := range i.inputChan {
		i.process(input)
	}
}

func (i *LogPipeline) cleanUp() {
	close(i.OutputChan)
	for _, output := range i.monitors {
		close(output)
	}
	atomic.StoreUint32(&i.isDone, 1)
	close(i.done)
}

func (i *LogPipeline) process(msg *common.Message) {
	log, err := i.logProcessorFunc(msg)
	if err != nil {
		return
	}
	wg := sync.WaitGroup{}
	wg.Add(len(i.monitors) + 1)
	for _, output := range i.monitors {
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
