package tablemap_test

import (
	"fmt"

	"github.com/kmio11/tablemap"
)

func ExampleUnmarshal() {
	header := []string{"name", "age"}
	data := [][]string{
		{"Alice", "23"},
		{"Bob", "25"},
		{"Charlie", "27"},
	}

	type User struct {
		Name string `table:"name"`
		Age  int    `table:"age"`
	}

	var users []User
	if err := tablemap.Unmarshal(header, data, &users); err != nil {
		panic(err)
	}

	for _, u := range users {
		fmt.Printf("%s is %d years old\n", u.Name, u.Age)
	}

	newHeader, newData, err := tablemap.Marshal(users)
	if err != nil {
		panic(err)
	}

	fmt.Println("\nMarshal result:")
	fmt.Println("Header:", newHeader)
	for _, row := range newData {
		fmt.Println("Data:", row)
	}

	// Output:
	// Alice is 23 years old
	// Bob is 25 years old
	// Charlie is 27 years old
	//
	// Marshal result:
	// Header: [name age]
	// Data: [Alice 23]
	// Data: [Bob 25]
	// Data: [Charlie 27]
}

func ExampleMarshal() {
	type User struct {
		Name string `table:"name"`
		Age  int    `table:"age"`
	}

	users := []User{
		{Name: "Alice", Age: 23},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 27},
	}

	header, data, err := tablemap.Marshal(users)
	if err != nil {
		panic(err)
	}

	fmt.Println("Header:", header)
	for _, row := range data {
		fmt.Println("Data:", row)
	}

	// Output:
	// Header: [name age]
	// Data: [Alice 23]
	// Data: [Bob 25]
	// Data: [Charlie 27]
}
