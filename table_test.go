package tablemap_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/kmio11/tablemap"
	"github.com/stretchr/testify/assert"
)

// CustomType implements TableMarshaller and CellUnmarshaler
type CustomType struct {
	value string
}

func (c *CustomType) MarshalCell() (string, error) {
	return "custom:" + c.value, nil
}

func (c *CustomType) UnmarshalCell(s string) error {
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
	header, data, err := tablemap.Marshal(testData)
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
	err := tablemap.Unmarshal(header, data, &result)
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

	header, data, err := tablemap.Marshal(testData)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var result []TestStruct
	err = tablemap.Unmarshal(header, data, &result)
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
		options  *tablemap.Options
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
			options: &tablemap.Options{NilValue: "NULL"},
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
			options: &tablemap.Options{NilValue: ""},
			expected: [][]string{
				{"", "hello", now.Format(time.RFC3339), "test1"},
				{"42", "", "", "test2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, data, err := tablemap.MarshalWithOptions(tt.input, tt.options)
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
		options  *tablemap.Options
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
			options: &tablemap.Options{NilValue: "NULL"},
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
			options: &tablemap.Options{NilValue: ""},
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
			err := tablemap.UnmarshalWithOptions(tt.header, tt.data, &result, tt.options)

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

type EmbeddedAddress struct {
	Street string `table:"street"`
	City   string `table:"city"`
}

type PersonWithAddress struct {
	Name string `table:"name"`
	Age  int    `table:"age"`
	EmbeddedAddress
}

type ConflictAddress struct {
	Street string `table:"addr"` // conflicting tag
}

type PersonWithConflict struct {
	Name string `table:"name"`
	ConflictAddress
	Addr string `table:"addr"` // same tag as embedded field
}

func TestMarshal_embedded(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedHeader []string
		expectedData   [][]string
		wantErr        bool
	}{
		{
			name: "basic embedding",
			input: []PersonWithAddress{
				{
					Name: "John",
					Age:  30,
					EmbeddedAddress: EmbeddedAddress{
						Street: "123 Main St",
						City:   "Springfield",
					},
				},
			},
			expectedHeader: []string{"name", "age", "street", "city"},
			expectedData: [][]string{
				{"John", "30", "123 Main St", "Springfield"},
			},
			wantErr: false,
		},
		{
			name: "tag conflict resolution",
			input: []PersonWithConflict{
				{
					Name: "John",
					Addr: "Primary Address", // This should win
					ConflictAddress: ConflictAddress{
						Street: "Should not appear",
					},
				},
			},
			expectedHeader: []string{"name", "addr"},
			expectedData: [][]string{
				{"John", "Primary Address"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, data, err := tablemap.Marshal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.expectedHeader, header)
				assert.Equal(t, tt.expectedData, data)
			}
		})
	}
}

func TestUnmarshal_embedded(t *testing.T) {
	tests := []struct {
		name    string
		header  []string
		data    [][]string
		want    interface{}
		wantErr bool
	}{
		{
			name:   "basic embedding",
			header: []string{"name", "age", "street", "city"},
			data: [][]string{
				{"John", "30", "123 Main St", "Springfield"},
			},
			want: []PersonWithAddress{
				{
					Name: "John",
					Age:  30,
					EmbeddedAddress: EmbeddedAddress{
						Street: "123 Main St",
						City:   "Springfield",
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "tag conflict resolution",
			header: []string{"name", "addr"},
			data: [][]string{
				{"John", "Primary Address"},
			},
			want: []PersonWithConflict{
				{
					Name: "John",
					Addr: "Primary Address", // This should be set
					ConflictAddress: ConflictAddress{
						Street: "", // This should remain empty
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}
			switch v := tt.want.(type) {
			case []PersonWithAddress:
				var result []PersonWithAddress
				got = &result
			case []PersonWithConflict:
				var result []PersonWithConflict
				got = &result
			default:
				t.Fatalf("Unexpected type: %T", v)
			}

			err := tablemap.Unmarshal(tt.header, tt.data, got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, tt.want, reflect.ValueOf(got).Elem().Interface())
			}
		})
	}
}

func TestMarshal_headerOrder(t *testing.T) {
	type Inner struct {
		B string `table:"b"`
		C string `table:"c"`
	}

	type DeepInner struct {
		X string `table:"x"`
		Y string `table:"y"`
	}

	type DeepOuter struct {
		DeepInner
		Z string `table:"z"`
	}

	type Outer struct {
		A string `table:"a"`
		Inner
		D string `table:"d"`
		E string `table:"e"`
	}

	type Override struct {
		Inner
		B string `table:"b"` // Inner.B should be overridden
		F string `table:"f"`
	}

	type MultipleEmbedded struct {
		Inner
		DeepOuter
		M string `table:"m"`
	}

	type Duplicate struct {
		A1 string `table:"same"`
		A2 string `table:"same"` // Duplicate tag
	}

	tests := []struct {
		name           string
		input          interface{}
		expectedHeader []string
		expectedData   [][]string
		wantErr        bool
	}{
		{
			name: "maintain declaration order with embedded struct",
			input: []Outer{
				{
					A: "a1",
					Inner: Inner{
						B: "b1",
						C: "c1",
					},
					D: "d1",
					E: "e1",
				},
			},
			expectedHeader: []string{"a", "b", "c", "d", "e"},
			expectedData: [][]string{
				{"a1", "b1", "c1", "d1", "e1"},
			},
		},
		{
			name: "override embedded field",
			input: []Override{
				{
					Inner: Inner{
						B: "should not appear",
						C: "c1",
					},
					B: "b1",
					F: "f1",
				},
			},
			expectedHeader: []string{"c", "b", "f"},
			expectedData: [][]string{
				{"c1", "b1", "f1"},
			},
		},
		{
			name:           "empty struct slice",
			input:          []Outer{},
			expectedHeader: nil,
			expectedData:   nil,
		},
		{
			name: "multiple level embedding",
			input: []MultipleEmbedded{
				{
					Inner: Inner{
						B: "b1",
						C: "c1",
					},
					DeepOuter: DeepOuter{
						DeepInner: DeepInner{
							X: "x1",
							Y: "y1",
						},
						Z: "z1",
					},
					M: "m1",
				},
			},
			expectedHeader: []string{"b", "c", "x", "y", "z", "m"},
			expectedData: [][]string{
				{"b1", "c1", "x1", "y1", "z1", "m1"},
			},
		},
		{
			name: "duplicate tags",
			input: []Duplicate{
				{
					A1: "first",
					A2: "second", // The last one should win
				},
			},
			expectedHeader: []string{"same"},
			expectedData: [][]string{
				{"second"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, data, err := tablemap.Marshal(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedHeader, header, "Headers should match expected order")
			assert.Equal(t, tt.expectedData, data, "Data should match expected order")
		})
	}
}

func TestRowHandler(t *testing.T) {
	type Person struct {
		Name   string  `table:"name"`
		Age    int     `table:"age"`
		Height float64 `table:"height"`
	}

	header := []string{"name", "age", "height"}
	data := []string{"Alice", "25", "165.5"}

	// Create type-safe row handler
	handler, err := tablemap.NewRowHandler[Person](header, nil)
	if err != nil {
		t.Fatalf("NewRowHandler failed: %v", err)
	}

	// Test unmarshaling
	person, err := handler.UnmarshalRow(data)
	if err != nil {
		t.Fatalf("UnmarshalRow failed: %v", err)
	}

	// Verify unmarshaled data
	if person.Name != "Alice" || person.Age != 25 || person.Height != 165.5 {
		t.Errorf("UnmarshalRow result mismatch: got %+v", person)
	}

	// Test marshaling
	out, err := handler.MarshalRow(person)
	if err != nil {
		t.Fatalf("MarshalRow failed: %v", err)
	}

	// Verify marshaled data
	if !reflect.DeepEqual(out, data) {
		t.Errorf("MarshalRow result mismatch: got %v, want %v", out, data)
	}
}
