package resources

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"github.com/integration-system/isp-lib/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const bufSize = 32 * 1024

func OpenTempFile() (io.WriteCloser, string, error) {
	path, err := GetTempFilePath()
	if err != nil {
		return nil, "", err
	}

	f, err := os.Create(path)
	if err != nil {
		return nil, "", err
	}

	return f, path, nil
}

func GetTempFilePath() (string, error) {
	if temp, err := ioutil.TempDir("", ""); err != nil {
		return "", err
	} else {
		return filepath.Join(temp, "info"), nil
	}
}

func CompressedCsvReader(path string, csvSep rune, readerHandler func(reader *csv.Reader) error) error {
	file, gzipReader, csvReader, err := makeReaders(path, csvSep)
	defer func() {
		if gzipReader != nil {
			_ = gzipReader.Close()
		}
		if file != nil {
			_ = file.Close()
		}
	}()
	if err != nil {
		return err
	}
	return readerHandler(csvReader)
}

func CompressedCsvWriter(path string, csvSep rune, writerHandler func(writer *csv.Writer) error) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	bufWriter := bufio.NewWriterSize(file, bufSize)
	gzipWriter := gzip.NewWriter(bufWriter)
	csvWriter := csv.NewWriter(gzipWriter)
	csvWriter.Comma = csvSep
	defer func() {
		if csvWriter != nil {
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				logger.Error(err)
			}
		}
		if gzipWriter != nil {
			if err := gzipWriter.Flush(); err != nil {
				logger.Error(err)
			}
			if err := gzipWriter.Close(); err != nil {
				logger.Error(err)
			}
		}
		if bufWriter != nil {
			if err := bufWriter.Flush(); err != nil {
				logger.Error(err)
			}
		}
		if file != nil {
			if err = file.Close(); err != nil {
				logger.Error(err)
			}
		}
	}()
	return writerHandler(csvWriter)
}

func makeReaders(path string, csvSep rune) (*os.File, *gzip.Reader, *csv.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "open file")
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, "invalid gzip format")
		_ = file.Close()
		return nil, nil, nil, err
	}

	csvReader := csv.NewReader(gzipReader)
	csvReader.Comma = csvSep

	return file, gzipReader, csvReader, nil
}
