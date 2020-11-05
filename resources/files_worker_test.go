package resources

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name          string
	compressed    bool
	separator     rune
	valuesToWrite [][]string
	want          string
}

type testCaseR struct {
	name         string
	compressed   bool
	separator    rune
	valuesToRead string
	want         [][]string
}

type testWriteCloser struct {
	io.Writer
}

func (tWC testWriteCloser) Close() error {
	return nil
}

type testReadCloser struct {
	io.Reader
}

func (tRC testReadCloser) Close() error {
	return nil
}

func compress(data string) string {
	var res []byte
	buffer := bytes.NewBuffer(res)
	gWriter := gzip.NewWriter(buffer)
	if _, err := gWriter.Write([]byte(data)); err != nil {
		return ""
	}

	if gWriter.Flush() != nil {
		return ""
	}
	if gWriter.Close() != nil {
		return ""
	}

	return buffer.String()
}

func TestCheckReader_Read(t *testing.T) {
	testCases := []testCaseR{
		{
			name:         "csv/1",
			separator:    ';',
			want:         [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "80", "human"}},
			valuesToRead: "name;age;race\nivan;13;human\npetya;80;human\n",
		},
		{
			name:         "csv/2",
			separator:    '@',
			want:         [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "80", "human"}},
			valuesToRead: "name@age@race\nivan@13@human\npetya@80@human\n",
		},
		{
			name:         "csv/3",
			separator:    ';',
			want:         [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			valuesToRead: "name;age;race\nivan;13;human\npetya;;human\n",
		},
		{
			name:         "zip/1",
			compressed:   true,
			separator:    ';',
			want:         [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "80", "human"}},
			valuesToRead: compress("name;age;race\nivan;13;human\npetya;80;human\n"),
		},
		{
			name:         "zip/2",
			compressed:   true,
			separator:    ';',
			want:         [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			valuesToRead: compress("name;age;race\nivan;13;human\npetya;;human\n"),
		},
		{
			name:         "zip/3",
			compressed:   true,
			separator:    '@',
			want:         [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			valuesToRead: compress("name@age@race\nivan@13@human\npetya@@human\n"),
		},
	}

	for _, value := range testCases {
		reader := testReadCloser{strings.NewReader(value.valuesToRead)}
		var records [][]string
		callback := func(reader *csv.Reader) (err error) {
			records, err = reader.ReadAll()
			if err != nil {
				t.Errorf("reading error: %v", err)
			}

			return
		}
		if err := CsvReader(&reader, callback, WithSeparator(value.separator), WithCompression(value.compressed)); err != nil {
			t.Errorf("reading error: %v", err)
		}

		assert.EqualValuesf(t, value.want, records, "want %s, but have %s at case %s", value.want, records, value.name)
	}
}

func TestCheckWriter_Write(t *testing.T) {
	testCases := []testCase{
		{
			name:          "csv/1",
			separator:     ';',
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "80", "human"}},
			want:          "name;age;race\nivan;13;human\npetya;80;human\n",
		},
		{
			name:          "csv/2",
			separator:     ';',
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name;age;race\nivan;13;human\npetya;;human\n",
		},
		{
			name:          "csv/3",
			separator:     '=',
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name=age=race\nivan=13=human\npetya==human\n",
		},
		{
			name:          "zip/1",
			separator:     ';',
			compressed:    true,
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name;age;race\nivan;13;human\npetya;;human\n",
		},
		{
			name:          "zip/2",
			separator:     '^',
			compressed:    true,
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name^age^race\nivan^13^human\npetya^^human\n",
		},
	}

	for _, value := range testCases {
		buffer := bytes.NewBuffer([]byte{})
		writer := testWriteCloser{buffer}

		callback := func(writer *csv.Writer) error {
			for _, data := range value.valuesToWrite {
				if err := writer.Write(data); err != nil {
					t.Errorf("write error: %v", err)
				}

				writer.Flush()
			}

			return nil
		}

		if err := CsvWriter(&writer, callback, WithCompression(value.compressed), WithSeparator(value.separator)); err != nil {
			t.Errorf("write error: %v", err)
		}

		if value.compressed {
			result := buffer.String()
			reader := strings.NewReader(result)
			gReader, err := gzip.NewReader(reader)
			if err != nil {
				t.Errorf("unexpected error (gzip): %v", err)
			}

			res, err := ioutil.ReadAll(gReader)
			if err != nil {
				t.Errorf("unexpected error (reading): %v", err)
			}

			assert.EqualValuesf(t, value.want, string(res), "want %s, but have %s at case %s", value.want, buffer.String(), value.name)
		} else {
			assert.EqualValuesf(t, value.want, buffer.String(), "want %s, but have %s at case %s", value.want, buffer.String(), value.name)
		}
	}
}

func TestCheckWriter_WriteAll(t *testing.T) {
	testCases := []testCase{
		{
			name:          "csv/1",
			separator:     ';',
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "80", "human"}},
			want:          "name;age;race\nivan;13;human\npetya;80;human\n",
		},
		{
			name:          "csv/2",
			separator:     ';',
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name;age;race\nivan;13;human\npetya;;human\n",
		},
		{
			name:          "csv/3",
			separator:     '=',
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name=age=race\nivan=13=human\npetya==human\n",
		},
		{
			name:          "zip/1",
			separator:     ';',
			compressed:    true,
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name;age;race\nivan;13;human\npetya;;human\n",
		},
		{
			name:          "zip/2",
			separator:     '^',
			compressed:    true,
			valuesToWrite: [][]string{{"name", "age", "race"}, {"ivan", "13", "human"}, {"petya", "", "human"}},
			want:          "name^age^race\nivan^13^human\npetya^^human\n",
		},
	}

	for _, value := range testCases {
		buffer := bytes.NewBuffer([]byte{})
		writer := testWriteCloser{buffer}

		callback := func(writer *csv.Writer) error {
			writer.WriteAll(value.valuesToWrite)
			writer.Flush()
			return nil
		}

		if err := CsvWriter(&writer, callback, WithCompression(value.compressed), WithSeparator(value.separator)); err != nil {
			t.Errorf("write error: %v", err)
		}

		if value.compressed {
			result := buffer.String()
			reader := strings.NewReader(result)
			gReader, err := gzip.NewReader(reader)
			if err != nil {
				t.Errorf("unexpected error (gzip): %v", err)
			}

			res, err := ioutil.ReadAll(gReader)
			if err != nil {
				t.Errorf("unexpected error (reading): %v", err)
			}

			assert.EqualValuesf(t, value.want, string(res), "want %s, but have %s at case %s", value.want, buffer.String(), value.name)
		} else {
			assert.EqualValuesf(t, value.want, buffer.String(), "want %s, but have %s at case %s", value.want, buffer.String(), value.name)
		}
	}
}
