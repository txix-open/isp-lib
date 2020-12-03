package streaming

import (
	"io"

	"github.com/integration-system/isp-lib/v2/isp"
)

type messageStreamFileWriter struct {
	stream DuplexMessageStream
}

func (m *messageStreamFileWriter) Write(p []byte) (n int, err error) {
	err = m.stream.Send(&isp.Message{Body: &isp.Message_BytesBody{BytesBody: p}})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (m *messageStreamFileWriter) Close() error {
	if err := m.stream.Send(FileEnd()); err != nil {
		return err
	}

	_, err := m.stream.Recv()
	switch err {
	case io.EOF, nil:
		if s, ok := m.stream.(interface{ CloseSend() error }); ok {
			return s.CloseSend()
		}
		return nil
	default:
		return err
	}
}

func NewMessageStreamFileWriter(stream DuplexMessageStream, bf BeginFile) (*messageStreamFileWriter, error) {
	err := stream.Send(bf.ToMessage())
	if err != nil {
		return nil, err
	}

	return &messageStreamFileWriter{stream: stream}, nil
}
