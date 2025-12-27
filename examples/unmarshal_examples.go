package main

import (
	"fmt"
	"log"

	"github.com/tron-format/trongo/pkg/tron"
)

// Example 1: Unmarshal into struct
func example1_UnmarshalStruct() {
	fmt.Println("=== Example 1: Unmarshal into Struct ===")

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// TRON with class definition
	input := `class A: name,age

A("Alice",30)`

	var person Person
	err := tron.Unmarshal([]byte(input), &person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %+v\n", person)
	// Output: {Name:Alice Age:30}

	// Also works with JSON-style object syntax
	input2 := `{"name":"Bob","age":25}`
	var person2 Person
	tron.Unmarshal([]byte(input2), &person2)
	fmt.Printf("From object syntax: %+v\n\n", person2)
}

// Example 2: Unmarshal array into slice
func example2_UnmarshalArray() {
	fmt.Println("=== Example 2: Unmarshal Array into Slice ===")

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	input := `class A: name,age

[A("Alice",30),A("Bob",25),A("Charlie",35)]`

	var people []Person
	err := tron.Unmarshal([]byte(input), &people)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Unmarshaled people:")
	for i, p := range people {
		fmt.Printf("  %d: %+v\n", i+1, p)
	}
	fmt.Println()
}

// Example 3: Unmarshal into interface{}
func example3_UnmarshalInterface() {
	fmt.Println("=== Example 3: Unmarshal into interface{} ===")

	inputs := []string{
		`42`,
		`"hello"`,
		`true`,
		`[1,2,3]`,
		`{"name":"Alice","age":30}`,
		`class A: x,y

A(10,20)`,
	}

	for _, input := range inputs {
		var result interface{}
		tron.Unmarshal([]byte(input), &result)
		fmt.Printf("Input:  %s\n", input[:min(len(input), 40)])
		fmt.Printf("Result: %v (type: %T)\n\n", result, result)
	}
}

// Example 4: Unmarshal nested structures
func example4_NestedStructuresUnmarshal() {
	fmt.Println("=== Example 4: Nested Structures ===")

	type Address struct {
		Street string `json:"street"`
		City   string `json:"city"`
	}

	type Person struct {
		Name    string  `json:"name"`
		Age     int     `json:"age"`
		Address Address `json:"address"`
	}

	input := `class Address: street,city
class Person: name,age,address

Person("Alice",30,Address("123 Main St","NYC"))`

	var person Person
	err := tron.Unmarshal([]byte(input), &person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Person: %+v\n", person)
	fmt.Printf("Address: %+v\n\n", person.Address)
}

// Example 5: Unmarshal with missing fields
func example5_MissingFields() {
	fmt.Println("=== Example 5: Handling Missing Fields ===")

	type Person struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Email string `json:"email"` // Not in TRON data
	}

	input := `{"name":"Alice","age":30}`

	var person Person
	tron.Unmarshal([]byte(input), &person)

	fmt.Printf("Result: %+v\n", person)
	fmt.Printf("Email field (missing in data): %q\n", person.Email)
	// Empty string - Go zero value
	fmt.Println()
}

// Example 6: Unmarshal with unknown fields
func example6_UnknownFields() {
	fmt.Println("=== Example 6: Handling Unknown Fields ===")

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
		// TRON data has 'city' but struct doesn't - it's ignored
	}

	input := `{"name":"Alice","age":30,"city":"NYC","country":"USA"}`

	var person Person
	err := tron.Unmarshal([]byte(input), &person)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Result: %+v\n", person)
	fmt.Println("Unknown fields (city, country) were ignored - no error!")
	fmt.Println()
}

// Example 7: Unmarshal into map
func example7_UnmarshalMap() {
	fmt.Println("=== Example 7: Unmarshal into Map ===")

	input := `{"name":"Alice","age":30,"active":true}`

	var result map[string]interface{}
	tron.Unmarshal([]byte(input), &result)

	fmt.Printf("Map: %v\n", result)
	fmt.Printf("Access name: %v\n", result["name"])
	fmt.Printf("Access age: %v (type: %T)\n\n", result["age"], result["age"])
}

// Example 8: Unmarshal with struct tags
func example8_StructTags() {
	fmt.Println("=== Example 8: Struct Tags ===")

	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		FullName string `json:"full_name"`
		IsActive bool   `json:"active"`
		Password string `json:"-"` // Ignored field
	}

	input := `{"id":1,"username":"alice","full_name":"Alice Smith","active":true,"password":"should_be_ignored"}`

	var user User
	tron.Unmarshal([]byte(input), &user)

	fmt.Printf("User: %+v\n", user)
	fmt.Printf("Password field (marked with -): %q\n", user.Password)
	fmt.Println()
}

// Example 9: Unmarshal primitives
func example9_Primitives() {
	fmt.Println("=== Example 9: Unmarshal Primitives ===")

	// String
	var str string
	tron.Unmarshal([]byte(`"hello world"`), &str)
	fmt.Printf("String: %q\n", str)

	// Number
	var num int
	tron.Unmarshal([]byte(`42`), &num)
	fmt.Printf("Int: %d\n", num)

	// Float
	var flt float64
	tron.Unmarshal([]byte(`3.14`), &flt)
	fmt.Printf("Float: %.2f\n", flt)

	// Boolean
	var b bool
	tron.Unmarshal([]byte(`true`), &b)
	fmt.Printf("Bool: %v\n", b)

	// Null into pointer
	var ptr *string
	tron.Unmarshal([]byte(`null`), &ptr)
	fmt.Printf("Null pointer: %v\n\n", ptr)
}

// Example 10: Unmarshal with type conversions
func example10_TypeConversions() {
	fmt.Println("=== Example 10: Type Conversions ===")

	// Number to different int types
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64

	tron.Unmarshal([]byte(`100`), &i8)
	tron.Unmarshal([]byte(`30000`), &i16)
	tron.Unmarshal([]byte(`2000000000`), &i32)
	tron.Unmarshal([]byte(`9000000000000000000`), &i64)

	fmt.Printf("int8:  %d\n", i8)
	fmt.Printf("int16: %d\n", i16)
	fmt.Printf("int32: %d\n", i32)
	fmt.Printf("int64: %d\n\n", i64)
}

// Example 11: Unmarshal error handling
func example11_ErrorHandling() {
	fmt.Println("=== Example 11: Error Handling ===")

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Invalid TRON syntax
	err1 := tron.Unmarshal([]byte(`[1,2,`), &Person{})
	if err1 != nil {
		fmt.Printf("Syntax error: %v\n", err1)
	}

	// Type mismatch
	var num int
	err2 := tron.Unmarshal([]byte(`"not a number"`), &num)
	if err2 != nil {
		fmt.Printf("Type error: %v\n", err2)
	}

	// Not a pointer
	var person Person
	err3 := tron.Unmarshal([]byte(`{}`), person) // Missing &
	if err3 != nil {
		fmt.Printf("Invalid target: %v\n", err3)
	}
	fmt.Println()
}

// Example 12: Unmarshal map with different key types
func example12_MapKeyTypes() {
	fmt.Println("=== Example 12: Map Key Types ===")

	// String keys
	var mapStr map[string]int
	tron.Unmarshal([]byte(`{"one":1,"two":2}`), &mapStr)
	fmt.Printf("String keys: %v\n", mapStr)

	// Int keys (keys are strings in TRON, converted to int)
	var mapInt map[int]string
	tron.Unmarshal([]byte(`{"1":"one","2":"two"}`), &mapInt)
	fmt.Printf("Int keys: %v\n\n", mapInt)
}

// Example 13: Complex real-world example
func example13_RealWorld() {
	fmt.Println("=== Example 13: Real-World API Response ===")

	type User struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	type APIResponse struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
		Data    []User `json:"data"`
	}

	input := `class User: id,username,email

{"success":true,"message":"Users retrieved","data":[User(1,"alice","alice@example.com"),User(2,"bob","bob@example.com"),User(3,"charlie","charlie@example.com")]}`

	var response APIResponse
	err := tron.Unmarshal([]byte(input), &response)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Success: %v\n", response.Success)
	fmt.Printf("Message: %s\n", response.Message)
	fmt.Printf("Users (%d):\n", len(response.Data))
	for _, user := range response.Data {
		fmt.Printf("  - %s (ID: %d, Email: %s)\n",
			user.Username, user.ID, user.Email)
	}
	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	example1_UnmarshalStruct()
	example2_UnmarshalArray()
	example3_UnmarshalInterface()
	example4_NestedStructuresUnmarshal()
	example5_MissingFields()
	example6_UnknownFields()
	example7_UnmarshalMap()
	example8_StructTags()
	example9_Primitives()
	example10_TypeConversions()
	example11_ErrorHandling()
	example12_MapKeyTypes()
	example13_RealWorld()

	fmt.Println("=== Summary ===")
	fmt.Println("TRON Unmarshal supports:")
	fmt.Println("✓ All Go types (primitives, structs, slices, maps, pointers)")
	fmt.Println("✓ Struct tags (json:\"name,omitempty\")")
	fmt.Println("✓ Class definitions AND JSON-style objects")
	fmt.Println("✓ Nested structures")
	fmt.Println("✓ Type conversions")
	fmt.Println("✓ Missing/unknown fields handled gracefully")
	fmt.Println("✓ Same API as encoding/json!")
}
