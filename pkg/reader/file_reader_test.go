package reader

import (
	"bytes"
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

const (
	testFilePath = "../../test_resources/sample_csv.txt"
)

func TestFileReader_readFile(t *testing.T) {
	// as we deal with goroutines, ensure there are no unexpected goroutines at the end of the test
	defer goleak.VerifyNone(t)

	fileSize := getFileSizeInBytes(t, testFilePath)
	fileContent, err := ioutil.ReadFile(testFilePath)
	assert.Nil(t, err)

	buf := bytes.Buffer{}
	logger := log.New(&buf, "", 0)
	outputChan := make(chan *common.Message)

	fileReader := NewFileReader(testFilePath, logger)
	fileReader.OutputChan = outputChan

	assert.NoErrorf(t, fileReader.Start(), "error starting file reader")
	var gotOutput []byte
	for input := range outputChan {
		content := input.Content
		content = append(content, '\n')
		gotOutput = append(gotOutput, content...)
	}
	assert.Equal(t, fileSize, int64(len(gotOutput)))
	assert.Equal(t, fileContent, gotOutput)
	assert.Equal(t, true, fileReader.IsStopped())
}

func TestFileReader_Stop(t *testing.T) {
	// as we deal with goroutines, ensure there are no unexpected goroutines at the end of the test
	defer goleak.VerifyNone(t)
	outputChan := make(chan *common.Message, 1)
	buf := bytes.Buffer{}
	logger := log.New(&buf, "", 0)
	fileReader := NewFileReader(testFilePath, logger)
	fileReader.OutputChan = outputChan
	assert.NoErrorf(t, fileReader.Start(), "error starting file reader")

	fileReader.Stop()
	// flush the first input that was sent
	<-outputChan
	// assert that the file reader has closed the output channel
	_, ok := <-outputChan
	assert.False(t, ok)
	assert.Equal(t, true, fileReader.IsStopped())
}

func getFileSizeInBytes(t *testing.T, path string) int64 {
	file, err := os.Open(path)
	if err != nil {
		t.Errorf("Error opening file: %s", err)
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		t.Errorf("Error getting stats for file: %s", err)
	}
	return stat.Size()
}
