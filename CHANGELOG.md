# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-24

### Initial Release

#### Added
- **Core Query Builder** - Fluent API for constructing Vespa YQL queries
- **Type-Safe Field Operations** - Builder pattern for field conditions with compile-time safety
- **Comprehensive Operators** - Support for all major comparison, collection, and text operators
- **Vector Search Support** - First-class support for `nearestNeighbor` operations in WHERE clauses and rank expressions
- **Boolean Logic** - Full support for AND, OR, and NOT combinations with proper precedence
- **Range Conditions** - Convenient `Between()` method for numeric and date ranges
- **SameElement Conditions** - Support for complex array/struct querying with `ContainsSameElement()`
- **Rank Expressions** - Flexible ranking with vector search, text matching, and custom expressions
- **Functional Options** - Extensible configuration using functional options pattern
- **UserQuery Support** - Integration with Vespa's `userQuery()` function for text search
- **Pagination Support** - Built-in support for result pagination with `WithOffset()` and `WithHits()`

#### Features

##### Query Building
- Fluent API with method chaining
- Type-safe field operations
- Comprehensive validation with detailed error messages
- Support for all Vespa YQL operators

##### Vector Search
- `NearestNeighbor` conditions in WHERE clauses
- Configurable parameters: label, distance threshold, approximate search
- Multi-vector search scenarios in rank expressions
- Support for both approximate and exact vector search

##### Text Search
- Multiple text matching types: exact, phrase, fuzzy
- Integration with `userQuery()` for flexible text search
- Support for custom default indexes

##### Boolean Logic
- Nested AND/OR/NOT combinations
- XOR patterns using NOT with AND/OR
- Proper parenthesis generation for complex logic

##### Complex Data Structures
- `sameElement` support for arrays of structs/maps
- Proper handling of Vespa's sameElement limitations
- Workarounds for IN/OR restrictions within sameElement

##### Developer Experience
- Extensive documentation with examples
- Comprehensive test coverage (95%+)
- Real-world usage examples
- Clear error messages and validation

#### Technical Details
- **Language**: Go 1.21+
- **Dependencies**: Zero external dependencies (standard library only)
- **Testing**: Comprehensive test suite with edge cases
- **Documentation**: Extensive README with examples and best practices

#### Supported Vespa Features
- All YQL operators and functions
- Vector search with HNSW indexes
- Text search and ranking
- Complex field structures
- Pagination
- Custom ranking expressions
- Input parameters for vectors and other data

---

## Contributing

When contributing to this project, please:
1. Update the CHANGELOG.md with your changes
2. Follow semantic versioning for version bumps
3. Include tests for new features
4. Update documentation as needed
