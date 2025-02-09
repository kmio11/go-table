package table_test

import (
	"testing"
	"time"

	"github.com/kmio11/go-table"

	"github.com/stretchr/testify/assert"
)

// CustomType implements TableMarshaller and TableUnmarshaller
type CustomType struct {
	value string
}

func (c *CustomType) MarshalTable() (string, error) {
	return "custom:" + c.value, nil
}

func (c *CustomType) UnmarshalTable(s string) error {
	if len(s) > 7 && s[0:7] == "custom:" {
		c.value = s[7:]
	}
	return nil
}

// TimeWrapper implements encoding.TextMarshaler and TextUnmarshaler
type TimeWrapper struct {
	Time time.Time
}

func (t *TimeWrapper) MarshalText() ([]byte, error) {
	return []byte(t.Time.Format(time.RFC3339)), nil
}

func (t *TimeWrapper) UnmarshalText(text []byte) error {
	parsed, err := time.Parse(time.RFC3339, string(text))
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

type TestStruct struct {
	String    string      `table:"string"`
	Int       int         `table:"int"`
	Bool      bool        `table:"bool"`
	CustomPtr *CustomType `table:"custom_ptr"`
	Custom    CustomType  `table:"custom"`
	Time      TimeWrapper `table:"time"`
	IntPtr    *int        `table:"int_ptr"`
	Ignored   string
}

func TestMarshal(t *testing.T) {
	intVal := 42
	testData := []TestStruct{
		{
			String:    "test",
			Int:       123,
			Bool:      true,
			CustomPtr: &CustomType{value: "hello"},
			Custom:    CustomType{value: "world"},
			Time: TimeWrapper{
				Time: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			IntPtr:  &intVal,
			Ignored: "ignored",
		},
	}

	// Test Marshal
	header, data, err := table.Marshal(testData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Verify header
	expectedHeader := []string{"string", "int", "bool", "custom_ptr", "custom", "time", "int_ptr"}
	if len(header) != len(expectedHeader) {
		t.Errorf("Expected header length %d, got %d", len(expectedHeader), len(header))
	}
	for i, h := range expectedHeader {
		if header[i] != h {
			t.Errorf("Expected header[%d]=%s, got %s", i, h, header[i])
		}
	}

	// Verify data
	expectedData := [][]string{
		{"test", "123", "true", "custom:hello", "custom:world", "2024-01-01T00:00:00Z", "42"},
	}
	if len(data) != len(expectedData) {
		t.Fatalf("Expected %d rows, got %d", len(expectedData), len(data))
	}
	for i, row := range expectedData {
		for j, val := range row {
			if data[i][j] != val {
				t.Errorf("Data[%d][%d]: expected %s, got %s", i, j, val, data[i][j])
			}
		}
	}
}

func TestUnmarshal(t *testing.T) {
	header := []string{"string", "int", "bool", "custom_ptr", "custom", "time", "int_ptr"}
	data := [][]string{
		{"test", "123", "true", "custom:hello", "custom:world", "2024-01-01T00:00:00Z", "42"},
	}

	var result []TestStruct
	err := table.Unmarshal(header, data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	decoded := result[0]
	// Verify each field
	if decoded.String != "test" {
		t.Errorf("String: expected %s, got %s", "test", decoded.String)
	}
	if decoded.Int != 123 {
		t.Errorf("Int: expected %d, got %d", 123, decoded.Int)
	}
	if !decoded.Bool {
		t.Errorf("Bool: expected true, got false")
	}
	if decoded.CustomPtr == nil || decoded.CustomPtr.value != "hello" {
		t.Errorf("Custom: expected 'hello', got %v", decoded.CustomPtr)
	}
	expectedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !decoded.Time.Time.Equal(expectedTime) {
		t.Errorf("Time: expected %v, got %v", expectedTime, decoded.Time.Time)
	}
	if decoded.IntPtr == nil || *decoded.IntPtr != 42 {
		t.Errorf("IntPtr: expected 42, got %v", decoded.IntPtr)
	}
}

func TestUnmarshal_nilValue(t *testing.T) {
	testData := []TestStruct{
		{
			String:    "test",
			CustomPtr: nil,
			IntPtr:    nil,
		},
	}

	header, data, err := table.Marshal(testData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result []TestStruct
	err = table.Unmarshal(header, data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result[0].CustomPtr != nil {
		t.Error("Expected nil Custom field")
	}
	if result[0].IntPtr != nil {
		t.Error("Expected nil IntPtr field")
	}
}

type nilValueTestStruct struct {
	IntPtr    *int       `table:"int"`
	StringPtr *string    `table:"str"`
	TimePtr   *time.Time `table:"time"`
	Normal    string     `table:"normal"`
}

func TestMarshalWithOptions_nilValue(t *testing.T) {
	str := "hello"
	now := time.Now().Round(time.Second) // Round to seconds to avoid precision issues
	num := 42

	tests := []struct {
		name     string
		input    []nilValueTestStruct
		options  *table.Options
		expected [][]string
	}{
		{
			name: "default nil value",
			input: []nilValueTestStruct{
				{IntPtr: nil, StringPtr: &str, TimePtr: &now, Normal: "test1"},
				{IntPtr: &num, StringPtr: nil, TimePtr: nil, Normal: "test2"},
			},
			options: nil,
			expected: [][]string{
				{"\\N", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "\\N", "\\N", "test2"},
			},
		},
		{
			name: "custom nil value",
			input: []nilValueTestStruct{
				{IntPtr: nil, StringPtr: &str, TimePtr: &now, Normal: "test1"},
				{IntPtr: &num, StringPtr: nil, TimePtr: nil, Normal: "test2"},
			},
			options: &table.Options{NilValue: "NULL"},
			expected: [][]string{
				{"NULL", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "NULL", "NULL", "test2"},
			},
		},
		{
			name: "empty string as nil value",
			input: []nilValueTestStruct{
				{IntPtr: nil, StringPtr: &str, TimePtr: &now, Normal: "test1"},
				{IntPtr: &num, StringPtr: nil, TimePtr: nil, Normal: "test2"},
			},
			options: &table.Options{NilValue: ""},
			expected: [][]string{
				{"", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "", "", "test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, data, err := table.MarshalWithOptions(tt.input, tt.options)
			assert.NoError(t, err)
			assert.Equal(t, []string{"int", "str", "time", "normal"}, header)

			// Compare time fields separately to handle potential precision differences
			for i, row := range data {
				for j, val := range row {
					if j == 2 && // time field
						(tt.options == nil && val != "\\N") &&
						(tt.options != nil && val != tt.options.NilValue) {
						expectedTime, _ := time.Parse(time.RFC3339, tt.expected[i][j])
						actualTime, err := time.Parse(time.RFC3339, val)
						assert.NoError(t, err)
						assert.True(t, expectedTime.Equal(actualTime), "Times should be equal")
					} else {
						assert.Equal(t, tt.expected[i][j], val)
					}
				}
			}
		})
	}
}

func TestUnmarshalWithOptions_nilValue(t *testing.T) {
	now := time.Now().Round(time.Second) // Round to seconds to avoid precision issues
	str := "hello"
	num := 42

	tests := []struct {
		name     string
		header   []string
		data     [][]string
		options  *table.Options
		expected []nilValueTestStruct
	}{
		{
			name:   "default nil value",
			header: []string{"int", "str", "time", "normal"},
			data: [][]string{
				{"\\N", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "\\N", "\\N", "test2"},
			},
			options: nil,
			expected: []nilValueTestStruct{
				{
					StringPtr: &str,
					TimePtr:   &now,
					Normal:    "test1",
				},
				{
					IntPtr: &num,
					Normal: "test2",
				},
			},
		},
		{
			name:   "custom nil value",
			header: []string{"int", "str", "time", "normal"},
			data: [][]string{
				{"NULL", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "NULL", "NULL", "test2"},
			},
			options: &table.Options{NilValue: "NULL"},
			expected: []nilValueTestStruct{
				{
					StringPtr: &str,
					TimePtr:   &now,
					Normal:    "test1",
				},
				{
					IntPtr: &num,
					Normal: "test2",
				},
			},
		},
		{
			name:   "empty string as nil value",
			header: []string{"int", "str", "time", "normal"},
			data: [][]string{
				{"", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "", "", "test2"},
			},
			options: &table.Options{NilValue: ""},
			expected: []nilValueTestStruct{
				{
					StringPtr: &str,
					TimePtr:   &now,
					Normal:    "test1",
				},
				{
					IntPtr: &num,
					Normal: "test2",
				},
			},
		},
		{
			name:     "error case - nil value for non-pointer field",
			header:   []string{"normal"},
			data:     [][]string{{"\\N"}},
			options:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result []nilValueTestStruct
			err := table.UnmarshalWithOptions(tt.header, tt.data, &result, tt.options)

			if tt.name == "error case - nil value for non-pointer field" {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result))

			for i := range tt.expected {
				if tt.expected[i].IntPtr != nil {
					assert.Equal(t, *tt.expected[i].IntPtr, *result[i].IntPtr)
				} else {
					assert.Nil(t, result[i].IntPtr)
				}
				if tt.expected[i].StringPtr != nil {
					assert.Equal(t, *tt.expected[i].StringPtr, *result[i].StringPtr)
				} else {
					assert.Nil(t, result[i].StringPtr)
				}
				if tt.expected[i].TimePtr != nil {
					assert.True(t, tt.expected[i].TimePtr.Equal(*result[i].TimePtr), "Times should be equal")
				} else {
					assert.Nil(t, result[i].TimePtr)
				}
				assert.Equal(t, tt.expected[i].Normal, result[i].Normal)
			}
		})
	}
}
