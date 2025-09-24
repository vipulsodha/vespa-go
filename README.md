# Vespa YQL Query Builder

[![Go Reference](https://pkg.go.dev/badge/github.com/vipulsodha/vespa-go.svg)](https://pkg.go.dev/github.com/vipulsodha/vespa-go)
[![GitHub release](https://img.shields.io/github/release/vipulsodha/vespa-go.svg)](https://github.com/vipulsodha/vespa-go/releases)
[![License](https://img.shields.io/github/license/vipulsodha/vespa-go.svg)](LICENSE)

A comprehensive, type-safe query builder for Vespa AI that provides a fluent API for constructing YQL (Vespa Query Language) queries. This library replaces manual string building with an intuitive, maintainable, and error-resistant approach to query construction.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Best Practices](#best-practices)
- [Testing](#testing)

## Installation

```bash
go get github.com/vipulsodha/vespa-go@v1.0.0
```

Or for the latest version:
```bash
go get github.com/vipulsodha/vespa-go@latest
```

Then import in your Go code:
```go
import "github.com/vipulsodha/vespa-go"
```

## Quick Start

### Basic Query

```go
query, err := vespa.NewQueryBuilder().
    Select("id", "title", "price").
    From("products").
    Where(vespa.Field("price").Between(10.0, 100.0)).
    Build()

if err != nil {
    log.Fatal(err)
}

// Generated YQL: select id, title, price from sources products where ((price >= 10) and (price <= 100))
```

### Vector Search with Ranking

```go
queryVector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

query, err := vespa.NewQueryBuilder().
    Select("*").
    From("products").
    Where(vespa.Field("category").In("electronics", "gadgets")).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field("embedding_field").NearestNeighbor("query_vector", 1000)).
            AddCondition(vespa.Field("brand").Contains("nike")),
    ).
    WithInput("input.query(query_vector)", queryVector).
    WithRanking("hybrid_profile").
    Build()
```

## What's New

### 🎉 Enhanced Vector Search Capabilities

This version introduces significant improvements to make the builder more flexible and aligned with Vespa's full YQL capabilities:

#### Vector Search in WHERE Clauses
- **NEW**: Use `nearestNeighbor` directly in WHERE clauses for primary filtering
- **NEW**: Combine vector search with traditional filters using AND/OR logic
- **NEW**: Support for labels and distance thresholds in WHERE conditions

#### Flexible RANK Expressions  
- **NEW**: Add any WHERE condition to RANK expressions using `AddCondition()`
- **ENHANCED**: Mix traditional rank features with condition-based ranking
- **BACKWARD COMPATIBLE**: All existing code continues to work unchanged

#### Example of New Capabilities
```go
// Vector search as primary filter with additional ranking
query := vespa.NewQueryBuilder().
    Where(
        vespa.And(
            vespa.Field("embedding").NearestNeighbor("query_vector", 1000),
            vespa.Field("price").Between(20.0, 200.0),
        ),
    ).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field("popularity").Gte(0.5)).
            AddCondition(vespa.Field("description").Contains("premium")),
    ).
    Build()
```

## Core Concepts

### 1. Query Builder

The `QueryBuilder` interface provides a fluent API for constructing queries:

```go
type QueryBuilder interface {
    Select(fields ...string) QueryBuilder
    From(sources ...string) QueryBuilder
    Where(condition WhereCondition) QueryBuilder
    Rank(rankExpression RankExpression) QueryBuilder
    WithRanking(profile string) QueryBuilder
    WithHits(hits int) QueryBuilder
    WithOffset(offset int) QueryBuilder
    WithDefaultIndex(index string) QueryBuilder
    WithInput(key string, value interface{}) QueryBuilder
    WithQuery(query string) QueryBuilder
    Build() (*VespaQuery, error)
    BuildYQL() (string, error)
}
```

### 2. Where Conditions

The library supports comprehensive comparison operators:

| Operator | Method | Example |
|----------|---------|---------|
| `=` | `Eq(value)` | `Field("price").Eq(100)` |
| `!=` | `NotEq(value)` | `Field("status").NotEq("inactive")` |
| `>` | `Gt(value)` | `Field("price").Gt(50)` |
| `>=` | `Gte(value)` | `Field("rating").Gte(4.0)` |
| `<` | `Lt(value)` | `Field("stock").Lt(10)` |
| `<=` | `Lte(value)` | `Field("price").Lte(200)` |
| `in` | `In(values...)` | `Field("category").In("a", "b")` |
| `not in` | `NotIn(values...)` | `Field("brand").NotIn("excluded")` |
| `contains` | `Contains(value)` | `Field("description").Contains("wireless")` |
| `not contains` | `NotContains(value)` | `Field("title").NotContains("refurbished")` |
| `matches` | `Matches(pattern)` | `Field("sku").Matches("^PRD-.*")` |
| **Logical Operators** | | |
| `not` | `Not(condition)` | `Not(Field("brand").Contains("excluded"))` |
| `sameElement` | `ContainsSameElement(conditions...)` | `Field("sizes").ContainsSameElement(Field("family").Contains("clothing"), Field("size_value").Contains("M"))` |
| **Vector Search** | | |
| `nearestNeighbor` | `NearestNeighbor(vector, hits, ...opts)` | `Field("embedding").NearestNeighbor("query_vector", 1000)` |
| `nearestNeighbor` | `NearestNeighbor(vector, hits, WithLabel())` | `Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithLabel("main"))` |
| `nearestNeighbor` | `NearestNeighbor(vector, hits, WithThreshold())` | `Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithThreshold(0.8))` |
| `nearestNeighbor` | `NearestNeighbor(vector, hits, WithApproximate())` | `Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithApproximate(false))` |

### 3. Rank Expressions

Rank expressions define how documents are scored and ranked. The API uses a single unified method `AddCondition()` for all ranking elements:

```go
rank := vespa.NewRank().
    AddCondition(vespa.Field("embedding").NearestNeighbor("query_vector", 1000)).        // Vector similarity
    AddCondition(vespa.Field("brand").Contains("nike air", vespa.WithPhraseMatching())). // Field conditions
    AddCondition(vespa.UserQuery("default_index")).                                  // Text queries (query text via WithQuery())
    AddCondition(vespa.Custom("log(popularity) * 0.1"))                              // Custom expressions
```

**Key Insight**: All ranking elements (vector search, field conditions, user queries, custom expressions) implement the same `WhereCondition` interface and can be added using `AddCondition()`. This provides a clean, consistent API.

## API Reference

### Field Builders

Use the generic field builder for all fields:

```go
vespa.Field("price")      // price field
vespa.Field("rating")     // rating field  
vespa.Field("category")   // category field
vespa.Field("brand")      // brand field
vespa.Field("color")      // color field
vespa.Field("stock")      // stock field
vespa.Field("status")     // status field
vespa.Field("any_name")   // any custom field
```

### Boolean Logic

Combine conditions with AND/OR/NOT logic:

```go
// AND combination
condition := vespa.And(
    vespa.Field("price").Gte(10),
    vespa.Field("price").Lte(100),
)

// OR combination  
condition := vespa.Or(
    vespa.Field("brand").Eq("nike"),
    vespa.Field("brand").Eq("adidas"),
)

// NOT combination - negates any condition
condition := vespa.Not(
    vespa.Field("brand").Contains("excluded"),
)

// Nested logic with NOT
condition := vespa.And(
    vespa.Field("price").Between(10, 100),
    vespa.Not(
        vespa.Or(
            vespa.Field("brand").Contains("excluded1"),
            vespa.Field("brand").Contains("excluded2"),
        ),
    ),
)

// XOR logic using NOT - either condition is true, but not both
condition := vespa.And(
    vespa.Or(
        vespa.Field("brand").Contains("nike"),
        vespa.Field("item_types").Contains("shoes"),
    ),
    vespa.Not(
        vespa.And(
            vespa.Field("brand").Contains("nike"),
            vespa.Field("item_types").Contains("shoes"),
        ),
    ),
)
```

### Range Conditions

Convenient range filtering:

```go
// Price between 10 and 100
vespa.Field("price").Between(10.0, 100.0)

// Equivalent to: (price >= 10) AND (price <= 100)
```

### SameElement Conditions for Complex Fields

**NEW**: Use `sameElement` for querying arrays of structs or maps where all conditions must match within the same element:

```go
// Find products with medium clothing sizes
vespa.Field("sizes").ContainsSameElement(
    vespa.Field("family").Contains("clothing"),
    vespa.Field("size_value").In("M", "Medium"),
)

// Find persons named John Smith born after 1980
vespa.Field("persons").ContainsSameElement(
    vespa.Field("first_name").Contains("John"),
    vespa.Field("last_name").Contains("Smith"),
    vespa.Field("year_of_birth").Gt(1980),
)

// Find attributes with specific key-value pairs
vespa.Field("attributes").ContainsSameElement(
    vespa.Field("key").Eq("color"),
    vespa.Field("value").In("red", "blue", "green"),
)

// Combined with other conditions
vespa.And(
    vespa.Field("sizes").ContainsSameElement(
        vespa.Field("family").Contains("clothing"),
        vespa.Field("size_value").Contains("L"),
    ),
    vespa.Field("brand").Contains("nike"),
)
```

**Generated YQL Examples:**
```yql
(sizes contains sameElement(family contains 'clothing', size_value in ('M', 'Medium')))
(persons contains sameElement(first_name contains 'John', last_name contains 'Smith', year_of_birth > 1980))
(attributes contains sameElement(key = 'color', value in ('red', 'blue', 'green')))
((sizes contains sameElement(family contains 'clothing', size_value contains 'L')) AND (brand contains 'nike'))
```

**Why SameElement?** Without `sameElement`, conditions might match across different elements of an array, leading to false positives. For example, a query for "John Smith" might match a document with one person named "John Doe" and another named "Jane Smith".

### Vector Search in WHERE Clause

**NEW**: Use `nearestNeighbor` directly in WHERE clauses for vector-based filtering:

```go
// Basic nearestNeighbor in WHERE clause
vespa.Field("embedding").NearestNeighbor("query_vector", 1000)

// With label for query tracking
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithLabel("main_search"))

// With distance threshold
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithThreshold(0.8))

// With approximate search (true for approximate, false for exact)
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithApproximate(true))

// With exact search (more precise but slower)
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithApproximate(false))

// Combined with other conditions
vespa.And(
    vespa.Field("embedding").NearestNeighbor("query_vector", 1000),
    vespa.Field("brand").Contains("nike"),
)
```

**Generated YQL Examples:**
```yql
{targetHits:1000}nearestNeighbor(embedding, query_vector)
{targetHits:1000,label:'main_search'}nearestNeighbor(embedding, query_vector)  
{targetHits:1000,distanceThreshold:0.800000}nearestNeighbor(embedding, query_vector)
{targetHits:1000,approximate:true}nearestNeighbor(embedding, query_vector)
{targetHits:1000,approximate:false}nearestNeighbor(embedding, query_vector)
({targetHits:1000}nearestNeighbor(embedding, query_vector) AND (brand contains 'nike'))
```

#### Understanding the `approximate` Parameter

The `approximate` parameter controls whether Vespa performs **approximate** or **exact** nearest neighbor search:

- **`approximate:true`** (Default for HNSW indexes): Uses approximate search algorithms for better performance
- **`approximate:false`**: Uses exact search for maximum precision but with higher computational cost

**When to use approximate vs exact search:**

```go
// Use approximate search for:
// - Real-time applications requiring fast response times
// - Large-scale vector databases where slight precision trade-offs are acceptable
// - Most production search scenarios
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithApproximate(true))

// Use exact search for:
// - High-precision requirements where accuracy is critical
// - Smaller datasets where performance isn't a bottleneck
// - Research or validation scenarios requiring perfect results
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithApproximate(false))

// Combining with other parameters for exact search with filtering
vespa.Field("embedding").NearestNeighbor("query_vector", 1000,
    vespa.WithApproximate(false),
    vespa.WithThreshold(0.95),
    vespa.WithLabel("high_precision_search"),
)
```

**Performance vs Accuracy Trade-off:**
- **Approximate**: ~10-100x faster, 95-99% accuracy
- **Exact**: Slower but 100% accurate results

### Rank Features

#### Nearest Neighbor Search

The query builder provides `NearestNeighbor` functionality that works in both WHERE clauses and RANK expressions using the Field-based API with functional options:

```go
// Basic nearest neighbor (works in both WHERE and RANK)
vespa.Field("embedding_field").NearestNeighbor("query_vector", 1000)

// With functional options
vespa.Field("embedding_field").NearestNeighbor("query_vector", 1000, 
    vespa.WithLabel("main_search"),
    vespa.WithThreshold(0.8),
    vespa.WithApproximate(false),
)

// In RANK expressions
rank.AddCondition(vespa.Field("embedding_field").NearestNeighbor("query_vector", 1000, vespa.WithLabel("main_search")))

// In WHERE clauses  
vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithThreshold(0.8))
```

#### Flexible Rank Expressions

**NEW**: Use any WHERE condition within rank expressions:

```go
rank := vespa.NewRank().
    // All rank features and conditions use the same AddCondition method
    AddCondition(vespa.Field("embedding").NearestNeighbor("query_vector", 1000)).
    AddCondition(vespa.Field("brand").Contains("nike")).
    AddCondition(vespa.Field("price").Gte(50.0)).
    AddCondition(vespa.Field("category").In("electronics", "gadgets"))

// Generated: rank({targetHits:1000}nearestNeighbor(embedding, query_vector), (brand contains 'nike'), (price >= 50), (category in ('electronics', 'gadgets')))
```

#### Text Matching

```go
// Exact match using conditions
rank.AddCondition(vespa.Field("color").Contains("red"))

// Phrase matching for multiple keywords
rank.AddCondition(vespa.Field("brand").Contains("nike air max", vespa.WithPhraseMatching()))

// User query (query text passed via WithQuery())
rank.AddCondition(vespa.UserQuery("default_index"))
```

#### UserQuery Function

**NEW**: The `UserQuery()` function provides flexible text search capabilities that can be used in both WHERE clauses and rank expressions.

**Key Features:**
- **Standalone WHERE condition**: Use `userQuery()` as a primary filter
- **Rank expression**: Use `userQuery()` for text-based ranking
- **Flexible defaultIndex**: Optional field specification for text search
- **Separate query text**: Query text passed via `WithQuery()` method

**Function Signature:**
```go
func UserQuery(defaultIndex ...string) WhereCondition
```

**Usage Examples:**

```go
// 1. Basic userQuery without default index
vespa.UserQuery()  // Generates: userQuery()

// 2. UserQuery with specific field/index
vespa.UserQuery("title")  // Generates: {defaultIndex:"title"}userQuery()

// 3. UserQuery in WHERE clause
query := vespa.NewQueryBuilder().
    Select("id", "title", "price").
    From("products").
    Where(vespa.UserQuery("title")).         // YQL: {defaultIndex:"title"}userQuery()
    WithQuery("wireless headphones").        // HTTP: {"query": "wireless headphones"}
    Build()

// 4. UserQuery combined with other conditions
query := vespa.NewQueryBuilder().
    Select("*").
    From("products").
    Where(
        vespa.And(
            vespa.UserQuery("description"),  // Text search in description field
            vespa.Field("price").Between(20, 100),
            vespa.Field("category").Eq("electronics"),
        ),
    ).
    WithQuery("bluetooth speaker").          // The actual search text
    Build()

// 5. UserQuery in rank expressions
query := vespa.NewQueryBuilder().
    Select("id", "title", "score").
    From("products").
    Where(vespa.Field("category").In("electronics", "gadgets")).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.UserQuery("title")).      // Text relevance ranking
            AddCondition(vespa.Field("brand").Contains("premium")),
    ).
    WithQuery("gaming headset").             // The actual search text
    WithRanking("text_ranking").
    Build()
```

**Generated Output:**
```json
{
  "yql": "select id, title, score from sources products where (category in ('electronics', 'gadgets')) and rank({defaultIndex:\"title\"}userQuery(), (brand contains 'premium'))",
  "query": "gaming headset",
  "ranking": "text_ranking"
}
```

**Important Notes:**
- Query text is passed via `WithQuery()`, not as a parameter to `UserQuery()`
- The YQL contains `userQuery()` or `{defaultIndex:"field"}userQuery()`
- The HTTP request contains the actual query text in the `"query"` parameter
- Works seamlessly in both WHERE clauses and rank expressions

#### Custom Features

```go
// Add custom ranking expressions
rank.AddCondition(vespa.Custom("log(popularity) * 0.1"))
```

### Pagination Support

**NEW**: The query builder now supports pagination through the `WithOffset()` method, enabling easy result paging:

```go
// Basic pagination - page 1 (first 20 results)
query := vespa.NewQueryBuilder().
    Select("id", "title", "price").
    From("products").
    Where(vespa.Field("category").Eq("electronics")).
    WithHits(20).
    Build()

// Page 2 (results 21-40)
query := vespa.NewQueryBuilder().
    Select("id", "title", "price").
    From("products").
    Where(vespa.Field("category").Eq("electronics")).
    WithHits(20).
    WithOffset(20).
    Build()

// Page 3 (results 41-60)
query := vespa.NewQueryBuilder().
    Select("id", "title", "price").
    From("products").
    Where(vespa.Field("category").Eq("electronics")).
    WithHits(20).
    WithOffset(40).
    Build()
```

**Pagination Helper Pattern:**
```go
func BuildPagedQuery(page int, itemsPerPage int) (*vespa.VespaQuery, error) {
    offset := (page - 1) * itemsPerPage
    
    return vespa.NewQueryBuilder().
        Select("id", "title", "price", "brand").
        From("products").
        Where(vespa.Field("active").Eq(true)).
        WithHits(itemsPerPage).
        WithOffset(offset).
        WithRanking("relevance").
        Build()
}

// Usage:
query, err := BuildPagedQuery(3, 25) // Get page 3 with 25 items per page (offset=50)
```

**Generated Query Parameters:**
- `hits`: Controls the number of results returned (equivalent to SQL `LIMIT`)
- `offset`: Controls the number of results to skip (equivalent to SQL `OFFSET`)

**Example Result:**
```json
{
  "yql": "select id, title, price from sources products where (category contains 'electronics')",
  "hits": 20,
  "offset": 40,
  "ranking": "bm25"
}
```

### Input Parameters

Handle query vectors and other input parameters:

```go
queryVector := []float32{0.1, 0.2, 0.3}
themeVector := []float32{0.4, 0.5, 0.6}

builder.
    WithInput("input.query(query_vector)", queryVector).
    WithInput("input.query(sort_vector)", themeVector)
```

## Examples

### 1. Simple Product Search

```go
query, err := vespa.NewQueryBuilder().
    Select("id", "title", "price", "brand").
    From("products").
    Where(
        vespa.And(
            vespa.And(
                vespa.Field("price").Between(20.0, 200.0),
                vespa.Field("category").Eq("electronics"),
            ),
            vespa.Field("stock").Gt(0),
        ),
    ).
    WithHits(20).
    Build()
```

### 2. Hybrid Vector + Text Search

```go
queryVector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

query, err := vespa.NewQueryBuilder().
    Select("*").
    From("products").
    Where(vespa.Field("category").In("electronics", "gadgets")).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field("embedding").NearestNeighbor("query_vector", 1000, vespa.WithLabel("semantic"))).
            AddCondition(vespa.UserQuery("text_field")).
            AddCondition(vespa.Field("brand").Contains("sony")),
    ).
    WithInput("input.query(query_vector)", queryVector).
    WithQuery("wireless bluetooth headphones").
    WithRanking("hybrid_semantic_text").
    WithHits(50).
    Build()
```

### 2.1 High-Precision Vector Search

```go
// Example: Research application requiring exact results
queryVector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

query, err := vespa.NewQueryBuilder().
    Select("id", "title", "similarity_score", "research_metadata").
    From("research_papers").
    Where(
        vespa.And(
            // Exact vector search for maximum precision
            vespa.Field("paper_embedding").NearestNeighbor("query_vector", 500, 
                vespa.WithApproximate(false),
                vespa.WithThreshold(0.90),
                vespa.WithLabel("exact_search"),
            ),
            vespa.Field("peer_reviewed").Eq(true),
        ),
    ).
    WithInput("input.query(query_vector)", queryVector).
    WithRanking("academic_relevance").
    WithHits(20).
    Build()

// Generated query uses exact search: {targetHits:500,distanceThreshold:0.900000,approximate:false,label:'exact_search'}nearestNeighbor(paper_embedding, query_vector)
```

### 2.2. Exclusion and Negation Queries

```go
// Example: E-commerce filters with exclusions
query, err := vespa.NewQueryBuilder().
    Select("id", "title", "price", "brand", "category").
    From("products").
    Where(
        vespa.And(
            vespa.Field("category").Eq("electronics"),
            vespa.Field("price").Between(50.0, 500.0),
            // Exclude specific brands
            vespa.Not(
                vespa.Or(
                    vespa.Field("brand").Contains("excluded_brand1"),
                    vespa.Field("brand").Contains("excluded_brand2"),
                ),
            ),
            // Exclude discontinued products
            vespa.Not(vespa.Field("status").Eq("discontinued")),
        ),
    ).
    WithHits(30).
    Build()

// Generated YQL: select id, title, price, brand, category from sources products 
// where ((category contains 'electronics') AND ((price >= 50) and (price <= 500)) 
//        AND !(((brand contains 'excluded_brand1') OR (brand contains 'excluded_brand2'))) 
//        AND !((status contains 'discontinued')))
```

### 2.3. XOR Logic for Mutually Exclusive Conditions  

```go
// Example: Find products with either nike brand OR shoes category, but not both
// This is useful for recommendation systems or A/B testing scenarios
query, err := vespa.NewQueryBuilder().
    Select("brand", "item_types", "title", "price").
    From("listings_sg_v1").
    Where(
        vespa.And(
            // At least one condition is true
            vespa.Or(
                vespa.Field("brand").Contains("nike"),
                vespa.Field("item_types").Contains("shoes"),
            ),
            // But not both are true (XOR logic)
            vespa.Not(
                vespa.And(
                    vespa.Field("brand").Contains("nike"),
                    vespa.Field("item_types").Contains("shoes"),
                ),
            ),
        ),
    ).
    WithHits(25).
    Build()

// Generated YQL: select brand, item_types, title, price from sources listings_sg_v1 
// where (((brand contains 'nike') OR (item_types contains 'shoes')) 
//        AND !(((brand contains 'nike') AND (item_types contains 'shoes'))))

// This will match:
// - Nike shirts (nike brand, non-shoes category) ✓
// - Adidas shoes (non-nike brand, shoes category) ✓  
// But exclude:
// - Nike shoes (nike brand AND shoes category) ✗
```

### 3. NEW: Vector Search in WHERE with Ranking

**This demonstrates the new capability to use nearestNeighbor directly in WHERE clauses:**

```go
queryVector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

query, err := vespa.NewQueryBuilder().
    Select("id", "title", "price", "brand").
    From("products").
    Where(
        vespa.And(
            // Vector search as primary filter
            vespa.Field("embedding").NearestNeighbor("query_vector", 1000),
            // Combined with traditional filters
            vespa.Field("price").Between(20.0, 200.0),
            vespa.Field("brand").Contains("nike"),
        ),
    ).
    // Additional ranking on top of the filtered results
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field("popularity").Gte(0.5)).
            AddContains("description", "premium"),
    ).
    WithInput("input.query(query_vector)", queryVector).
    WithRanking("popularity_boost").
    WithHits(50).
    Build()

// Generated YQL:
// select id, title, price, brand from sources products 
// where ({targetHits:1000}nearestNeighbor(embedding, query_vector) AND ((price >= 20) and (price <= 200)) AND (brand contains 'nike')) 
// rank((popularity >= 0.5), description contains 'premium')
```

### 4. NEW: SameElement for Complex Fields

**This demonstrates querying arrays of structs or maps with sameElement:**

```go
// Find products with specific size characteristics and attributes
query, err := vespa.NewQueryBuilder().
    Select("id", "name", "sizes", "attributes").
    From("products").
    Where(
        vespa.And(
            // Find products with medium clothing sizes
            vespa.Field("sizes").ContainsSameElement(
                vespa.Field("family").Contains("clothing"),
                vespa.Field("size_value").In("M", "Medium"),
            ),
            // Find products with red color attribute  
            vespa.Field("attributes").ContainsSameElement(
                vespa.Field("key").Eq("color"),
                vespa.Field("value").Contains("red"),
            ),
            // Regular filters
            vespa.Field("brand").Contains("nike"),
            vespa.Field("price").Between(50.0, 200.0),
        ),
    ).
    WithHits(25).
    Build()

if err != nil {
    log.Fatal(err)
}

// Generated YQL:
// select id, name, sizes, attributes from sources products 
// where ((sizes contains sameElement(family contains 'clothing', size_value in ('M', 'Medium'))) 
//        AND (attributes contains sameElement(key = 'color', value contains 'red')) 
//        AND (brand contains 'nike') 
//        AND ((price >= 50) and (price <= 200)))
```

**Real-world use case**: This query ensures that:
- The "M" or "Medium" size is specifically for "clothing" (not shoes, accessories, etc.)
- The "red" color is a defined attribute (not just mentioned in description)
- Avoids false matches where conditions span different array elements

### 5. Complex E-commerce Query

```go
// Complex e-commerce search with vector and text filtering
queryVector := []float32{0.1, 0.2, 0.3}
categories := []string{"fashion", "accessories"}
sizeFilter := []string{"M", "L", "XL"}

query, err := vespa.NewQueryBuilder().
    Select("id", "title", "price", "brand", "category").
    From("products").
    Where(
        vespa.And(
            vespa.Field("category").In(categories...),
            vespa.Field("size").In(sizeFilter...),
            vespa.Field("price").Between(25.0, 150.0),
        ),
    ).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field("embedding_field").NearestNeighbor("query_vector", 1000)).
            AddCondition(vespa.UserQuery("text_field")).
            AddCondition(vespa.Field("brand").Contains("nike")).
            AddCondition(vespa.Field("color").Contains("blue")),
    ).
    WithInput("input.query(query_vector)", queryVector).
    WithQuery("summer casual wear").
    Build()

if err != nil {
    log.Fatal(err)
}

// The generated query includes:
// - Vector search with query vector
// - Category, size, and price filters  
// - Text search with user query
// - Brand and color ranking features
```

### 6. Advanced Filtering with Custom Logic

```go
query, err := vespa.NewQueryBuilder().
    Select("id", "title", "price", "rating", "brand").
    From("products").
    Where(
        vespa.And(
            vespa.And(
                vespa.And(
                    vespa.Field("price").Between(10.0, 500.0),
                    vespa.Or(
                        vespa.And(
                            vespa.Field("brand").In("premium1", "premium2"),
                            vespa.Field("rating").Gte(4.5),
                        ),
                        vespa.And(
                            vespa.And(
                                vespa.Field("brand").In("budget1", "budget2"),  
                                vespa.Field("rating").Gte(4.0),
                            ),
                            vespa.Field("price").Lte(100.0),
                        ),
                    ),
                ),
                vespa.Field("stock").Gt(0),
            ),
            vespa.Field("status").Eq("active"),
        ),
    ).
    WithHits(30).
    Build()
```

### 7. Multi-Vector Semantic Search

```go
queryVector := []float32{/* main embedding */}
imageVector := []float32{/* image embedding */}
colorVector := []float32{/* color embedding */}

query, err := vespa.NewQueryBuilder().
    Select("id", "title", "image_url", "price").
    From("visual_products").
    Where(vespa.Field("category").In("fashion", "accessories")).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field("text_embedding").NearestNeighbor("query_vector", 1000, vespa.WithLabel("text_sim"))).
            AddCondition(vespa.Field("image_embedding").NearestNeighbor("image_vector", 500, vespa.WithLabel("visual_sim"))).
            AddCondition(vespa.Field("color_embedding").NearestNeighbor("color_vector", 200, vespa.WithLabel("color_sim"))),
    ).
    WithInput("input.query(query_vector)", queryVector).
    WithInput("input.query(image_vector)", imageVector).
    WithInput("input.query(color_vector)", colorVector).
    WithRanking("multi_modal_ranking").
    Build()
```

### 8. Pagination Example

```go
// Realistic e-commerce pagination scenario
func SearchProductsWithPagination(category string, page int, itemsPerPage int) (*vespa.VespaQuery, error) {
    // Calculate offset for the requested page
    offset := (page - 1) * itemsPerPage
    
    query, err := vespa.NewQueryBuilder().
        Select("id", "title", "price", "image_url", "rating").
        From("products").
        Where(
            vespa.And(
                vespa.Field("category").Contains(category),
                vespa.Field("active").Eq(true),
                vespa.Field("stock").Gt(0),
            ),
        ).
        WithHits(itemsPerPage).
        WithOffset(offset).
        WithRanking("popularity_boost").
        Build()
    
    return query, err
}

// Usage examples:
// Page 1: results 1-20
query1, _ := SearchProductsWithPagination("electronics", 1, 20) // offset=0, hits=20

// Page 3: results 41-60  
query3, _ := SearchProductsWithPagination("electronics", 3, 20) // offset=40, hits=20

// Large page: results 981-1000
queryLarge, _ := SearchProductsWithPagination("electronics", 50, 20) // offset=980, hits=20
```

**Generated queries show pagination parameters:**
- Page 1: `{"hits": 20, "offset": 0}`
- Page 3: `{"hits": 20, "offset": 40}`
- Page 50: `{"hits": 20, "offset": 980}`

## Helper Functions

### Input Parameters

Handle query vectors and other input parameters:

```go
queryVector := []float32{0.1, 0.2, 0.3}

builder.WithInput("input.query(query_vector)", queryVector)
```

## Best Practices

### 1. Use Type-Safe Builders

```go
// ✅ Good: Use field builders
vespa.Field("price").Gte(10.0)

// ❌ Avoid: Manual string building  
"price >= 10.0"
```

### 2. Use Range Methods for Convenience

```go
// ✅ Good: Use Between for ranges
condition := vespa.Field("price").Between(10.0, 100.0)

// ❌ Avoid: Manual range building
condition := vespa.And(
    vespa.Field("price").Gte(10.0),
    vespa.Field("price").Lte(100.0),
)
```

### 3. Use Structured Input Parameters

```go
// ✅ Good: Direct parameter assignment
builder.WithInput("input.query(query_vector)", vector)
```

### 4. Organize Complex Conditions

```go
// ✅ Good: Break down complex logic
priceCondition := vespa.Field("price").Between(10, 100)
categoryCondition := vespa.Field("category").In("electronics", "gadgets") 
stockCondition := vespa.Field("stock").Gt(0)

finalCondition := vespa.And(priceCondition, categoryCondition, stockCondition)

// ❌ Avoid: Deeply nested inline conditions
```



## Testing

Run the test suite:

```bash
go test ./internal/lib/vespa/...
```

Key test coverage includes:

- ✅ All comparison operators (eq, neq, gt, gte, lt, lte, in, not in, contains, matches)
- ✅ Boolean logic combinations (AND, OR, nested conditions)
- ✅ Range conditions and field builders
- ✅ SameElement conditions for complex fields (arrays of structs/maps)
- ✅ Rank expressions and features  
- ✅ Nearest neighbor search with all parameters (label, distanceThreshold, approximate)
- ✅ Approximate vs exact vector search functionality
- ✅ Text matching (exact, phrase, fuzzy)
- ✅ Complete query building and validation
- ✅ Helper functions and convenience methods
- ✅ Input parameter handling
- ✅ Error cases and validation failures

## Error Handling

The library provides detailed validation errors:

```go
type ValidationError struct {
    Field   string
    Message string  
}

// Example usage
query, err := builder.Build()
if err != nil {
    if validationErr, ok := err.(*vespa.ValidationError); ok {
        log.Printf("Validation error in field '%s': %s", 
            validationErr.Field, validationErr.Message)
    }
    return err
}
```

## Migration from Manual String Building

### Before (Manual String Building)

```go
yqlBuilder := strings.Builder{}
yqlBuilder.WriteString(fmt.Sprintf("select color, brand, price from sources %s where ", indexName))

if len(categories) > 0 {
    categoryFilter := buildCategoryFilter(categories)
    yqlBuilder.WriteString(fmt.Sprintf("(%s) and ", categoryFilter))
}

if priceFilter.Min > 0 {
    yqlBuilder.WriteString(fmt.Sprintf("(price >= %f) and ", priceFilter.Min))
}

yqlBuilder.WriteString(fmt.Sprintf("rank(({targetHits:%d}nearestNeighbor(%s, query_vector))", 
    1000, retrievalFieldName))
// ... more manual building
```

### After (Query Builder)

```go
query, err := vespa.NewQueryBuilder().
    Select("color", "brand", "price").
    From(indexName).
    Where(
        vespa.And(
            vespa.Field("category").In(categories...),
            vespa.Field("price").Between(priceFilter.Min, priceFilter.Max),
        ),
    ).
    Rank(
        vespa.NewRank().
            AddCondition(vespa.Field(retrievalFieldName).NearestNeighbor("query_vector", 1000)),
    ).
    Build()
```

## Performance Considerations

- **Memory**: Query builders reuse internal slices and maps efficiently  
- **String Building**: YQL generation uses efficient string building techniques
- **Type Safety**: Compile-time type checking prevents runtime errors

## Contributing

When adding new features:

1. Add corresponding types in `types.go`
2. Implement functionality in appropriate files (`conditions.go`, `rank.go`, etc.)  
3. Add helper functions in `helpers.go`
4. Include validation in `validation.go`
5. Write comprehensive tests in `builder_test.go`
6. Update this README with examples

---

This query builder transforms complex, error-prone string manipulation into clean, maintainable, and type-safe Go code. It provides the full power of Vespa's YQL while ensuring correctness and readability.
