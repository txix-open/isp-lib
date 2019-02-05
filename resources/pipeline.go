package resources

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
)

var (
	noop = func(err error) {}
)

type Pipe interface {
	Map(src chan interface{}) chan interface{}
}

type Result struct {
	Errors      map[int]string `json:"omitempty"`
	Lines       int
	ErrorsCount int
	FileName    string
}

type LineScanner struct {
	filePattern string //Example: /var/log/log_%d
	maxBuffSize int
	skipLines   int
	onError     func(err error)
	makeReader  func(file *os.File) (io.Reader, error)
	ch          chan interface{}
}

// iterate from 0 to n, replace index in 'filePattern' and tries to read file
// it closes channel 'ch' and return if file is not exists
func (s *LineScanner) Run() []Result {
	files := make([]Result, 0)
	totalLines := 0

	for i := 0; true; i++ {
		fileName := fmt.Sprintf(s.filePattern, i)
		file, err := os.Open(fileName)
		if err == os.ErrNotExist {
			close(s.ch)
			break
		} else if err != nil {
			s.onError(err)
			continue
		}

		reader, err := s.makeReader(file)
		if err != nil {
			_ = file.Close()
			s.onError(err)
			continue
		}

		result := Result{FileName: fileName, Errors: map[int]string{}}
		buf := make([]byte, s.maxBuffSize/4)
		scanner := bufio.NewScanner(reader)
		scanner.Buffer(buf, s.maxBuffSize)
		for scanner.Scan() {
			bytes := scanner.Bytes()
			if len(bytes) == 0 {
				continue
			}

			totalLines++
			result.Lines++

			if totalLines > s.skipLines {
				if scanner.Err() != nil {
					result.ErrorsCount++
					result.Errors[result.Lines] = scanner.Err().Error()
				} else {
					c := make([]byte, len(bytes))
					copy(c, bytes)
					s.ch <- c
				}
			}
		}

		files = append(files, result)
		_ = file.Close()

	}

	return files
}

func (s *LineScanner) OnError(f func(err error)) *LineScanner {
	s.onError = f
	return s
}

func (s *LineScanner) Skip(lineCount int) *LineScanner {
	s.skipLines = lineCount
	return s
}

func (s *LineScanner) Reader(makeReader func(f *os.File) (io.Reader, error)) *LineScanner {
	s.makeReader = makeReader
	return s
}

func (s *LineScanner) Lines() chan interface{} {
	return s.ch
}

func NewLineScanner(filePattern string, maxBuffSize int) *LineScanner {
	return &LineScanner{
		filePattern: filePattern,
		maxBuffSize: maxBuffSize,
		ch:          make(chan interface{}),
		onError:     noop,
		skipLines:   0,
		makeReader: func(file *os.File) (io.Reader, error) {
			return file, nil
		},
	}
}

type JsonUnmarshaler struct {
	makePtr    func() interface{}
	unmarshal  func(bytes []byte, ptr interface{}) error
	onError    func(err error)
	goroutines int
	closer     sync.Once
}

func (u *JsonUnmarshaler) Map(src chan interface{}) chan interface{} {
	ch := make(chan interface{})

	for i := 0; i < u.goroutines; i++ {
		go func() {
			for val := range src {
				if val == nil {
					continue
				}

				if bytes, ok := val.([]byte); ok {
					ptr := u.makePtr()
					if err := u.unmarshal(bytes, ptr); err != nil {
						u.onError(err)
					} else {
						ch <- ptr
					}
				} else {
					u.onError(fmt.Errorf("jsonUnmarshaler: expecting []byte, got: %v", reflect.TypeOf(val).String()))
				}
			}

			u.closer.Do(func() {
				close(ch)
			})
		}()
	}

	return ch
}

func (u *JsonUnmarshaler) OnError(f func(err error)) *JsonUnmarshaler {
	u.onError = f
	return u
}

func (u *JsonUnmarshaler) Goroutines(count int) *JsonUnmarshaler {
	u.goroutines = count
	return u
}

func (u *JsonUnmarshaler) Unmarshal(unmarshal func(bytes []byte, ptr interface{}) error) *JsonUnmarshaler {
	u.unmarshal = unmarshal
	return u
}

func NewJsonUnmarshaler(makePtr func() interface{}) *JsonUnmarshaler {
	return &JsonUnmarshaler{
		makePtr:    makePtr,
		unmarshal:  json.Unmarshal,
		goroutines: 1,
		onError:    noop,
	}
}

type Batcher struct {
	batchSize int
}

func (b *Batcher) Map(src chan interface{}) chan interface{} {
	ch := make(chan interface{})

	go func() {
		currentSize := 0
		batch := make([]interface{}, b.batchSize)

		for val := range src {
			batch[currentSize] = val
			currentSize++

			if currentSize%b.batchSize == 0 {
				ch <- batch
				currentSize = 0
				batch = make([]interface{}, b.batchSize)
			}
		}

		if currentSize > 0 {
			ch <- batch[0:currentSize]
		}

		close(ch)
	}()

	return ch
}

func NewBatcher(batchSize int) *Batcher {
	return &Batcher{batchSize: batchSize}
}

func Pipeline(src chan interface{}, pipes ...Pipe) chan interface{} {
	for _, p := range pipes {
		src = p.Map(src)
	}
	return src
}
