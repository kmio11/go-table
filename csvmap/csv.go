package csvmap

import (
	"encoding/csv"
	"io"

	"github.com/kmio11/tablemap"
)

// Reader is a CSV reader that can unmarshal data into structs.
type Reader[T any] struct {
	R       *csv.Reader
	opts    *tablemap.Options
	handler *tablemap.RowHandler[T]
}

// NewReader creates a new Reader with optional tablemap.Options.
func NewReader[T any](r io.Reader, opts *tablemap.Options) *Reader[T] {
	return &Reader[T]{
		R:    csv.NewReader(r),
		opts: opts,
	}
}

// Read reads one record and converts it to struct T.
// The first call to Read will read the header row.
func (r *Reader[T]) Read() (*T, error) {
	// Read header on first read
	if r.handler == nil {
		header, err := r.R.Read()
		if err != nil {
			return nil, err
		}

		handler, err := tablemap.NewRowHandler[T](header, r.opts)
		if err != nil {
			return nil, err
		}
		r.handler = handler
	}

	// Read data row
	row, err := r.R.Read()
	if err != nil {
		return nil, err
	}

	return r.handler.UnmarshalRow(row)
}

// ReadAll reads all records from CSV and converts them to a slice of struct T.
func (r *Reader[T]) ReadAll() ([]T, error) {
	var result []T

	records, err := r.R.ReadAll()
	if err != nil {
		return nil, err
	}
	if err := tablemap.UnmarshalWithOptions(records[0], records[1:], &result, r.opts); err != nil {
		return nil, err
	}

	return result, nil
}

// Writer is a CSV writer that can marshal structs into CSV format.
type Writer[T any] struct {
	W       *csv.Writer
	opts    *tablemap.Options
	handler *tablemap.RowHandler[T]
}

// NewWriter creates a new Writer with optional tablemap.Options.
func NewWriter[T any](w io.Writer, opts *tablemap.Options) *Writer[T] {
	return &Writer[T]{
		W:    csv.NewWriter(w),
		opts: opts,
	}
}

// Write writes a single record to CSV.
// The first call to Write will write the header row.
func (w *Writer[T]) Write(data T) error {
	// Initialize handler and write header on first write
	if w.handler == nil {
		var zero T
		header, _, err := tablemap.MarshalWithOptions([]T{zero}, w.opts)
		if err != nil {
			return err
		}

		handler, err := tablemap.NewRowHandler[T](header, w.opts)
		if err != nil {
			return err
		}
		w.handler = handler

		if err := w.W.Write(header); err != nil {
			return err
		}
	}

	// Write data row
	row, err := w.handler.MarshalRow(&data)
	if err != nil {
		return err
	}

	if err := w.W.Write(row); err != nil {
		return err
	}

	return nil
}

// WriteAll writes a slice of struct T as CSV data.
func (w *Writer[T]) WriteAll(data []T) error {
	defer w.W.Flush()
	header, rows, err := tablemap.MarshalWithOptions(data, w.opts)
	if err != nil {
		return err
	}
	return w.W.WriteAll(append([][]string{header}, rows...))
}
