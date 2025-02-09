package tblcsv_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kmio11/go-table/tblcsv"
)

func ExampleReadAll() {
	csvData := `name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com`

	type Person struct {
		Name  string `table:"name"`
		Age   int    `table:"age"`
		Email string `table:"email"`
	}

	reader := tblcsv.NewReader(strings.NewReader(csvData), nil)
	persons, err := tblcsv.ReadAll[Person](reader)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, p := range persons {
		fmt.Printf("%s is %d years old (email: %s)\n", p.Name, p.Age, p.Email)
	}
	// Output:
	// John Doe is 30 years old (email: john@example.com)
	// Jane Smith is 25 years old (email: jane@example.com)
}

func ExampleWriteAll() {
	type Person struct {
		Name  string `table:"name"`
		Age   int    `table:"age"`
		Email string `table:"email"`
	}

	persons := []Person{
		{Name: "John Doe", Age: 30, Email: "john@example.com"},
		{Name: "Jane Smith", Age: 25, Email: "jane@example.com"},
	}

	var buf bytes.Buffer
	writer := tblcsv.NewWriter(&buf, nil)
	if err := tblcsv.WriteAll(writer, persons); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(buf.String())
	// Output:
	// name,age,email
	// John Doe,30,john@example.com
	// Jane Smith,25,jane@example.com
}
