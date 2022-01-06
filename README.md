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
