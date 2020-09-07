package resources

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

const bufSize = 32 * 1024

type CsvOption func(opts *csvOpts)

type csvOpts struct {
	closeErrorHandler func(err error)
	csvSep            rune
	compressed        bool
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

func CsvReader(readCloser io.ReadCloser, readerHandler func(reader *csv.Reader) error, opts ...CsvOption) error {
	opt := newCsvOptions()
	for _, op := range opts {
		op(opt)
	}

	gzipReader, csvReader, err := makeReaders(readCloser, *opt)
	defer func() {
		if gzipReader != nil && opt.compressed {
			err := gzipReader.Close()
			if err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close gzip reader"))
			}
		}
		if readCloser != nil {
			err := readCloser.Close()
			if err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close stream"))
			}
		}
	}()
	if err != nil {
		return err
	}

	return readerHandler(csvReader)
}

func CsvWriter(writer io.WriteCloser, writerHandler func(writer *csv.Writer) error, opts ...CsvOption) error {
	var (
		bufWriter  *bufio.Writer
		gzipWriter *gzip.Writer
		csvWriter  *csv.Writer
	)

	opt := newCsvOptions()
	for _, op := range opts {
		op(opt)
	}

	bufWriter = bufio.NewWriterSize(writer, bufSize)

	if opt.compressed {
		gzipWriter = gzip.NewWriter(bufWriter)
		csvWriter = csv.NewWriter(gzipWriter)
	} else {
		csvWriter = csv.NewWriter(bufWriter)
	}
	csvWriter.Comma = opt.csvSep

	defer func() {
		if csvWriter != nil {
			csvWriter.Flush()
			if err := csvWriter.Error(); err != nil {
				opt.closeErrorHandler(errors.WithMessage(err, "close csv writer"))
			}
		}
		if gzipWriter != nil && opt.compressed {
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
		if err := writer.Close(); err != nil {
			opt.closeErrorHandler(errors.WithMessage(err, "close stream"))
		}
	}()

	return writerHandler(csvWriter)
}

func newCsvOptions() *csvOpts {
	return &csvOpts{
		closeErrorHandler: func(err error) {
		},
		csvSep:     ';',
		compressed: true,
	}
}

func makeReaders(readCloser io.ReadCloser, opts csvOpts) (*gzip.Reader, *csv.Reader, error) {
	if opts.compressed {
		gzipReader, err := gzip.NewReader(readCloser)
		if err != nil {
			_ = readCloser.Close()
			return nil, nil, errors.WithMessage(err, "open gzip reader")
		}

		csvReader := csv.NewReader(gzipReader)
		csvReader.Comma = opts.csvSep

		return gzipReader, csvReader, nil
	} else {
		csvReader := csv.NewReader(readCloser)
		csvReader.Comma = opts.csvSep

		return nil, csvReader, nil
	}
}
