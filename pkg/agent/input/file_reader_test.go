package input

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"io/ioutil"
	"os"
	"testing"
)

func TestFileReader_readFile(t *testing.T) {
	// as we deal with goroutines, ensure there are no unexpected goroutines at the end of the test
	defer goleak.VerifyNone(t)

	testFilePath := "../../../test_resources/sample_csv.txt"
	fileSize := getFileSizeInBytes(t, testFilePath)
	fileContent, err := ioutil.ReadFile(testFilePath)
	assert.Nil(t, err)

	outputChan := make(chan *Input)

	fileReader := NewFileReader(testFilePath, outputChan)

	assert.NoErrorf(t, fileReader.Start(), "error starting file reader")
	var gotOutput []byte
	for input := range outputChan {
		gotOutput = append(gotOutput, input.Content...)
	}
	assert.Equal(t, fileSize, int64(len(gotOutput)))
	assert.Equal(t, fileContent, gotOutput)
}

func TestFileReader_Stop(t *testing.T) {
	// as we deal with goroutines, ensure there are no unexpected goroutines at the end of the test
	defer goleak.VerifyNone(t)
	//given
	testFilePath := "../../../test_resources/sample_csv.txt"
	outputChan := make(chan *Input, 1)

	//when
	fileReader := NewFileReader(testFilePath, outputChan)

	//then
	assert.NoErrorf(t, fileReader.Start(), "error starting file reader")

	fileReader.Stop()
	// flush the first input that was sent
	<-outputChan
	// assert that the file reader has closed the output channel
	_, ok := <-outputChan
	assert.False(t, ok)
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
