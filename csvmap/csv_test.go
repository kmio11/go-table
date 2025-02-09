package csvmap_test

import (
	"bytes"
	"testing"
	"text/template"
	"time"

	"github.com/kmio11/tablemap"
	"github.com/kmio11/tablemap/csvmap"
	"github.com/stretchr/testify/assert"
)

type TestTime struct {
	Time time.Time
}

func (t *TestTime) MarshalTable() (string, error) {
	return t.Time.Format(time.RFC3339), nil
}

func (t *TestTime) UnmarshalTable(s string) error {
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	t.Time = parsed
	return nil
}

type TestStruct struct {
	String string   `table:"string"`
	Int    int      `table:"int"`
	Time   TestTime `table:"time"`
}

type TestStructPtr struct {
	String *string   `table:"string"`
	Int    *int      `table:"int"`
	Time   *TestTime `table:"time"`
}

type testData struct {
	String string
	Int    string
	Time   string
}

var csvTemplate = template.Must(template.New("csv").Parse(`string,int,time
{{- range .}}
{{.String}},{{.Int}},{{.Time}}
{{- end}}
`))

func TestReader(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name     string
		data     []testData
		expected []TestStruct
	}{
		{
			name: "single row",
			data: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
			},
			expected: []TestStruct{
				{
					String: "test1",
					Int:    123,
					Time:   TestTime{Time: now},
				},
			},
		},
		{
			name: "multiple rows",
			data: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
				{
					String: "test2",
					Int:    "456",
					Time:   now.Add(24 * time.Hour).Format(time.RFC3339),
				},
			},
			expected: []TestStruct{
				{
					String: "test1",
					Int:    123,
					Time:   TestTime{Time: now},
				},
				{
					String: "test2",
					Int:    456,
					Time:   TestTime{Time: now.Add(24 * time.Hour)},
				},
			},
		},
		{
			name: "with empty string",
			data: []testData{
				{
					String: "",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
			},
			expected: []TestStruct{
				{
					String: "",
					Int:    123,
					Time:   TestTime{Time: now},
				},
			},
		},
		{
			name: "with quoted string",
			data: []testData{
				{
					String: `"test,1"`,
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
			},
			expected: []TestStruct{
				{
					String: "test,1",
					Int:    123,
					Time:   TestTime{Time: now},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := csvTemplate.Execute(&buf, tt.data)
			assert.NoError(t, err)

			reader := csvmap.NewReader(&buf, nil)
			result, err := csvmap.ReadAll[TestStruct](reader)
			assert.NoError(t, err)

			assert.Equal(t, len(tt.expected), len(result))
			for i := range tt.expected {
				assert.Equal(t, tt.expected[i].String, result[i].String)
				assert.Equal(t, tt.expected[i].Int, result[i].Int)
				assert.Equal(t, tt.expected[i].Time.Time.Unix(), result[i].Time.Time.Unix())
			}
		})
	}
}

func P[T any](t T) *T {
	return &t
}

func TestReader_nil_options(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name    string
		opts    *tablemap.Options
		data    []testData
		expect  []TestStructPtr
		wantErr bool
	}{
		{
			name: "with NULL as nil value",
			opts: &tablemap.Options{NilValue: "NULL"},
			data: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
				{
					String: "NULL",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
			},
			expect: []TestStructPtr{
				{
					String: P("test1"),
					Int:    P(123),
					Time:   &TestTime{Time: now},
				},
				{
					String: nil,
					Int:    P(123),
					Time:   &TestTime{Time: now},
				},
			},
		},
		{
			name: "with custom nil value",
			opts: &tablemap.Options{NilValue: "-"},
			data: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
				{
					String: "-",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
			},
			expect: []TestStructPtr{
				{
					String: P("test1"),
					Int:    P(123),
					Time:   &TestTime{Time: now},
				},
				{
					String: nil,
					Int:    P(123),
					Time:   &TestTime{Time: now},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := csvTemplate.Execute(&buf, tt.data)
			assert.NoError(t, err)

			reader := csvmap.NewReader(&buf, tt.opts)
			result, err := csvmap.ReadAll[TestStructPtr](reader)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			assert.Equal(t, len(tt.expect), len(result))
			for i := range tt.expect {
				assert.Equal(t, tt.expect[i].String, result[i].String)
				assert.Equal(t, tt.expect[i].Int, result[i].Int)
				assert.Equal(t, tt.expect[i].Time.Time.Unix(), result[i].Time.Time.Unix())
			}
		})
	}
}

func TestWriter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name     string
		input    []TestStruct
		expected []testData
	}{
		{
			name: "single row",
			input: []TestStruct{
				{
					String: "test1",
					Int:    123,
					Time:   TestTime{Time: now},
				},
			},
			expected: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
			},
		},
		{
			name: "multiple rows",
			input: []TestStruct{
				{
					String: "test1",
					Int:    123,
					Time:   TestTime{Time: now},
				},
				{
					String: "test2",
					Int:    456,
					Time:   TestTime{Time: now.Add(24 * time.Hour)},
				},
			},
			expected: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
				{
					String: "test2",
					Int:    "456",
					Time:   now.Add(24 * time.Hour).Format(time.RFC3339),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := csvmap.NewWriter(&buf, nil)

			err := csvmap.WriteAll(writer, tt.input)
			assert.NoError(t, err)

			var expected bytes.Buffer
			err = csvTemplate.Execute(&expected, tt.expected)
			assert.NoError(t, err)

			assert.Equal(t, expected.String(), buf.String())
		})
	}
}

func TestWriter_nil_options(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name     string
		opts     *tablemap.Options
		input    []TestStructPtr
		expected []testData
		wantErr  bool
	}{
		{
			name: "with NULL as nil value",
			opts: &tablemap.Options{NilValue: "NULL"},
			input: []TestStructPtr{
				{
					String: P("test1"),
					Int:    P(123),
					Time:   &TestTime{Time: now},
				},
				{
					String: nil,
					Int:    nil,
					Time:   nil,
				},
			},
			expected: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
				{
					String: "NULL",
					Int:    "NULL",
					Time:   "NULL",
				},
			},
		},
		{
			name: "with custom nil value",
			opts: &tablemap.Options{NilValue: "-"},
			input: []TestStructPtr{
				{
					String: P("test1"),
					Int:    P(123),
					Time:   &TestTime{Time: now},
				},
				{
					String: nil,
					Int:    nil,
					Time:   nil,
				},
			},
			expected: []testData{
				{
					String: "test1",
					Int:    "123",
					Time:   now.Format(time.RFC3339),
				},
				{
					String: "-",
					Int:    "-",
					Time:   "-",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := csvmap.NewWriter(&buf, tt.opts)

			err := csvmap.WriteAll(writer, tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var expected bytes.Buffer
			err = csvTemplate.Execute(&expected, tt.expected)
			assert.NoError(t, err)

			assert.Equal(t, expected.String(), buf.String())
		})
	}
}
