package main

import (
	"encoding/json"
	"fmt"
	"log"

	vespa "github.com/vipulsodha/vespa-go"
)

func main() {
	fmt.Println("Basic Search Examples")
	fmt.Println("====================")

	// Example 1: Simple product search with filters
	basicProductSearch()

	// Example 2: Range filtering
	priceRangeSearch()

	// Example 3: Boolean combinations
	booleanLogicSearch()

	// Example 4: Text matching
	textMatchingSearch()
}

func basicProductSearch() {
	fmt.Println("\n1. Basic Product Search:")

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "price", "brand").
		From("products").
		Where(
			vespa.And(
				vespa.Field("category").Eq("electronics"),
				vespa.Field("stock").Gt(0),
			),
		).
		WithHits(20).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func priceRangeSearch() {
	fmt.Println("\n2. Price Range Search:")

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "price").
		From("products").
		Where(vespa.Field("price").Between(50.0, 200.0)).
		WithHits(10).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func booleanLogicSearch() {
	fmt.Println("\n3. Boolean Logic Search:")

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "brand", "category").
		From("products").
		Where(
			vespa.And(
				vespa.Or(
					vespa.Field("brand").In("nike", "adidas"),
					vespa.Field("category").Eq("sportswear"),
				),
				vespa.Field("rating").Gte(4.0),
				vespa.Not(vespa.Field("discontinued").Eq(true)),
			),
		).
		WithHits(15).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func textMatchingSearch() {
	fmt.Println("\n4. Text Matching Search:")

	query, err := vespa.NewQueryBuilder().
		Select("id", "title", "description").
		From("products").
		Where(
			vespa.And(
				vespa.Field("title").Contains("wireless"),
				vespa.Field("description").Contains("bluetooth"),
			),
		).
		WithHits(25).
		Build()

	if err != nil {
		log.Fatalf("Error building query: %v", err)
	}

	printQuery(query)
}

func printQuery(query *vespa.VespaQuery) {
	jsonBytes, err := json.MarshalIndent(query, "", "  ")
	if err != nil {
		log.Printf("Error marshaling query: %v", err)
		return
	}
	fmt.Printf("Generated Query:\n%s\n", string(jsonBytes))
}
