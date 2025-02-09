package table

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
)

// TableMarshaler is the interface implemented by types that
// can marshal themselves into a table cell string representation.
type TableMarshaler interface {
	MarshalTable() (string, error)
}

// TableUnmarshaler is the interface implemented by types that
// can unmarshal a table cell string representation of themselves.
type TableUnmarshaler interface {
	UnmarshalTable(string) error
}

// Options defines configuration options for marshaling and unmarshaling.
type Options struct {
	// NilValue is the string representation of nil values.
	// Default is "\N".
	NilValue string
}

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		NilValue: "\\N",
	}
}

const (
	tagTable = "table"
	ignore   = "-"
)

// Unmarshal converts table data into a slice of structs using default options.
func Unmarshal(header []string, data [][]string, v any) error {
	return UnmarshalWithOptions(header, data, v, DefaultOptions())
}

// UnmarshalWithOptions converts table data into a slice of structs with custom options.
func UnmarshalWithOptions(header []string, data [][]string, v any, opts *Options) error {
	if opts == nil {
		opts = DefaultOptions()
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer to a slice")
	}

	sliceVal := rv.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("v must be a pointer to a slice")
	}

	// Get the type of elements in the slice
	sliceElemType := sliceVal.Type().Elem()
	if sliceElemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs")
	}

	// Get field mapping including embedded fields
	fields := getFieldMap(sliceElemType).fields

	// Process each row
	for _, row := range data {
		if len(row) != len(header) {
			return fmt.Errorf("inconsistent data length")
		}

		// Create new struct
		newStruct := reflect.New(sliceElemType).Elem()

		// Fill the struct fields
		for i, col := range row {
			if info, ok := fields[header[i]]; ok {
				// Navigate to the field through the embedded structs
				field := newStruct
				for _, idx := range info.index {
					field = field.Field(idx)
				}
				if err := setField(field, col, opts); err != nil {
					return fmt.Errorf("setting field %s: %v", header[i], err)
				}
			}
		}

		sliceVal.Set(reflect.Append(sliceVal, newStruct))
	}

	return nil
}

// Marshal converts a slice of structs into table data using default options.
func Marshal(v any) ([]string, [][]string, error) {
	return MarshalWithOptions(v, DefaultOptions())
}

// MarshalWithOptions converts a slice of structs into table data with custom options.
func MarshalWithOptions(v any, opts *Options) ([]string, [][]string, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil, nil, fmt.Errorf("v must be a slice")
	}

	if rv.Len() == 0 {
		return nil, nil, nil
	}

	// Get the type of elements in the slice
	elemType := rv.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("slice elements must be structs")
	}

	// Get field mapping including embedded fields and ordered tags
	fm := getFieldMap(elemType)
	fields, orderedTags := fm.fields, fm.orderedTags

	// Create data rows
	data := make([][]string, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		row := make([]string, len(orderedTags))
		item := rv.Index(i)

		for j, tag := range orderedTags {
			info := fields[tag]
			// Navigate to the field through the embedded structs
			field := item
			for _, idx := range info.index {
				field = field.Field(idx)
			}
			row[j] = formatField(field, opts)
		}

		data[i] = row
	}

	return orderedTags, data, nil
}

// fieldInfo stores information about a struct field including its path through embedded structs
type fieldInfo struct {
	index    []int
	tag      string
	position int // Field position to maintain declaration order
}

// fieldMap contains the result of field mapping
type fieldMap struct {
	fields      map[string]fieldInfo
	orderedTags []string
}

// getFieldMap creates a map of tag names to field paths and maintains declaration order
func getFieldMap(t reflect.Type) fieldMap {
	result := fieldMap{
		fields:      make(map[string]fieldInfo),
		orderedTags: make([]string, 0),
	}

	pos := 0

	var addFields func(t reflect.Type, index []int, isEmbedded bool)
	addFields = func(t reflect.Type, index []int, isEmbedded bool) {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			currIndex := append(index, i)

			// Handle embedded struct
			if field.Anonymous && field.Type.Kind() == reflect.Struct {
				addFields(field.Type, currIndex, true)
				continue
			}

			// Skip fields without table tag
			tag := field.Tag.Get(tagTable)
			if tag == "" || tag == ignore {
				continue
			}

			// For embedded fields, skip if tag already exists
			if isEmbedded && result.hasTag(tag) {
				continue
			}

			// Update field info
			result.fields[tag] = fieldInfo{
				index:    currIndex,
				tag:      tag,
				position: pos,
			}

			// Update orderedTags
			if existingIdx := result.findTagIndex(tag); existingIdx >= 0 {
				// Remove existing tag if being overwritten by non-embedded field
				result.orderedTags = append(result.orderedTags[:existingIdx], result.orderedTags[existingIdx+1:]...)
			}
			result.orderedTags = append(result.orderedTags, tag)
			pos++
		}
	}

	addFields(t, nil, false)
	return result
}

// findTagIndex returns the index of the tag in orderedTags, or -1 if not found
func (fm *fieldMap) findTagIndex(tag string) int {
	for i, t := range fm.orderedTags {
		if t == tag {
			return i
		}
	}
	return -1
}

// hasTag checks if a tag already exists in orderedTags
func (fm *fieldMap) hasTag(tag string) bool {
	for _, t := range fm.orderedTags {
		if t == tag {
			return true
		}
	}
	return false
}

// setField sets the value of a struct field from a string with custom options
func setField(field reflect.Value, value string, opts *Options) error {
	// Handle nil value
	if value == opts.NilValue {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		// Non-pointer fields cannot be nil
		return fmt.Errorf("cannot set nil to non-pointer field of type: %v", field.Type())
	}

	// Handle pointer types
	if field.Kind() == reflect.Ptr {
		if value == "" {
			field.Set(reflect.Zero(field.Type()))
			return nil
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setField(field.Elem(), value, opts)
	}

	// 1. Check for TableUnmarshaler
	if field.CanAddr() {
		if tu, ok := field.Addr().Interface().(TableUnmarshaler); ok {
			return tu.UnmarshalTable(value)
		}
	}

	// 2. Check for encoding.TextUnmarshaler
	if field.CanAddr() {
		if tu, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return tu.UnmarshalText([]byte(value))
		}
	}

	// 3. Built-in type conversions
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}

// formatField converts a struct field to string
func formatField(field reflect.Value, opts *Options) string {
	// Handle pointer types
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return opts.NilValue
		}
		return formatField(field.Elem(), opts)
	}

	// 1. Check for TableMarshaler
	if field.CanAddr() {
		if tm, ok := field.Addr().Interface().(TableMarshaler); ok {
			str, err := tm.MarshalTable()
			if err == nil {
				return str
			}
			// Fall through on error
		}
	}

	// 2. Check for encoding.TextMarshaler
	if field.CanAddr() {
		if tm, ok := field.Addr().Interface().(encoding.TextMarshaler); ok {
			bytes, err := tm.MarshalText()
			if err == nil {
				return string(bytes)
			}
			// Fall through on error
		}
	}

	// 3. Built-in type conversions
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(field.Bool())
	default:
		return fmt.Sprintf("%v", field.Interface())
	}
}
