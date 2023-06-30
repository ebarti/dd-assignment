Built Late 2021

## Datadog Take Home Exercise
#### By Eloi Barti

---
![](coverage_badge.png)

## Running the exercise
There are two options to run the exercise:
1. Build from source
2. Use one of the pre-built binaries

### Build from source
1. Install Go 1.17+
2. Clone the repository to $GOPATH/src/github.com/ebarti/dd-assignment
3. Run `go build .`
4. Binary is built as `dd-assignment`

### Using one of the pre-built binaries
Check the `/bin` directory for the binaries.


## Getting Started
Run `./dd-assignment --help` to see the available options.

## About my solution
I chose Go for my implementation as the language features a set of concurrency primitives that are very useful for this exercise.
I used [cobra](https://github.com/spf13/cobra) to handle some CLI basics like flags, help menus, etc.

Here is the architecture of the solution:

```
 SERVICE
┌───────────────────────────────────────────────┐
│  ┌─────────────────┐                          │
│  │  FILE READER    │                          │
│  └───────┬─────────┘                          │
│          │Message           ┌───────────────┐ │
│          │              ┌───► LOG MONITOR   │ │
│  ┌───────▼────────┐     │   │               │ │
│  │                ├─────┘   └───────────────┘ │
│  │  LOG PIPELINE  │ProcessedLog               │
│  │                ├─────┐   ┌───────────────┐ │
│  └───────┬────────┘     │   │               │ │
│          │              └───► LOG MONITOR   │ │
│          │ ProcessedLog     └───────────────┘ │
│  ┌───────▼────────┐               .....       │
│  │METRICS PIPELINE│                           │
│  └───────┬────────┘                           │
│          │ []MetricSample                     │
│  ┌───────▼────────┐                           │
│  │    METRICS     │                           │
│  │   AGGREGATOR   │                           │
│  └────────────────┘                           │
└───────────────────────────────────────────────┘
```

### File Reader
- Reads line by line from the input file
- Creates a `Message` for each line
- Feeds the message to its output channel

### Log Pipeline
- Reads the `Message` from its input channel
- Processes the message according to its `LogProcessorFunc`
- Asynchronously feeds the processed message to its output channel and all observing `LogMonitor`s

### Log Monitor
- Reads the processedLog from its input channel
- Generates a metric sample for each processed log
- Aggregates the metric according to the monitor's configuration
- Monitors the value of the metric and, if the value is above the threshold, prints an alert to the console

### Metrics Pipeline
- Reads the processedLog from its input channel
- Generates a (set of) metric sample(s) for each processed log  `LogFilter` and a `LogMeasure`. Tags the metrics according to its `groupBy` configuration
- Sends the samples to its output channel 

### Metrics Aggregator
- Reads the metric samples from its input channel
- Aggregates metrics by interval
- When an interval is complete, it flushes the metrics and renders them to the console

## Notes
- As the log monitor and the metrics aggregator run on different goroutines, **the order of the console output is not guaranteed**.

- As I think it is good practice (and it is so simple to do in Go), I have vendored all the project dependencies via `go mod vendor` and are under the `vendor` directory. 

## Future improvements
### Order of output
If the output needs to be ordered, a custom logger could be implemented. 
This logger would be fed logs and timestamps, and would keep a cache of some logs to be printed.
Multiple conditions could be implemented on how the cache is flushed, like:
- If the cache is full (basic)
- If the last M logs received are all of increasing timestamp
The flushing would obviously flush the cache in increasing order of timestamp.

### Reader becomes a fully fledged agent
My initial implementation of the solution included a simple "datadog-like" agent. However,
due to the added complexity and the little value it added to my solution, I decided to scrap it.
A future improvement would be to create a fully fledged agent, able to tail log files in real-time, handle log file rotations, etc.

### Backend exposes an API
Right now the file-reader (the closest to a datadog-agent in this solution) is using a channel to send the 
messages it reads. A future improvement would be to decouple it from the backend by exposing an API
for the agent to send the messages.

### Decoupling
There is a bit of coupling between the different components of the solution. Although some of my components
do actually already implement interfaces like `Restartable`, it would be a good idea to 
clean up the code so that the components are referenced by their interface rather than a pointer.

### Types of Metrics
Right now only count based metrics are supported. Some other metrics could be added, like:
- Rates
- Gauges
- Histograms
- etc.

### Service metrics
Right now there are no metrics on the actual backend. It would be nice to know things like
ingestion latency, how many resources the backend is using, etc. 

