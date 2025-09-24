package main

import (
	"encoding/json"
	"fmt"
	"log"

	vespa "github.com/vipulsodha/vespa-go"
)

func main() {
	fmt.Println("Complex Query Examples")
	fmt.Println("======================")

	// Example 1: SameElement queries for complex data structures
	sameElementQueries()

	// Example 2: XOR logic for mutually exclusive conditions
	xorLogicQuery()

	// Example 3: Advanced filtering with exclusions
	advancedFiltering()

	// Example 4: Pagination scenarios
	paginationExamples()

	// Example 5: E-commerce recommendation system
	ecommerceRecommendation()
}

func sameElementQueries() {
	fmt.Println("\n1. SameElement Queries:")

	// Query for products with specific size characteristics
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
				// Find products with specific color attributes
				vespa.Field("attributes").ContainsSameElement(
					vespa.Field("key").Eq("color"),
					vespa.Field("value").In("red", "blue", "green"),
				),
				vespa.Field("brand").Contains("nike"),
			),
		).
		WithHits(25).
		Build()

	if err != nil {
		log.Fatalf("Error building sameElement query: %v", err)
	}

	printQuery(query)
}

func xorLogicQuery() {
	fmt.Println("\n2. XOR Logic Query:")

	// Find products with either Nike brand OR shoes category, but not both
	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "brand", "item_types").
		From("listings").
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
		WithHits(30).
		Build()

	if err != nil {
		log.Fatalf("Error building XOR query: %v", err)
	}

	printQuery(query)
}

func advancedFiltering() {
	fmt.Println("\n3. Advanced Filtering with Exclusions:")

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "price", "brand", "category").
		From("products").
		Where(
			vespa.And(
				vespa.Field("category").Eq("electronics"),
				vespa.Field("price").Between(50.0, 500.0),
				// Exclude specific problematic brands
				vespa.Not(
					vespa.Or(
						vespa.Field("brand").Contains("excluded_brand1"),
						vespa.Field("brand").Contains("excluded_brand2"),
					),
				),
				// Exclude discontinued products
				vespa.Not(vespa.Field("status").Eq("discontinued")),
				// Must have good ratings
				vespa.Field("rating").Gte(4.0),
			),
		).
		WithHits(40).
		Build()

	if err != nil {
		log.Fatalf("Error building advanced filtering query: %v", err)
	}

	printQuery(query)
}

func paginationExamples() {
	fmt.Println("\n4. Pagination Examples:")

	// Helper function to build paginated queries
	buildPagedQuery := func(page int, itemsPerPage int) (*vespa.VespaQuery, error) {
		offset := (page - 1) * itemsPerPage

		return vespa.NewQueryBuilder().
			Select("id", "title", "price", "brand", "rating").
			From("products").
			Where(
				vespa.And(
					vespa.Field("category").In("electronics", "gadgets"),
					vespa.Field("active").Eq(true),
					vespa.Field("stock").Gt(0),
				),
			).
			WithHits(itemsPerPage).
			WithOffset(offset).
			WithRanking("popularity_boost").
			Build()
	}

	// Page 1 (results 1-20)
	page1Query, err := buildPagedQuery(1, 20)
	if err != nil {
		log.Fatalf("Error building page 1 query: %v", err)
	}
	fmt.Println("Page 1 Query:")
	printQuery(page1Query)

	// Page 3 (results 41-60)
	page3Query, err := buildPagedQuery(3, 20)
	if err != nil {
		log.Fatalf("Error building page 3 query: %v", err)
	}
	fmt.Println("\nPage 3 Query:")
	printQuery(page3Query)
}

func ecommerceRecommendation() {
	fmt.Println("\n5. E-commerce Recommendation System:")

	// Complex recommendation query combining multiple signals
	queryVector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
	userPreferences := []string{"electronics", "gadgets", "smartphones"}

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "price", "brand", "category", "similarity_score").
		From("products").
		Where(
			vespa.And(
				// User's preferred categories
				vespa.Field("category").In(userPreferences...),
				// Available and in stock
				vespa.Field("available").Eq(true),
				vespa.Field("stock").Gt(0),
				// Good quality products only
				vespa.Field("rating").Gte(3.5),
				// Reasonable price range
				vespa.Field("price").Between(10.0, 1000.0),
				// Exclude items user already bought
				vespa.Not(
					vespa.Field("purchased_by_user").Contains("current_user_id"),
				),
			),
		).
		Rank(
			vespa.NewRank().
				// Semantic similarity based on user's browsing history
				AddCondition(vespa.Field("product_embedding").NearestNeighbor("user_preference_vector", 1000,
					vespa.WithLabel("semantic_similarity"))).
				// Text relevance for user's search history
				AddCondition(vespa.UserQuery("description")).
				// Boost popular products
				AddCondition(vespa.Field("popularity_score").Gte(0.6)).
				// Boost products with good reviews
				AddCondition(vespa.Field("review_sentiment").Gte(0.7)).
				// Custom business logic
				AddCondition(vespa.Custom("freshness(timestamp) * 0.1 + margin_boost * 0.2")),
		).
		WithInput("input.query(user_preference_vector)", queryVector).
		WithQuery("high quality electronics with good reviews").
		WithRanking("recommendation_profile").
		WithHits(50).
		WithOffset(0).
		Build()

	if err != nil {
		log.Fatalf("Error building recommendation query: %v", err)
	}

	printQuery(query)
}

func printQuery(query *vespa.VespaQuery) {
	jsonBytes, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		log.Printf("Error marshaling query: %v", err)
		return
	}
	fmt.Printf("%s\n", string(jsonBytes))
}
