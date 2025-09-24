# Examples

This directory contains practical examples demonstrating various features of the vespa-go library.

## Running the Examples

Each example is a Go program that demonstrates specific functionality. Run them from the project root:

```bash
# Basic search operations
go run ./examples/basic_search

# Vector search capabilities  
go run ./examples/vector_search

# Complex query patterns
go run ./examples/complex_queries
```

Or from within each example directory:
```bash
cd examples/basic_search && go run main.go
```

## Example Categories

### 1. Basic Search (`basic_search/`)
- Simple product filtering
- Price range queries
- Boolean logic combinations
- Text matching operations

**Key Features Demonstrated:**
- Field-based conditions
- AND/OR combinations
- Range filtering with `Between()`
- NOT conditions for exclusions

### 2. Vector Search (`vector_search/`)
- Nearest neighbor search
- Hybrid vector + text search
- Multi-vector scenarios
- Approximate vs exact search

**Key Features Demonstrated:**
- `NearestNeighbor()` in WHERE clauses
- Vector search with filtering
- Rank expressions with vectors
- Functional options (`WithLabel`, `WithThreshold`, `WithApproximate`)

### 3. Complex Queries (`complex_queries/`)
- SameElement conditions for complex data
- XOR logic patterns
- Advanced filtering scenarios
- Pagination examples
- E-commerce recommendation system

**Key Features Demonstrated:**
- `ContainsSameElement()` for arrays of structs
- Complex boolean logic
- Pagination with `WithOffset()`
- Real-world recommendation queries

## Understanding the Output

Each example prints the generated Vespa query as JSON, showing:
- `yql`: The generated YQL query string
- `input`: Vector parameters and other inputs
- `query`: Text query for user queries
- `ranking`: Ranking profile name
- `hits`: Number of results to return
- `offset`: Pagination offset

Example output:
```json
{
  "yql": "select id, title, price from sources products where ((price >= 10) and (price <= 100))",
  "hits": 20,
  "ranking": "bm25"
}
```

## Customizing Examples

You can modify these examples to:
- Change field names to match your schema
- Adjust vector dimensions and values
- Modify filtering logic
- Test different ranking profiles
- Experiment with pagination parameters

## Production Usage

These examples generate valid Vespa queries that can be sent directly to a Vespa cluster. In production:

1. Marshal the query to JSON
2. Send as HTTP POST to `/search/`
3. Parse the Vespa response

```go
queryJSON, _ := json.Marshal(query)
// Send queryJSON to Vespa cluster
```

## Additional Resources

- [Main README](../README.md) - Comprehensive API documentation
- [Vespa Documentation](https://docs.vespa.ai/) - Official Vespa docs
- [YQL Reference](https://docs.vespa.ai/en/reference/query-language-reference.html) - YQL syntax guide
