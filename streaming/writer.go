package streaming

import (
	"io"

	"github.com/integration-system/isp-lib/v2/isp"
)

type fileStreamWriter struct {
	stream    DuplexMessageStream
	beginFile BeginFile
}

type FileStream interface {
	BeginFile() BeginFile
}

func (m *fileStreamWriter) Write(p []byte) (n int, err error) {
	err = m.stream.Send(&isp.Message{Body: &isp.Message_BytesBody{BytesBody: p}})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (m *fileStreamWriter) Close() error {
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

func (m *fileStreamWriter) BeginFile() BeginFile {
	return m.beginFile
}

func NewFileStreamWriter(stream DuplexMessageStream, bf BeginFile) (io.WriteCloser, error) {
	err := stream.Send(bf.ToMessage())
	if err != nil {
		return nil, err
	}

	return &fileStreamWriter{stream: stream, beginFile: bf}, nil
}
