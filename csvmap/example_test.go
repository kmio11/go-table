package csvmap_test

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/kmio11/tablemap/csvmap"
)

func ExampleReader_ReadAll() {
	csvData := `name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com`

	type Person struct {
		Name  string `table:"name"`
		Age   int    `table:"age"`
		Email string `table:"email"`
	}

	reader := csvmap.NewReader[Person](strings.NewReader(csvData), nil)
	persons, err := reader.ReadAll()
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

func ExampleWriter_WriteAll() {
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
	writer := csvmap.NewWriter[Person](&buf, nil)
	if err := writer.WriteAll(persons); err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(buf.String())
	// Output:
	// name,age,email
	// John Doe,30,john@example.com
	// Jane Smith,25,jane@example.com
}

func ExampleReader_Read() {
	csvData := `name,age,email
John Doe,30,john@example.com
Jane Smith,25,jane@example.com`

	type Person struct {
		Name  string `table:"name"`
		Age   int    `table:"age"`
		Email string `table:"email"`
	}

	reader := csvmap.NewReader[Person](strings.NewReader(csvData), nil)
	for {
		person, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Printf("%s is %d years old (email: %s)\n", person.Name, person.Age, person.Email)
	}
	// Output:
	// John Doe is 30 years old (email: john@example.com)
	// Jane Smith is 25 years old (email: jane@example.com)
}

func ExampleWriter_Write() {
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
	writer := csvmap.NewWriter[Person](&buf, nil)

	for _, p := range persons {
		if err := writer.Write(p); err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	writer.W.Flush()

	fmt.Println(buf.String())
	// Output:
	// name,age,email
	// John Doe,30,john@example.com
	// Jane Smith,25,jane@example.com
}
