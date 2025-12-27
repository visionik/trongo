package main

import (
	"encoding/json"
	"fmt"

	"github.com/tron-format/trongo/pkg/tron"
)

// Use Case 1: API Responses with repeated structure
func useCase1_APIResponses() {
	type Product struct {
		ID          int     `json:"id"`
		Name        string  `json:"name"`
		Price       float64 `json:"price"`
		InStock     bool    `json:"in_stock"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
	}

	products := []Product{
		{1, "Laptop", 999.99, true, "Electronics", "High-performance laptop"},
		{2, "Mouse", 29.99, true, "Electronics", "Wireless mouse"},
		{3, "Keyboard", 79.99, true, "Electronics", "Mechanical keyboard"},
		{4, "Monitor", 299.99, false, "Electronics", "27-inch 4K monitor"},
		{5, "Desk", 399.99, true, "Furniture", "Adjustable standing desk"},
	}

	fmt.Println("=== Use Case 1: API Response (Product Catalog) ===")

	jsonData, _ := json.Marshal(products)
	tronData, _ := tron.Marshal(products)

	fmt.Printf("JSON size: %d bytes\n", len(jsonData))
	fmt.Printf("TRON size: %d bytes\n", len(tronData))
	fmt.Printf("Savings: %.1f%%\n\n", 100.0*(1.0-float64(len(tronData))/float64(len(jsonData))))

	fmt.Println("TRON output:")
	fmt.Println(string(tronData))
	fmt.Println()
}

// Use Case 2: Time-series data
func useCase2_TimeSeriesData() {
	type DataPoint struct {
		Timestamp int64   `json:"timestamp"`
		Value     float64 `json:"value"`
		Sensor    string  `json:"sensor"`
		Status    string  `json:"status"`
	}

	// Simulating IoT sensor data
	var data []DataPoint
	for i := 0; i < 10; i++ {
		data = append(data, DataPoint{
			Timestamp: 1640000000 + int64(i*60),
			Value:     20.5 + float64(i)*0.1,
			Sensor:    "temp-sensor-01",
			Status:    "ok",
		})
	}

	fmt.Println("=== Use Case 2: Time-Series/IoT Data (10 data points) ===")

	jsonData, _ := json.Marshal(data)
	tronData, _ := tron.Marshal(data)

	fmt.Printf("JSON size: %d bytes\n", len(jsonData))
	fmt.Printf("TRON size: %d bytes\n", len(tronData))
	fmt.Printf("Savings: %.1f%%\n\n", 100.0*(1.0-float64(len(tronData))/float64(len(jsonData))))

	fmt.Println("First 3 points in JSON:")
	jsonFirst3, _ := json.Marshal(data[:3])
	fmt.Println(string(jsonFirst3))
	fmt.Println()

	fmt.Println("First 3 points in TRON:")
	tronFirst3, _ := tron.Marshal(data[:3])
	fmt.Println(string(tronFirst3))
	fmt.Println()
}

// Use Case 3: Database query results
func useCase3_DatabaseResults() {
	type User struct {
		ID        int    `json:"id"`
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Active    bool   `json:"active"`
	}

	users := []User{
		{1, "alice", "alice@example.com", "Alice", "Smith", true},
		{2, "bob", "bob@example.com", "Bob", "Jones", true},
		{3, "charlie", "charlie@example.com", "Charlie", "Brown", false},
		{4, "david", "david@example.com", "David", "Wilson", true},
		{5, "eve", "eve@example.com", "Eve", "Davis", true},
	}

	fmt.Println("=== Use Case 3: Database Query Results ===")

	jsonData, _ := json.Marshal(users)
	tronData, _ := tron.Marshal(users)

	fmt.Printf("JSON size: %d bytes\n", len(jsonData))
	fmt.Printf("TRON size: %d bytes\n", len(tronData))
	fmt.Printf("Savings: %.1f%%\n\n", 100.0*(1.0-float64(len(tronData))/float64(len(jsonData))))

	// Pretty print for comparison
	jsonPretty, _ := json.MarshalIndent(users[:2], "", "  ")
	fmt.Println("First 2 users in JSON (pretty):")
	fmt.Println(string(jsonPretty))
	fmt.Println()

	fmt.Println("Same data in TRON:")
	tronSample, _ := tron.Marshal(users[:2])
	fmt.Println(string(tronSample))
	fmt.Println()
}

// Use Case 4: Configuration data
func useCase4_Configuration() {
	type ServerConfig struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
		Timeout  int    `json:"timeout"`
	}

	type Config struct {
		Environment string         `json:"environment"`
		Debug       bool           `json:"debug"`
		Servers     []ServerConfig `json:"servers"`
	}

	config := Config{
		Environment: "production",
		Debug:       false,
		Servers: []ServerConfig{
			{"api.example.com", 443, "https", 30},
			{"db.example.com", 5432, "postgresql", 60},
			{"cache.example.com", 6379, "redis", 10},
		},
	}

	fmt.Println("=== Use Case 4: Configuration Files ===")

	jsonData, _ := json.MarshalIndent(config, "", "  ")
	tronData, _ := tron.Marshal(config)

	fmt.Println("JSON (pretty-printed):")
	fmt.Println(string(jsonData))
	fmt.Printf("Size: %d bytes\n\n", len(jsonData))

	fmt.Println("TRON (compact):")
	fmt.Println(string(tronData))
	fmt.Printf("Size: %d bytes\n\n", len(tronData))

	fmt.Printf("Savings: %.1f%%\n\n", 100.0*(1.0-float64(len(tronData))/float64(len(jsonData))))
}

// Use Case 5: Real-world comparison - when JSON is actually better
func useCase5_WhenJSONIsBetter() {
	// Single object or heterogeneous data
	data := map[string]interface{}{
		"name":        "Alice",
		"age":         30,
		"city":        "NYC",
		"coordinates": []float64{40.7128, -74.0060},
		"metadata": map[string]string{
			"source": "manual",
			"type":   "verified",
		},
	}

	fmt.Println("=== Use Case 5: When JSON is Fine ===")
	fmt.Println("(Single objects or very heterogeneous data)")
	fmt.Println()

	jsonData, _ := json.MarshalIndent(data, "", "  ")
	tronData, _ := tron.Marshal(data)

	fmt.Println("JSON:")
	fmt.Println(string(jsonData))
	fmt.Printf("Size: %d bytes\n\n", len(jsonData))

	fmt.Println("TRON:")
	fmt.Println(string(tronData))
	fmt.Printf("Size: %d bytes\n\n", len(tronData))

	fmt.Println("For single objects, sizes are similar.")
	fmt.Println("TRON shines with arrays of similar structures!\n")
}

// Summary function
func summary() {
	fmt.Println("=== SUMMARY: When to Use TRON vs JSON ===")
	fmt.Println()
	fmt.Println("✅ Use TRON when:")
	fmt.Println("   • Arrays of objects with same structure (API responses, query results)")
	fmt.Println("   • Time-series or IoT data")
	fmt.Println("   • Repetitive data structures")
	fmt.Println("   • Token/bandwidth efficiency matters")
	fmt.Println("   • LLM token costs are a concern")
	fmt.Println()
	fmt.Println("📄 JSON is fine for:")
	fmt.Println("   • Single objects")
	fmt.Println("   • Highly heterogeneous data")
	fmt.Println("   • When readability > efficiency")
	fmt.Println("   • One-off data structures")
	fmt.Println()
	fmt.Println("💡 Key Advantage:")
	fmt.Println("   TRON uses same API as encoding/json!")
	fmt.Println("   Just swap: json.Marshal() → tron.Marshal()")
	fmt.Println()
}

func main() {
	useCase1_APIResponses()
	useCase2_TimeSeriesData()
	useCase3_DatabaseResults()
	useCase4_Configuration()
	useCase5_WhenJSONIsBetter()
	summary()
}
