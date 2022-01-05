package input

import (
	"github.com/ebarti/dd-assignment/pkg/common"
	"os"
	"path/filepath"
	"sync/atomic"
)

type FileReader struct {
	filePath   string
	osFile     *os.File
	outputChan chan *Input
	stop       chan struct{}
	done       chan struct{}
	isDone     uint32
}

func NewFileReader(filePath string, outputChan chan *Input) common.Restartable {
	return &FileReader{
		filePath:   filePath,
		outputChan: outputChan,
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
}

func (f *FileReader) Start() error {
	if err := f.setup(); err != nil {
		return err
	}
	go f.readFile()
	return nil
}

func (f *FileReader) Stop() {
	if atomic.CompareAndSwapUint32(&f.isDone, 0, 1) {
		f.stop <- struct{}{}
		<-f.done
	}
}

func (f *FileReader) readFile() {
	defer f.cleanUp()
	origin := f.osFile.Name()
	for {
		inBuf := make([]byte, 4096)
		n, err := f.osFile.Read(inBuf)
		if err != nil {
			// Whether this error is expected or not, we need to exit
			return
		}
		select {
		case f.outputChan <- NewInput(inBuf[:n], origin):
			continue
		case <-f.stop:
			return
		}
	}
}

func (f *FileReader) cleanUp() {
	atomic.StoreUint32(&f.isDone, 1)
	f.osFile.Close()
	close(f.outputChan)
	close(f.done)
}

func (f *FileReader) setup() error {
	fullpath, err := filepath.Abs(f.filePath)
	if err != nil {
		return err
	}
	f.osFile, err = os.Open(fullpath)
	return err
}
