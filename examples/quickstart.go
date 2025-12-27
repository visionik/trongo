package quickstart

import (
	"fmt"
	"log"

	"github.com/tron-format/trongo/pkg/tron"
)

func main() {
	// Define a struct (same as you would for JSON)
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	// Create some data
	people := []Person{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	fmt.Println("Original data:")
	fmt.Printf("%+v\n\n", people)

	// ========================================
	// MARSHALING (Go → TRON)
	// ========================================

	// Marshal to TRON - exactly like json.Marshal()
	tronData, err := tron.Marshal(people)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("TRON encoded:")
	fmt.Println(string(tronData))
	fmt.Println()

	// ========================================
	// UNMARSHALING (TRON → Go)
	// ========================================

	// Unmarshal from TRON - exactly like json.Unmarshal()
	var decoded []Person
	err = tron.Unmarshal(tronData, &decoded)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Decoded back to Go:")
	fmt.Printf("%+v\n\n", decoded)

	// ========================================
	// MIGRATION FROM JSON
	// ========================================

	// Your existing code:
	// import "encoding/json"
	// data, err := json.Marshal(myStruct)
	// err = json.Unmarshal(data, &result)

	// Just change the import!
	// import "github.com/tron-format/trongo/pkg/tron"
	// data, err := tron.Marshal(myStruct)
	// err = tron.Unmarshal(data, &result)

	// That's it! Same API, more efficient format.
}
