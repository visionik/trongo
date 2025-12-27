package comparison

import (
	"encoding/json"
	"fmt"

	"github.com/tron-format/trongo/pkg/tron"
)

// Example 1: Simple struct marshaling
func example1_SimpleStruct() {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	person := Person{Name: "Alice", Age: 30}

	fmt.Println("=== Example 1: Simple Struct ===")

	// JSON
	jsonData, _ := json.Marshal(person)
	fmt.Printf("JSON:  %s\n", jsonData)
	// Output: {"name":"Alice","age":30}

	// TRON
	tronData, _ := tron.Marshal(person)
	fmt.Printf("TRON:  %s\n", tronData)
	// Output: {name:"Alice",age:30}

	fmt.Println()
}

// Example 2: Array of structs (TRON's strength!)
func example2_ArrayOfStructs() {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	people := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	fmt.Println("=== Example 2: Array of Structs (TRON shines here!) ===")

	// JSON - repeats field names for each object
	jsonData, _ := json.Marshal(people)
	fmt.Printf("JSON (%d bytes):\n%s\n\n", len(jsonData), jsonData)
	// Output: [{"name":"Alice","age":30},{"name":"Bob","age":25},{"name":"Charlie","age":35}]

	// TRON - defines structure once, reuses it
	tronData, _ := tron.Marshal(people)
	fmt.Printf("TRON (%d bytes):\n%s\n", len(tronData), tronData)
	// Output:
	// class A: name,age
	//
	// [A("Alice",30),A("Bob",25),A("Charlie",35)]

	fmt.Printf("\nToken savings: %.1f%%\n\n", 100.0*(1.0-float64(len(tronData))/float64(len(jsonData))))
}

// Example 3: Unmarshaling
func example3_Unmarshaling() {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	fmt.Println("=== Example 3: Unmarshaling ===")

	// JSON unmarshaling
	jsonInput := `{"name":"Alice","age":30}`
	var personFromJSON Person
	json.Unmarshal([]byte(jsonInput), &personFromJSON)
	fmt.Printf("From JSON: %+v\n", personFromJSON)

	// TRON unmarshaling
	tronInput := `{name:"Alice",age:30}`
	var personFromTRON Person
	tron.Unmarshal([]byte(tronInput), &personFromTRON)
	fmt.Printf("From TRON: %+v\n", personFromTRON)

	// TRON with class definition
	tronWithClass := `class A: name,age

A("Bob",25)`
	var personFromClass Person
	tron.Unmarshal([]byte(tronWithClass), &personFromClass)
	fmt.Printf("From TRON class: %+v\n\n", personFromClass)
}

// Example 4: Nested structures
func example4_NestedStructures() {
	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	people := []Person{
		{Name: "Alice", Age: 30, Address: Address{Street: "123 Main St", City: "NYC"}},
		{Name: "Bob", Age: 25, Address: Address{Street: "456 Oak Ave", City: "LA"}},
	}

	fmt.Println("=== Example 4: Nested Structures ===")

	// JSON
	jsonData, _ := json.MarshalIndent(people, "", "  ")
	fmt.Printf("JSON (%d bytes):\n%s\n\n", len(jsonData), jsonData)

	// TRON
	tronData, _ := tron.Marshal(people)
	fmt.Printf("TRON (%d bytes):\n%s\n\n", len(tronData), tronData)
}

// Example 5: Working with maps
func example5_Maps() {
	data := map[string]interface{}{
		"name":   "Alice",
		"age":    30,
		"active": true,
	}

	fmt.Println("=== Example 5: Maps ===")

	// JSON
	jsonData, _ := json.Marshal(data)
	fmt.Printf("JSON:  %s\n", jsonData)

	// TRON
	tronData, _ := tron.Marshal(data)
	fmt.Printf("TRON:  %s\n\n", tronData)
}

// Example 6: Primitives and arrays
func example6_Primitives() {
	fmt.Println("=== Example 6: Primitives ===")

	// String
	jsonStr, _ := json.Marshal("hello")
	tronStr, _ := tron.Marshal("hello")
	fmt.Printf("String - JSON: %s, TRON: %s\n", jsonStr, tronStr)

	// Number
	jsonNum, _ := json.Marshal(42)
	tronNum, _ := tron.Marshal(42)
	fmt.Printf("Number - JSON: %s, TRON: %s\n", jsonNum, tronNum)

	// Boolean
	jsonBool, _ := json.Marshal(true)
	tronBool, _ := tron.Marshal(true)
	fmt.Printf("Bool   - JSON: %s, TRON: %s\n", jsonBool, tronBool)

	// Array
	jsonArr, _ := json.Marshal([]int{1, 2, 3})
	tronArr, _ := tron.Marshal([]int{1, 2, 3})
	fmt.Printf("Array  - JSON: %s, TRON: %s\n\n", jsonArr, tronArr)
}

// Example 7: struct tags and omitempty
func example7_StructTags() {
	type User struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Email    string `json:"email,omitempty"`
		Password string `json:"-"` // Always omitted
	}

	user := User{
		ID:       1,
		Name:     "Alice",
		Email:    "",
		Password: "secret123",
	}

	fmt.Println("=== Example 7: Struct Tags (omitempty, ignore) ===")

	// JSON
	jsonData, _ := json.Marshal(user)
	fmt.Printf("JSON:  %s\n", jsonData)
	// Output: {"id":1,"name":"Alice"}
	// Note: email is omitted (empty), password is ignored

	// TRON
	tronData, _ := tron.Marshal(user)
	fmt.Printf("TRON:  %s\n", tronData)
	// Output: {id:1,name:"Alice"}

	fmt.Println()
}

// Example 8: Real-world API response
func example8_APIResponse() {
	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Active   bool   `json:"active"`
	}

	type APIResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Users   []User `json:"users"`
	}

	response := APIResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Users: []User{
			{ID: 1, Username: "alice", Email: "alice@example.com", Active: true},
			{ID: 2, Username: "bob", Email: "bob@example.com", Active: true},
			{ID: 3, Username: "charlie", Email: "charlie@example.com", Active: false},
		},
	}

	fmt.Println("=== Example 8: Real-World API Response ===")

	// JSON
	jsonData, _ := json.MarshalIndent(response, "", "  ")
	fmt.Printf("JSON (%d bytes):\n%s\n\n", len(jsonData), jsonData)

	// TRON
	tronData, _ := tron.Marshal(response)
	fmt.Printf("TRON (%d bytes):\n%s\n", len(tronData), tronData)

	fmt.Printf("\nSavings: %d bytes (%.1f%%)\n\n",
		len(jsonData)-len(tronData),
		100.0*(1.0-float64(len(tronData))/float64(len(jsonData))))
}

// Example 9: Unmarshaling into interface{}
func example9_UnmarshalInterface() {
	fmt.Println("=== Example 9: Unmarshal into interface{} ===")

	// JSON
	jsonInput := `{"name":"Alice","age":30,"active":true}`
	var jsonResult interface{}
	json.Unmarshal([]byte(jsonInput), &jsonResult)
	fmt.Printf("JSON result: %+v\n", jsonResult)
	fmt.Printf("Type: %T\n\n", jsonResult)

	// TRON
	tronInput := `{name:"Alice",age:30,active:true}`
	var tronResult interface{}
	tron.Unmarshal([]byte(tronInput), &tronResult)
	fmt.Printf("TRON result: %+v\n", tronResult)
	fmt.Printf("Type: %T\n\n", tronResult)
}

// Example 10: Round-trip conversion
func example10_RoundTrip() {
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	original := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}

	fmt.Println("=== Example 10: Round-trip Conversion ===")
	fmt.Printf("Original: %+v\n\n", original)

	// JSON round-trip
	jsonData, _ := json.Marshal(original)
	var jsonResult []Person
	json.Unmarshal(jsonData, &jsonResult)
	fmt.Printf("After JSON round-trip:  %+v\n", jsonResult)

	// TRON round-trip
	tronData, _ := tron.Marshal(original)
	var tronResult []Person
	tron.Unmarshal(tronData, &tronResult)
	fmt.Printf("After TRON round-trip:  %+v\n", tronResult)

	fmt.Printf("\nJSON data: %s\n", jsonData)
	fmt.Printf("TRON data:\n%s\n\n", tronData)
}

func main() {
	example1_SimpleStruct()
	example2_ArrayOfStructs()
	example3_Unmarshaling()
	example4_NestedStructures()
	example5_Maps()
	example6_Primitives()
	example7_StructTags()
	example8_APIResponse()
	example9_UnmarshalInterface()
	example10_RoundTrip()
}
