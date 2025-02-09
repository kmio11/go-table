package csvmap

import (
	"encoding/csv"
	"io"

	"github.com/kmio11/tablemap"
)

// Reader is a CSV reader that can unmarshal data into structs.
type Reader struct {
	R    *csv.Reader
	opts *tablemap.Options
}

// NewReader creates a new Reader with optional tablemap.Options.
func NewReader(r io.Reader, opts *tablemap.Options) *Reader {
	return &Reader{
		R:    csv.NewReader(r),
		opts: opts,
	}
}

// ReadAll reads all records from CSV and converts them to a slice of struct T.
func ReadAll[T any](r *Reader) ([]T, error) {
	// Read header
	header, err := r.R.Read()
	if err != nil {
		return nil, err
	}

	// Read data
	var rows [][]string
	for {
		row, err := r.R.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}

	// Convert to struct slice
	var result []T
	if err := tablemap.UnmarshalWithOptions(header, rows, &result, r.opts); err != nil {
		return nil, err
	}

	return result, nil
}

// Writer is a CSV writer that can marshal structs into CSV format.
type Writer struct {
	W    *csv.Writer
	opts *tablemap.Options
}

// NewWriter creates a new Writer with optional tablemap.Options.
func NewWriter(w io.Writer, opts *tablemap.Options) *Writer {
	return &Writer{
		W:    csv.NewWriter(w),
		opts: opts,
	}
}

// WriteAll writes a slice of struct T as CSV data.
func WriteAll[T any](w *Writer, data []T) error {
	defer w.W.Flush()

	// Convert struct slice to table format
	var header []string
	var rows [][]string
	var err error

	header, rows, err = tablemap.MarshalWithOptions(data, w.opts)
	if err != nil {
		return err
	}

	// Write header
	if err := w.W.Write(header); err != nil {
		return err
	}

	// Write data rows
	for _, row := range rows {
		if err := w.W.Write(row); err != nil {
			return err
		}
	}

	return nil
}
