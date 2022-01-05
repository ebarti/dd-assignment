package agent

import (
	"github.com/ebarti/dd-assignment/pkg/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"io/ioutil"
	"strings"
	"testing"
)

func TestAgent(t *testing.T) {
	defer goleak.VerifyNone(t)
	//given
	testFilePath := "../../test_resources/short_sample_csv.txt"
	outputChan := make(chan *common.Message)
	fileContent, err := ioutil.ReadFile(testFilePath)
	splitFileContent := strings.Split(string(fileContent), "\n")
	assert.Nil(t, err)
	agent := NewFileReaderAgent(testFilePath, outputChan)

	assert.NoError(t, agent.Start(), "Start() returned an error")
	ii := 0
	for msg := range outputChan {
		assert.Equal(t, splitFileContent[ii], string(msg.Content), "Line content does not match")
		ii++
	}
	_, ok := <-outputChan
	assert.False(t, ok, "outputChan is not closed")
}

func TestAgent_Stop(t *testing.T) {
	// as we deal with goroutines, ensure there are no unexpected goroutines at the end of the test
	defer goleak.VerifyNone(t)
	// test stop with a large csv so we can test the stop logic
	testFilePath := "../../test_resources/sample_csv.txt"
	outputChan := make(chan *common.Message, 1)
	agent := NewFileReaderAgent(testFilePath, outputChan)

	assert.NoError(t, agent.Start(), "Start() returned an error")
	// drain one message to verify it works
	_, ok := <-outputChan
	assert.True(t, ok, "outputChan should not be closed")
	agent.Stop()
	// Draining the output is required as the elements down the line wait for their buffer to be flushed before stopping
	// This test is useful in that it verifies that everything stops and there are no goroutine leaks
	// However, we cannot do any useful assertions here as we cannot know how many messages are in the channel
	for msg := range outputChan {
		_ = msg
	}

}
