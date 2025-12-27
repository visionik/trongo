# TRON Examples

This directory contains examples comparing TRON to JSON encoding.

## Quick Start

```bash
# Run the quickstart example
go run quickstart.go

# See detailed comparisons
go run comparison.go

# Explore real-world use cases
go run use_cases.go
```

## API Comparison

### JSON
```go
import "encoding/json"

// Marshal
data, err := json.Marshal(myStruct)

// Unmarshal
var result MyStruct
err = json.Unmarshal(data, &result)
```

### TRON (Identical API!)
```go
import "github.com/tron-format/trongo/pkg/tron"

// Marshal - same function signature
data, err := tron.Marshal(myStruct)

// Unmarshal - same function signature
var result MyStruct
err = tron.Unmarshal(data, &result)
```

## Format Comparison

### Example: Array of Structs

**JSON** (79 bytes):
```json
[{"name":"Alice","age":30},{"name":"Bob","age":25},{"name":"Charlie","age":35}]
```

**TRON** (62 bytes - 21.5% smaller):
```
class A: name,age

[A("Alice",30),A("Bob",25),A("Charlie",35)]
```

### Why TRON?

**✅ Token Efficiency:**
- Reduces redundancy by defining structure once
- Perfect for arrays of similar objects
- Significant savings: 20-50% for typical use cases

**✅ Same API as JSON:**
- Drop-in replacement for `encoding/json`
- Same struct tags (`json:"name,omitempty"`)
- Same behavior and error handling

**✅ Great For:**
- API responses with repeated structure
- Database query results
- Time-series/IoT data
- Any data where bandwidth/tokens matter
- LLM applications (lower token costs!)

**📄 Stick with JSON for:**
- Single objects
- Highly heterogeneous data
- When readability trumps efficiency

## Examples Overview

### 1. quickstart.go
Basic usage showing Marshal and Unmarshal operations.

### 2. comparison.go
Side-by-side comparisons of TRON vs JSON for:
- Simple structs
- Arrays of structs
- Nested structures
- Maps and primitives
- Struct tags and omitempty
- API responses
- Round-trip conversion

### 3. use_cases.go
Real-world scenarios:
- Product catalogs (37.7% savings)
- Time-series data (42.4% savings)
- Database results (41.5% savings)
- Configuration files (50.0% savings)

## Running Examples

```bash
cd examples

# Run individual examples
go run quickstart.go
go run comparison.go
go run use_cases.go

# Or run all at once
for f in *.go; do
    echo "=== Running $f ==="
    go run $f
    echo ""
done
```

## Migration from JSON

Migrating from JSON to TRON is trivial:

**Before:**
```go
import "encoding/json"

data, _ := json.Marshal(users)
json.Unmarshal(data, &result)
```

**After:**
```go
import "github.com/tron-format/trongo/pkg/tron"

data, _ := tron.Marshal(users)  // Same API!
tron.Unmarshal(data, &result)   // Same API!
```

That's it! Your struct tags, error handling, and everything else stays the same.
