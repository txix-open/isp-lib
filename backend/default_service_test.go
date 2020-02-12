package backend

import (
	"testing"

	"github.com/integration-system/isp-lib/v2/streaming"
	"github.com/integration-system/isp-lib/v2/structure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

type testController struct {
}

func (testController) streamMethod(streaming.DuplexMessageStream, metadata.MD) error {
	return nil
}

func streamMethod(streaming.DuplexMessageStream, metadata.MD) error {
	return nil
}

var streamConsumerMethod streaming.StreamConsumer = func(streaming.DuplexMessageStream, metadata.MD) error {
	return nil
}

func TestResolveHandlers(t *testing.T) {
	descriptors := []structure.EndpointDescriptor{
		{
			Path:    "streaming/1",
			Handler: testController{}.streamMethod,
		},
		{
			Path:    "streaming/2",
			Handler: streamMethod,
		},
		{
			Path:    "streaming/3",
			Handler: streamConsumerMethod,
		},
		{
			Path:    "func/1",
			Handler: func() {},
		},
	}
	functions, streamFunctions, err := resolveHandlersByDescriptors(descriptors)

	assert.NoError(t, err)
	assert.Len(t, functions, 1)
	assert.Len(t, streamFunctions, 3)
}
