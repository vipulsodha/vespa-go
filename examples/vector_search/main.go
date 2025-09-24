package main

import (
	"encoding/json"
	"fmt"
	"log"

	vespa "github.com/vipulsodha/vespa-go"
)

func main() {
	fmt.Println("Vector Search Examples")
	fmt.Println("======================")

	// Example 1: Basic vector search in WHERE clause
	basicVectorSearch()

	// Example 2: Vector search with filtering
	vectorSearchWithFilters()

	// Example 3: Hybrid search (vector + rank expression)
	hybridSearch()

	// Example 4: Multi-vector search
	multiVectorSearch()

	// Example 5: Approximate vs Exact search
	approximateVsExactSearch()
}

func basicVectorSearch() {
	fmt.Println("\n1. Basic Vector Search:")

	// Sample query vector (in practice, this would come from an embedding model)
	queryVector := []float32{0.1, 0.2, 0.3, 0.4, 0.5}

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "embedding_score").
		From("products").
		Where(
			vespa.Field("embedding").NearestNeighbor("query_vector", 1000),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithHits(50).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func vectorSearchWithFilters() {
	fmt.Println("\n2. Vector Search with Filters:")

	queryVector := []float32{0.2, 0.3, 0.4, 0.5, 0.6}

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "price", "category").
		From("products").
		Where(
			vespa.And(
				vespa.Field("embedding").NearestNeighbor("query_vector", 1000,
					vespa.WithLabel("main_search"),
					vespa.WithThreshold(0.7),
				),
				vespa.Field("category").In("electronics", "gadgets"),
				vespa.Field("price").Between(20.0, 500.0),
			),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithHits(30).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func hybridSearch() {
	fmt.Println("\n3. Hybrid Search (Vector + Text Ranking):")

	queryVector := []float32{0.3, 0.4, 0.5, 0.6, 0.7}

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "description", "brand").
		From("products").
		Where(vespa.Field("category").Eq("fashion")).
		Rank(
			vespa.NewRank().
				AddCondition(vespa.Field("embedding").NearestNeighbor("query_vector", 1000)).
				AddCondition(vespa.UserQuery("description")).
				AddCondition(vespa.Field("brand").Contains("premium")),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithQuery("summer casual wear").
		WithRanking("hybrid_profile").
		WithHits(40).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func multiVectorSearch() {
	fmt.Println("\n4. Multi-Vector Search:")

	textVector := []float32{0.1, 0.2, 0.3}
	imageVector := []float32{0.4, 0.5, 0.6}
	colorVector := []float32{0.7, 0.8, 0.9}

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "image_url", "color").
		From("fashion_products").
		Where(vespa.Field("available").Eq(true)).
		Rank(
			vespa.NewRank().
				AddCondition(vespa.Field("text_embedding").NearestNeighbor("text_vector", 800,
					vespa.WithLabel("text_similarity"))).
				AddCondition(vespa.Field("image_embedding").NearestNeighbor("image_vector", 500,
					vespa.WithLabel("visual_similarity"))).
				AddCondition(vespa.Field("color_embedding").NearestNeighbor("color_vector", 200,
					vespa.WithLabel("color_similarity"))),
		).
		WithInput("input.query(text_vector)", textVector).
		WithInput("input.query(image_vector)", imageVector).
		WithInput("input.query(color_vector)", colorVector).
		WithRanking("multi_modal_ranking").
		WithHits(25).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func approximateVsExactSearch() {
	fmt.Println("\n5. Approximate vs Exact Search:")

	queryVector := []float32{0.15, 0.25, 0.35, 0.45, 0.55}

	// Approximate search (faster, slightly less accurate)
	approximateQuery, err := vespa.NewQueryBuilder().
		Select("id", "title").
		From("research_papers").
		Where(
			vespa.Field("paper_embedding").NearestNeighbor("query_vector", 1000,
				vespa.WithApproximate(true),
				vespa.WithLabel("approximate_search"),
			),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithHits(50).
		Build()

	if err != nil {
		log.Fatalf("Error building approximate query: %v", err)
	}

	fmt.Println("Approximate Search (Fast):")
	printQuery(approximateQuery)

	// Exact search (slower, maximum accuracy)
	exactQuery, err := vespa.NewQueryBuilder().
		Select("id", "title").
		From("research_papers").
		Where(
			vespa.Field("paper_embedding").NearestNeighbor("query_vector", 500,
				vespa.WithApproximate(false),
				vespa.WithThreshold(0.95),
				vespa.WithLabel("exact_search"),
			),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithHits(20).
		Build()

	if err != nil {
		log.Fatalf("Error building exact query: %v", err)
	}

	fmt.Println("\nExact Search (High Precision):")
	printQuery(exactQuery)
}

func printQuery(query *vespa.VespaQuery) {
	jsonBytes, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		log.Printf("Error marshaling query: %v", err)
		return
	}
	fmt.Printf("%s\n", string(jsonBytes))
}
