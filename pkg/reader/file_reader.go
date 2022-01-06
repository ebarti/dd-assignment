package reader

import (
	"bufio"
	"github.com/ebarti/dd-assignment/pkg/common"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

const defaultContentLenLimit = 256 * 1000

// FileReader is a reader that reads a file from a directory
type FileReader struct {
	filePath   string
	osFile     *os.File
	scanner    *bufio.Scanner
	OutputChan chan *common.Message
	stop       chan struct{}
	done       chan struct{}
	isDone     uint32
	logger     *log.Logger
}

// NewFileReader creates a new FileReader
func NewFileReader(filePath string, logger *log.Logger) *FileReader {
	return &FileReader{
		filePath:   filePath,
		OutputChan: make(chan *common.Message),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
		logger:     logger,
	}
}

// Start starts the FileReader
func (f *FileReader) Start() error {
	if err := f.setup(); err != nil {
		return err
	}
	go f.readFile()
	return nil
}

// Stop stops the FileReader
func (f *FileReader) Stop() {
	if atomic.CompareAndSwapUint32(&f.isDone, 0, 1) {
		f.stop <- struct{}{}
		<-f.done
	}
}

// IsStopped returns true if the FileReader is stopped
func (f *FileReader) IsStopped() bool {
	return atomic.LoadUint32(&f.isDone) == 1
}

// readFile reads the file line by line, builds a common.Message for each line and sends it to the FileReader's OutputChan
func (f *FileReader) readFile() {
	defer f.cleanUp()
	origin := f.osFile.Name()
	for {
		if !f.scanner.Scan() {
			if err := f.scanner.Err(); err != nil {
				f.logger.Panicf("Error while reading file %s: %s", origin, err)
			}
			return
		}
		select {
		case f.OutputChan <- common.NewMessage(f.scanner.Bytes(), origin, time.Now().Unix()):
			continue
		case <-f.stop:
			return
		}
	}
}

// cleanUp closes the FileReader's osFile, as well as its OutputChan and stores the done state
func (f *FileReader) cleanUp() {
	f.osFile.Close()
	close(f.OutputChan)
	atomic.StoreUint32(&f.isDone, 1)
	close(f.done)
}

// setup opens the FileReader's osFile and sets up the scanner with the defaultContentLenLimit as the buffer size
func (f *FileReader) setup() error {
	fullpath, err := filepath.Abs(f.filePath)
	if err != nil {
		return err
	}
	f.osFile, err = os.Open(fullpath)
	if err != nil {
		return err
	}
	f.scanner = bufio.NewScanner(f.osFile)
	buffer := make([]byte, 0, defaultContentLenLimit)
	f.scanner.Buffer(buffer, defaultContentLenLimit)
	return nil
}
