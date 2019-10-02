package resources

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const bufSize = 32 * 1024

type CsvOption func(opts *csvOpts)

type csvOpts struct {
	closeErrorHandler func(err error)
	csvSep            rune
}

func WithCloseErrorHandler(handler func(err error)) CsvOption {
	return func(opts *csvOpts) {
		opts.closeErrorHandler = handler
	}
}

func WithSeparator(sep rune) CsvOption {
	return func(opts *csvOpts) {
		opts.csvSep = sep
	}
}

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

func CompressedCsvReader(path string, readerHandler func(reader *csv.Reader) error, opts ...CsvOption) error {
	opt := newCsvOptions()
	for _, op := range opts {
		op(opt)
	}
	file, gzipReader, csvReader, err := makeReaders(path, opt.csvSep)
	defer func() {
		if gzipReader != nil {
			err := gzipReader.Close()
			if err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close gzip reader"))
			}
		}
		if file != nil {
			err := file.Close()
			if err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close file"))
			}
		}
	}()
	if err != nil {
		return err
	}
	return readerHandler(csvReader)
}

func CompressedCsvWriter(path string, writerHandler func(writer *csv.Writer) error, opts ...CsvOption) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	opt := newCsvOptions()
	for _, op := range opts {
		op(opt)
	}
	bufWriter := bufio.NewWriterSize(file, bufSize)
	gzipWriter := gzip.NewWriter(bufWriter)
	csvWriter := csv.NewWriter(gzipWriter)
	csvWriter.Comma = opt.csvSep
	defer func() {
		if csvWriter != nil {
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close csv writer"))
			}
		}
		if gzipWriter != nil {
			if err := gzipWriter.Flush(); err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "flash gzip writer"))
			}
			if err := gzipWriter.Close(); err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close gzip writer"))
			}
		}
		if bufWriter != nil {
			if err := bufWriter.Flush(); err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "flash buffer"))
			}
		}
		if file != nil {
			if err = file.Close(); err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close file"))
			}
		}
	}()
	return writerHandler(csvWriter)
}

func newCsvOptions() *csvOpts {
	return &csvOpts{
		closeErrorHandler: func(err error) {
		},
		csvSep: ';',
	}
}

func makeReaders(path string, csvSep rune) (*os.File, *gzip.Reader, *csv.Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, nil, errors.WithMessage(err, "open file")
	}

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		_ = file.Close()
		return nil, nil, nil, err
	}

	csvReader := csv.NewReader(gzipReader)
	csvReader.Comma = csvSep

	return file, gzipReader, csvReader, nil
}
