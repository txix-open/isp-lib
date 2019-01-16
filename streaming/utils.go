package streaming

import (
	"github.com/integration-system/isp-lib/proto/stubs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"os"
)

const (
	bufferSize = 4096
)

type FileFactory func(bf BeginFile) (*os.File, error)

func ReadFile(stream DuplexMessageStream, fileFactory func(bf BeginFile) (*os.File, error), sendResponse bool) (*BeginFile, error) {
	msg, err := stream.Recv()
	if err != nil {
		return nil, err
	}
	bf := &BeginFile{}
	err = bf.FromMessage(msg)
	if err != nil {
		return nil, err
	}

	f, err := fileFactory(*bf)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return bf, err
	}

	for {
		msg, err = stream.Recv()
		isEof := IsEndOfFile(msg)
		if isEof || err == io.EOF {
			if sendResponse {
				err := stream.Send(bf.ToMessage())
				return bf, err
			} else {
				return bf, nil
			}
		}
		bytes := msg.GetBytesBody()
		if bytes == nil {
			return bf, status.Errorf(codes.InvalidArgument, "Expected bytes array")
		}
		_, err := f.Write(bytes)
		if err != nil {
			return bf, err
		}
	}

	return bf, nil
}

func WriteFile(stream DuplexMessageStream, path string, bf BeginFile) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return err
	}

	err = stream.Send(bf.ToMessage())
	if err != nil {
		return err
	}

	buf := make([]byte, bufferSize)
	for {
		n, err := f.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		err = stream.Send(&isp.Message{Body: &isp.Message_BytesBody{BytesBody: buf[:n]}})
		if err != nil {
			return err
		}
	}

	return nil
}
