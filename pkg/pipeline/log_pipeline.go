package pipeline

import (
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/ebarti/dd-assignment/pkg/logs"
	"github.com/ebarti/dd-assignment/pkg/monitors"
	"sync"
	"sync/atomic"
)

// LogProcessorFunc is the message processing function that runs for every message received
type LogProcessorFunc func(*common.Message) (*logs.ProcessedLog, error)

// LogPipeline is a pipeline that processes read log lines as *common.Message and returns *logs.ProcessedLog
type LogPipeline struct {
	inputChan        chan *common.Message
	monitors         []chan *logs.ProcessedLog
	OutputChan       chan *logs.ProcessedLog
	logProcessorFunc LogProcessorFunc
	done             chan struct{}
	isDone           uint32
}

// NewLogPipeline creates a new LogPipeline
func NewLogPipeline(logProcessorFunc LogProcessorFunc) *LogPipeline {
	return &LogPipeline{
		inputChan:        make(chan *common.Message),
		OutputChan:       make(chan *logs.ProcessedLog),
		logProcessorFunc: logProcessorFunc,
		done:             make(chan struct{}),
	}
}

// From is used to set the input channel for the pipeline
func (i *LogPipeline) From(inputChan chan *common.Message) {
	i.inputChan = inputChan
}

// AddMonitors is used to add an array of *monitors.LogMonitor to the pipeline
func (i *LogPipeline) AddMonitors(logMonitor []*monitors.LogMonitor) {
	for _, monitor := range logMonitor {
		i.addMonitoredChannel(monitor.InputChan)
	}
}

// addMonitoredChannel is a helper function that adds a monitor input chan to the monitors array
// used to make the pipeline testable
func (i *LogPipeline) addMonitoredChannel(c chan *logs.ProcessedLog) {
	i.monitors = append(i.monitors, c)
}

// Start starts the pipeline
func (i *LogPipeline) Start() error {
	go i.run()
	return nil
}

// Stop stops the pipeline
func (i *LogPipeline) Stop() {
	if atomic.CompareAndSwapUint32(&i.isDone, 0, 1) {
		close(i.inputChan)
		<-i.done
	}
}

// IsStopped returns true if the pipeline is stopped
func (i *LogPipeline) IsStopped() bool {
	return atomic.LoadUint32(&i.isDone) == 1
}

// run is the main loop of the pipeline
func (i *LogPipeline) run() {
	defer i.cleanUp()
	for input := range i.inputChan {
		i.process(input)
	}
}

// cleanUp closes the output channel, the input channel of the observing monitors and stores the done state
func (i *LogPipeline) cleanUp() {
	close(i.OutputChan)
	for _, output := range i.monitors {
		close(output)
	}
	atomic.StoreUint32(&i.isDone, 1)
	close(i.done)
}

// process is the main function that processes a message and forwards the logs.ProcessedLog to all observing channels in a non-blocking fashion
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
