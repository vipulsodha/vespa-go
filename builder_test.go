package vespa

import (
	"strings"
	"testing"
)

func TestFieldConditions(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		operator Operator
		value    interface{}
		expected string
	}{
		{"Equal string", "brand", EQ, "nike", "(brand contains 'nike')"},
		{"Not equal string", "brand", NEQ, "adidas", "!(brand contains 'adidas')"},
		{"Greater than int", "price", GT, 100, "(price > 100)"},
		{"Greater than equal float", "price", GTE, 99.99, "(price >= 99.99)"},
		{"Less than int", "stock", LT, 50, "(stock < 50)"},
		{"Less than equal float", "rating", LTE, 4.5, "(rating <= 4.5)"},
		{"Contains text", "description", CONTAINS, "wireless", "(description contains 'wireless')"},
		{"Not contains text", "title", NOT_CONTAINS, "refurbished", "(title not contains 'refurbished')"},
		{"Matches pattern", "sku", MATCHES, "^PRD-.*", "(sku matches '^PRD-.*')"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := &FieldCondition{
				Field:    tt.field,
				Operator: tt.operator,
				Value:    tt.value,
			}

			result := condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFieldBuilder(t *testing.T) {
	tests := []struct {
		name     string
		builder  func() WhereCondition
		expected string
	}{
		{"Price equal", func() WhereCondition { return Field("price").Eq(100.0) }, "(price = 100)"},
		{"Brand not equal", func() WhereCondition { return Field("brand").NotEq("excluded") }, "!(brand contains 'excluded')"},
		{"Rating greater than", func() WhereCondition { return Field("rating").Gt(4.0) }, "(rating > 4)"},
		{"Stock greater than equal", func() WhereCondition { return Field("stock").Gte(10) }, "(stock >= 10)"},
		{"Price less than", func() WhereCondition { return Field("price").Lt(200.0) }, "(price < 200)"},
		{"Rating less than equal", func() WhereCondition { return Field("rating").Lte(3.5) }, "(rating <= 3.5)"},
		{"Category in list", func() WhereCondition { return Field("category").In("electronics", "gadgets") }, "(category in ('electronics', 'gadgets'))"},
		{"Brand not in list", func() WhereCondition { return Field("brand").NotIn("excluded1", "excluded2") }, "(brand not in ('excluded1', 'excluded2'))"},
		{"Description contains", func() WhereCondition { return Field("description").Contains("wireless") }, "(description contains 'wireless')"},
		{"Title not contains", func() WhereCondition { return Field("title").NotContains("refurbished") }, "(title not contains 'refurbished')"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := tt.builder()
			result := condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRangeCondition(t *testing.T) {
	condition := Field("price").Between(10.0, 100.0)
	expected := "((price >= 10) and (price <= 100))"

	result := condition.ToYQL()
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestBooleanConditions(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"AND condition",
			And(Field("price").Gte(10.0), Field("price").Lte(100.0)),
			"((price >= 10) AND (price <= 100))",
		},
		{
			"OR condition",
			Or(Field("brand").Eq("nike"), Field("brand").Eq("adidas")),
			"((brand contains 'nike') OR (brand contains 'adidas'))",
		},
		{
			"Complex nested condition",
			And(
				Field("price").Between(10.0, 100.0),
				Or(Field("brand").Eq("nike"), Field("category").Eq("electronics")),
			),
			"(((price >= 10) and (price <= 100)) AND ((brand contains 'nike') OR (category contains 'electronics')))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNearestNeighbor(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"Basic nearest neighbor",
			Field("embedding").NearestNeighbor("query_vector", 1000),
			"({targetHits:1000}nearestNeighbor(embedding, query_vector))",
		},
		{
			"With label",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithLabel("main_query")),
			"({targetHits:1000,label:'main_query'}nearestNeighbor(embedding, query_vector))",
		},
		{
			"With distance threshold",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithThreshold(0.8)),
			"({targetHits:1000,distanceThreshold:0.800000}nearestNeighbor(embedding, query_vector))",
		},
		{
			"With approximate true",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithApproximate(true)),
			"({targetHits:1000,approximate:true}nearestNeighbor(embedding, query_vector))",
		},
		{
			"With approximate false",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithApproximate(false)),
			"({targetHits:1000,approximate:false}nearestNeighbor(embedding, query_vector))",
		},
		{
			"With all parameters",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithLabel("main_query"), WithThreshold(0.8), WithApproximate(false)),
			"({targetHits:1000,label:'main_query',distanceThreshold:0.800000,approximate:false}nearestNeighbor(embedding, query_vector))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestContainsCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"Exact match",
			Field("color").Contains("red", WithExactMatching()),
			"(color contains 'red')",
		},
		{
			"Phrase match with keywords",
			Field("brand").Contains([]string{"nike", "air"}, WithPhraseMatching()),
			"(brand contains phrase('nike', 'air'))",
		},
		{
			"Fuzzy match",
			Field("description").Contains("wireless", WithFuzzyMatching()),
			"(description contains fuzzy('wireless'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRankExpression(t *testing.T) {
	rank := NewRank().
		AddCondition(Field("embedding").NearestNeighbor("query_vector", 1000)).
		AddCondition(Field("brand").Contains("nike")).
		AddCondition(Field("description").Contains([]string{"wireless", "headphones"}, WithPhraseMatching()))

	result := rank.ToYQL()
	expected := "rank(({targetHits:1000}nearestNeighbor(embedding, query_vector)), (brand contains 'nike'), (description contains phrase('wireless', 'headphones')))"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestNearestNeighborInWhereClause(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"Basic nearestNeighbor in WHERE",
			Field("embedding").NearestNeighbor("query_vector", 1000),
			"({targetHits:1000}nearestNeighbor(embedding, query_vector))",
		},
		{
			"nearestNeighbor with label in WHERE",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithLabel("main_search")),
			"({targetHits:1000,label:'main_search'}nearestNeighbor(embedding, query_vector))",
		},
		{
			"nearestNeighbor with threshold in WHERE",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithThreshold(0.8)),
			"({targetHits:1000,distanceThreshold:0.800000}nearestNeighbor(embedding, query_vector))",
		},
		{
			"nearestNeighbor with approximate true in WHERE",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithApproximate(true)),
			"({targetHits:1000,approximate:true}nearestNeighbor(embedding, query_vector))",
		},
		{
			"nearestNeighbor with approximate false in WHERE",
			Field("embedding").NearestNeighbor("query_vector", 1000, WithApproximate(false)),
			"({targetHits:1000,approximate:false}nearestNeighbor(embedding, query_vector))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNearestNeighborCombinedWithOtherConditions(t *testing.T) {
	condition := And(
		Field("embedding").NearestNeighbor("query_vector", 1000),
		Field("brand").Contains("nike"),
	)

	result := condition.ToYQL()
	expected := "(({targetHits:1000}nearestNeighbor(embedding, query_vector)) AND (brand contains 'nike'))"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRankWithMixedConditionsAndFeatures(t *testing.T) {
	rank := NewRank().
		AddCondition(Field("embedding").NearestNeighbor("query_vector", 1000)).
		AddCondition(Field("brand").Contains("nike")).
		AddCondition(Field("price").Gte(10.0))

	result := rank.ToYQL()
	expected := "rank(({targetHits:1000}nearestNeighbor(embedding, query_vector)), (brand contains 'nike'), (price >= 10))"

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestQueryBuilder_BasicQuery(t *testing.T) {
	query, err := NewQueryBuilder().
		Select("id", "title", "price").
		From("products").
		Where(Field("price").Between(10.0, 100.0)).
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedYQL := "select id, title, price from sources products where ((price >= 10) and (price <= 100))"
	if query.YQL != expectedYQL {
		t.Errorf("Expected YQL %q, got %q", expectedYQL, query.YQL)
	}
}

func TestQueryBuilder_ComplexQuery(t *testing.T) {
	queryVector := []float32{0.1, 0.2, 0.3}
	themeVector := []float32{0.4, 0.5, 0.6}

	query, err := NewQueryBuilder().
		Select("color", "brand", "product_type", "price", "product_id").
		From("products").
		Where(
			And(
				Field("price").Between(10.0, 100.0),
				Field("category").In("electronics", "gadgets"),
			),
		).
				Rank(
			NewRank().
			AddCondition(Field("embedding_field").NearestNeighbor("query_vector", 1000, WithLabel("query_vector"))).
			AddCondition(Field("embedding_field").NearestNeighbor("sort_vector", 1000, WithLabel("sort_vector"))).
			AddCondition(Field("brand").Contains([]string{"nike", "air"}, WithPhraseMatching())).
			AddCondition(Field("color").Contains("red")),
		).
		WithRanking("hybrid_profile").
		WithHits(50).
		WithDefaultIndex("refined_text").
		WithInput("input.query(query_vector)", queryVector).
		WithInput("input.query(sort_vector)", themeVector).
		WithQuery("wireless headphones").
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify YQL structure - exact match
	expectedYQLParts := []string{
		"select color, brand, product_type, price, product_id",
		"from sources products",
		"where (((price >= 10) and (price <= 100)) AND (category in ('electronics', 'gadgets')))",
		"and rank(({targetHits:1000,label:'query_vector'}nearestNeighbor(embedding_field, query_vector)), ({targetHits:1000,label:'sort_vector'}nearestNeighbor(embedding_field, sort_vector)), (brand contains phrase('nike', 'air')), (color contains 'red'))",
	}
	expectedYQL := strings.Join(expectedYQLParts, " ")

	if query.YQL != expectedYQL {
		t.Errorf("Expected YQL %q, got %q", expectedYQL, query.YQL)
	}

	// Verify other parameters
	if query.Ranking != "hybrid_profile" {
		t.Errorf("Expected ranking %q, got %q", "hybrid_profile", query.Ranking)
	}
	if query.Hits != 50 {
		t.Errorf("Expected hits %d, got %d", 50, query.Hits)
	}
	if query.DefaultIndex != "refined_text" {
		t.Errorf("Expected defaultIndex %q, got %q", "refined_text", query.DefaultIndex)
	}
	if query.Query != "wireless headphones" {
		t.Errorf("Expected query %q, got %q", "wireless headphones", query.Query)
	}

	// Verify input parameters
	if len(query.Input) != 2 {
		t.Errorf("Expected 2 input parameters, got %d", len(query.Input))
	}
}

func TestQueryBuilder_WithNearestNeighborInWhere(t *testing.T) {
	queryVector := []float32{0.1, 0.2, 0.3}

	query, err := NewQueryBuilder().
		Select("id", "title", "price").
		From("products").
		Where(
			And(
				Field("embedding").NearestNeighbor("query_vector", 1000),
				Field("brand").Contains("nike"),
			),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithRanking("hybrid_profile").
		WithHits(50).
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify YQL structure
	expectedYQLParts := []string{
		"select id, title, price",
		"from sources products",
		"where (({targetHits:1000}nearestNeighbor(embedding, query_vector)) AND (brand contains 'nike'))",
	}
	expectedYQL := strings.Join(expectedYQLParts, " ")

	if query.YQL != expectedYQL {
		t.Errorf("Expected YQL %q, got %q", expectedYQL, query.YQL)
	}

	// Verify other parameters
	if query.Ranking != "hybrid_profile" {
		t.Errorf("Expected ranking %q, got %q", "hybrid_profile", query.Ranking)
	}
	if query.Hits != 50 {
		t.Errorf("Expected hits %d, got %d", 50, query.Hits)
	}
}

func TestQueryBuilder_MixedWhereAndRank(t *testing.T) {
	query, err := NewQueryBuilder().
		Select("id", "title", "price").
		From("products").
		Where(
			And(
				Field("embedding").NearestNeighbor("query_vector", 1000),
				Field("price").Gte(10.0),
			),
		).
		Rank(
			NewRank().
				AddCondition(Field("brand").Contains("nike")).
				AddCondition(Field("description").Contains("wireless")),
		).
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify YQL has both WHERE and RANK - exact match
	expectedYQLParts := []string{
		"select id, title, price",
		"from sources products",
		"where (({targetHits:1000}nearestNeighbor(embedding, query_vector)) AND (price >= 10))",
		"and rank((brand contains 'nike'), (description contains 'wireless'))",
	}
	expectedYQL := strings.Join(expectedYQLParts, " ")

	if query.YQL != expectedYQL {
		t.Errorf("Expected YQL %q, got %q", expectedYQL, query.YQL)
	}
}

func TestQueryBuilder_Validation(t *testing.T) {
	tests := []struct {
		name        string
		builder     func() QueryBuilder
		expectError bool
	}{
		{
			"Valid query",
			func() QueryBuilder {
				return NewQueryBuilder().From("products").Where(Field("price").Gt(0))
			},
			false,
		},
		{
			"Missing source",
			func() QueryBuilder {
				return NewQueryBuilder().Where(Field("price").Gt(0))
			},
			true,
		},
		{
			"Invalid input parameter key",
			func() QueryBuilder {
				return NewQueryBuilder().
					From("products").
					WithInput("invalid_key", []float32{1, 2, 3})
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder().Build()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}





func TestValueFormatting(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"String", "test", "'test'"},
		{"String with quotes", "test's", "'test\\'s'"},
		{"Integer", 42, "42"},
		{"Float", 3.14, "3.14"},
		{"Boolean true", true, "true"},
		{"Boolean false", false, "false"},
		{"Nil", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestStringEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No quotes", "test", "test"},
		{"Single quote", "test's", "test\\'s"},
		{"Multiple quotes", "it's a 'test'", "it\\'s a \\'test\\'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeString(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
	// TestSameElementCondition tests basic sameElement functionality
func TestSameElementCondition(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"Simple sameElement with two conditions",
			Field("sizes").ContainsSameElement(
				Field("family").Contains("clothing"),
				Field("size_value").Contains("M"),
			),
			"(sizes contains sameElement((family contains 'clothing'), (size_value contains 'M')))",
		},
		{
			"Complex sameElement with multiple condition types",
			Field("products").ContainsSameElement(
				Field("name").Contains("shirt"),
				Field("price").Gt(20.0),
				Field("brand").Eq("nike"),
			),
			"(products contains sameElement((name contains 'shirt'), (price > 20), (brand contains 'nike')))",
		},
		{
			"SameElement with numeric and text conditions",
			Field("persons").ContainsSameElement(
				Field("first_name").Contains("John"),
				Field("last_name").Contains("Doe"),
				Field("year_of_birth").Lt(1990),
			),
			"(persons contains sameElement((first_name contains 'John'), (last_name contains 'Doe'), (year_of_birth < 1990)))",
		},
		{
			"SameElement with IN condition",
			Field("attributes").ContainsSameElement(
				Field("key").Contains("color"),
				Field("value").In("red", "blue", "green"),
			),
			"(attributes contains sameElement((key contains 'color'), (value in ('red', 'blue', 'green'))))",
		},
		{
			"SameElement with range condition",
			Field("inventory").ContainsSameElement(
				Field("location").Contains("warehouse1"),
				Field("quantity").Between(10, 100),
			),
			"(inventory contains sameElement((location contains 'warehouse1'), ((quantity >= 10) and (quantity <= 100))))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSameElementConditionWithBooleanLogic tests sameElement combined with AND/OR
func TestSameElementConditionWithBooleanLogic(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"SameElement AND regular condition",
			And(
				Field("sizes").ContainsSameElement(
					Field("family").Contains("clothing"),
					Field("size_value").Contains("L"),
				),
				Field("brand").Contains("nike"),
			),
			"((sizes contains sameElement((family contains 'clothing'), (size_value contains 'L'))) AND (brand contains 'nike'))",
		},
		{
			"SameElement OR regular condition",
			Or(
				Field("attributes").ContainsSameElement(
					Field("key").Contains("color"),
					Field("value").Contains("red"),
				),
				Field("price").Lt(50.0),
			),
			"((attributes contains sameElement((key contains 'color'), (value contains 'red'))) OR (price < 50))",
		},
		{
			"Multiple sameElement conditions with AND",
			And(
				Field("sizes").ContainsSameElement(
					Field("family").Contains("clothing"),
					Field("size_value").Contains("M"),
				),
				Field("attributes").ContainsSameElement(
					Field("key").Contains("material"),
					Field("value").Contains("cotton"),
				),
			),
			"((sizes contains sameElement((family contains 'clothing'), (size_value contains 'M'))) AND (attributes contains sameElement((key contains 'material'), (value contains 'cotton'))))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSameElementConditionChaining tests And/Or method chaining
func TestSameElementConditionChaining(t *testing.T) {
	condition := Field("sizes").ContainsSameElement(
		Field("family").Contains("clothing"),
		Field("size_value").Contains("L"),
	).And(Field("brand").Contains("adidas"))

	expected := "((sizes contains sameElement((family contains 'clothing'), (size_value contains 'L'))) AND (brand contains 'adidas'))"
	result := condition.ToYQL()

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}



// TestSameElementConditionEdgeCases tests edge cases and error conditions
func TestSameElementConditionEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		condition *SameElementCondition
		expected  string
	}{
		{
			"Empty conditions list",
			&SameElementCondition{
				Field:      "sizes",
				Conditions: []WhereCondition{},
			},
			"",
		},
		{
			"Single condition",
			&SameElementCondition{
				Field: "sizes",
				Conditions: []WhereCondition{
					Field("family").Contains("clothing"),
				},
			},
			"(sizes contains sameElement((family contains 'clothing')))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSameElementConditionRealWorldExamples tests realistic use cases
func TestSameElementConditionRealWorldExamples(t *testing.T) {
	tests := []struct {
		name        string
		description string
		condition   WhereCondition
		expected    string
	}{
		{
			"Product sizing query",
			"Find products with medium clothing sizes",
			Field("sizes").ContainsSameElement(
				Field("family").Contains("clothing"),
				Field("size_value").In("M", "Medium"),
			),
			"(sizes contains sameElement((family contains 'clothing'), (size_value in ('M', 'Medium'))))",
		},
		{
			"Person search query",
			"Find documents with a person named John Smith born after 1980",
			Field("persons").ContainsSameElement(
				Field("first_name").Contains("John"),
				Field("last_name").Contains("Smith"),
				Field("year_of_birth").Gt(1980),
			),
			"(persons contains sameElement((first_name contains 'John'), (last_name contains 'Smith'), (year_of_birth > 1980)))",
		},
		{
			"Attribute filtering",
			"Find products with specific color attribute",
			Field("attributes").ContainsSameElement(
				Field("key").Eq("color"),
				Field("value").In("red", "blue", "green"),
			),
			"(attributes contains sameElement((key contains 'color'), (value in ('red', 'blue', 'green'))))",
		},
		{
			"Complex inventory query",
			"Find inventory items in specific location with sufficient quantity",
			And(
				Field("inventory_items").ContainsSameElement(
					Field("location_code").Contains("WH01"),
					Field("quantity").Gte(100),
					Field("status").Eq("available"),
				),
				Field("product_type").Contains("electronics"),
			),
			"((inventory_items contains sameElement((location_code contains 'WH01'), (quantity >= 100), (status contains 'available'))) AND (product_type contains 'electronics'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Test: %s\nDescription: %s\nExpected: %q\nGot: %q", tt.name, tt.description, tt.expected, result)
			}
		})
	}
}

// TestQueryBuilder_WithOffset tests offset functionality
func TestQueryBuilder_WithOffset(t *testing.T) {
	tests := []struct {
		name           string
		offset         int
		expectOffset   bool
		expectedOffset int
	}{
		{"Zero offset", 0, false, 0},
		{"Positive offset", 10, true, 10},
		{"Large offset", 1000, true, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := NewQueryBuilder().
				Select("id", "title").
				From("products").
				Where(Field("price").Gt(0)).
				WithOffset(tt.offset).
				Build()

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if tt.expectOffset {
				if query.Offset != tt.expectedOffset {
					t.Errorf("Expected offset %d, got %d", tt.expectedOffset, query.Offset)
				}
			} else {
				if query.Offset != 0 {
					t.Errorf("Expected offset to be 0 for zero input, got %d", query.Offset)
				}
			}
		})
	}
}

// TestQueryBuilder_WithHitsAndOffset tests hits and offset together
func TestQueryBuilder_WithHitsAndOffset(t *testing.T) {
	query, err := NewQueryBuilder().
		Select("id", "title", "price").
		From("products").
		Where(Field("category").Eq("electronics")).
		WithHits(20).
		WithOffset(50).
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedYQL := "select id, title, price from sources products where (category contains 'electronics')"
	if query.YQL != expectedYQL {
		t.Errorf("Expected YQL %q, got %q", expectedYQL, query.YQL)
	}

	if query.Hits != 20 {
		t.Errorf("Expected hits %d, got %d", 20, query.Hits)
	}

	if query.Offset != 50 {
		t.Errorf("Expected offset %d, got %d", 50, query.Offset)
	}
}

// TestQueryBuilder_PaginationScenario tests a realistic pagination scenario
func TestQueryBuilder_PaginationScenario(t *testing.T) {
	// Simulate page 3 with 25 items per page (offset = 50)
	page := 3
	itemsPerPage := 25
	offset := (page - 1) * itemsPerPage

	query, err := NewQueryBuilder().
		Select("id", "name", "price", "brand").
		From("products").
		Where(
			And(
				Field("category").In("electronics", "gadgets"),
				Field("price").Between(10.0, 500.0),
			),
		).
		WithHits(itemsPerPage).
		WithOffset(offset).
		WithRanking("popularity").
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if query.Hits != 25 {
		t.Errorf("Expected hits %d, got %d", 25, query.Hits)
	}

	if query.Offset != 50 {
		t.Errorf("Expected offset %d, got %d", 50, query.Offset)
	}

	if query.Ranking != "popularity" {
		t.Errorf("Expected ranking %q, got %q", "popularity", query.Ranking)
	}
}

// TestQueryBuilder_ComplexQueryWithOffset tests offset with vector search
func TestQueryBuilder_ComplexQueryWithOffset(t *testing.T) {
	queryVector := []float32{0.1, 0.2, 0.3}

	query, err := NewQueryBuilder().
		Select("id", "title", "embedding_score").
		From("products").
		Where(
			And(
				Field("embedding").NearestNeighbor("query_vector", 1000),
				Field("category").Contains("fashion"),
			),
		).
		Rank(
			NewRank().
				AddCondition(Field("popularity").Gte(0.7)).
				AddCondition(Field("brand").Contains("premium")),
		).
		WithInput("input.query(query_vector)", queryVector).
		WithHits(15).
		WithOffset(30).
		WithRanking("hybrid_search").
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify offset and hits are set correctly
	if query.Offset != 30 {
		t.Errorf("Expected offset %d, got %d", 30, query.Offset)
	}

	if query.Hits != 15 {
		t.Errorf("Expected hits %d, got %d", 15, query.Hits)
	}

	// Verify other parameters are preserved
	if query.Ranking != "hybrid_search" {
		t.Errorf("Expected ranking %q, got %q", "hybrid_search", query.Ranking)
	}

	if len(query.Input) != 1 {
		t.Errorf("Expected 1 input parameter, got %d", len(query.Input))
	}
}

// =============================================================================
// UserQuery Tests
// =============================================================================

// TestUserQueryFeature tests the UserQuery function with different parameter combinations
func TestUserQueryFeature(t *testing.T) {
	tests := []struct {
		name      string
		userQuery func() WhereCondition
		expected  string
	}{
		{
			"UserQuery without default index",
			func() WhereCondition { return UserQuery() },
			"userQuery()",
		},
		{
			"UserQuery with default index",
			func() WhereCondition { return UserQuery("title") },
			"{defaultIndex:\"title\"}userQuery()",
		},
		{
			"UserQuery with empty string default index",
			func() WhereCondition { return UserQuery("") },
			"userQuery()",
		},
		{
			"UserQuery with field name default index",
			func() WhereCondition { return UserQuery("description") },
			"{defaultIndex:\"description\"}userQuery()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := tt.userQuery()
			result := condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUserQueryInWhereClause tests UserQuery as a standalone WHERE condition
func TestUserQueryInWhereClause(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"UserQuery alone in WHERE",
			UserQuery(),
			"userQuery()",
		},
		{
			"UserQuery with index in WHERE",
			UserQuery("title"),
			"{defaultIndex:\"title\"}userQuery()",
		},
		{
			"UserQuery combined with AND",
			And(UserQuery("title"), Field("price").Gt(10)),
			"({defaultIndex:\"title\"}userQuery() AND (price > 10))",
		},
		{
			"UserQuery combined with OR",
			Or(UserQuery(), Field("category").Eq("electronics")),
			"(userQuery() OR (category contains 'electronics'))",
		},
		{
			"Multiple conditions with UserQuery",
			And(
				UserQuery("description"),
				Field("price").Between(10, 100),
				Field("brand").Contains("nike"),
			),
			"(({defaultIndex:\"description\"}userQuery() AND ((price >= 10) and (price <= 100))) AND (brand contains 'nike'))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUserQueryInRankExpression tests UserQuery in rank expressions
func TestUserQueryInRankExpression(t *testing.T) {
	tests := []struct {
		name     string
		rank     RankExpression
		expected string
	}{
		{
			"UserQuery alone in rank",
			NewRank().AddCondition(UserQuery()),
			"rank(userQuery())",
		},
		{
			"UserQuery with index in rank",
			NewRank().AddCondition(UserQuery("title")),
			"rank({defaultIndex:\"title\"}userQuery())",
		},
		{
			"UserQuery combined with other conditions in rank",
			NewRank().
				AddCondition(UserQuery("description")).
				AddCondition(Field("brand").Contains("nike")).
				AddCondition(Field("price").Gte(50)),
			"rank({defaultIndex:\"description\"}userQuery(), (brand contains 'nike'), (price >= 50))",
		},
		{
			"UserQuery with vector search in rank",
			NewRank().
				AddCondition(Field("embedding").NearestNeighbor("query_vector", 1000)).
				AddCondition(UserQuery("title")),
			"rank(({targetHits:1000}nearestNeighbor(embedding, query_vector)), {defaultIndex:\"title\"}userQuery())",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rank.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestUserQueryWithCompleteQueryBuilder tests UserQuery with full query building
func TestUserQueryWithCompleteQueryBuilder(t *testing.T) {
	tests := []struct {
		name          string
		builder       func() QueryBuilder
		expectedYQL   string
		expectedQuery string
	}{
		{
			"Basic UserQuery in WHERE with query text",
			func() QueryBuilder {
				return NewQueryBuilder().
					Select("id", "title", "price").
					From("products").
					Where(UserQuery("title")).
					WithQuery("wireless headphones")
			},
			"select id, title, price from sources products where {defaultIndex:\"title\"}userQuery()",
			"wireless headphones",
		},
		{
			"UserQuery with filters and query text",
			func() QueryBuilder {
				return NewQueryBuilder().
					Select("*").
					From("products").
					Where(
						And(
							UserQuery("description"),
							Field("price").Between(20, 100),
							Field("category").Eq("electronics"),
						),
					).
					WithQuery("bluetooth speaker").
					WithHits(25)
			},
			"select * from sources products where (({defaultIndex:\"description\"}userQuery() AND ((price >= 20) and (price <= 100))) AND (category contains 'electronics'))",
			"bluetooth speaker",
		},
		{
			"UserQuery in rank with WHERE conditions",
			func() QueryBuilder {
				return NewQueryBuilder().
					Select("id", "title", "score").
					From("products").
					Where(Field("category").In("electronics", "gadgets")).
					Rank(
						NewRank().
							AddCondition(UserQuery("title")).
							AddCondition(Field("brand").Contains("premium")),
					).
					WithQuery("gaming headset").
					WithRanking("text_ranking")
			},
			"select id, title, score from sources products where (category in ('electronics', 'gadgets')) and rank({defaultIndex:\"title\"}userQuery(), (brand contains 'premium'))",
			"gaming headset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, err := tt.builder().Build()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if query.YQL != tt.expectedYQL {
				t.Errorf("Expected YQL %q, got %q", tt.expectedYQL, query.YQL)
			}

			if query.Query != tt.expectedQuery {
				t.Errorf("Expected query %q, got %q", tt.expectedQuery, query.Query)
			}
		})
	}
}

// TestUserQueryMethodChaining tests method chaining with UserQuery
func TestUserQueryMethodChaining(t *testing.T) {
	condition := UserQuery("title").
		And(Field("price").Gt(50)).
		Or(Field("category").Eq("premium"))

	expected := "(({defaultIndex:\"title\"}userQuery() AND (price > 50)) OR (category contains 'premium'))"
	result := condition.ToYQL()

	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// =============================================================================
// Not Condition Tests
// =============================================================================

func TestNotConditionBasic(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"Not simple contains condition",
			Not(Field("brand").Contains("nike")),
			"!((brand contains 'nike'))",
		},
		{
			"Not equals condition",
			Not(Field("category").Eq("electronics")),
			"!((category contains 'electronics'))",
		},
		{
			"Not complex AND condition",
			Not(And(Field("brand").Contains("nike"), Field("item_types").Contains("shoes"))),
			"!(((brand contains 'nike') AND (item_types contains 'shoes')))",
		},
		{
			"Not OR condition",
			Not(Or(Field("price").Gt(100), Field("stock").Lt(5))),
			"!(((price > 100) OR (stock < 5)))",
		},
		{
			"Not range condition",
			Not(Field("price").Between(50.0, 150.0)),
			"!(((price >= 50) and (price <= 150)))",
		},
		{
			"Not IN condition",
			Not(Field("category").In("electronics", "clothing")),
			"!((category in ('electronics', 'clothing')))",
		},
		{
			"Double negation",
			Not(Not(Field("active").Eq(true))),
			"!(!((active = true)))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNotWithBooleanCombinations(t *testing.T) {
	tests := []struct {
		name      string
		condition WhereCondition
		expected  string
	}{
		{
			"And with Not condition",
			And(
				Field("category").Eq("electronics"),
				Not(Or(
					Field("brand").Contains("excluded1"),
					Field("brand").Contains("excluded2"),
				)),
			),
			"((category contains 'electronics') AND !(((brand contains 'excluded1') OR (brand contains 'excluded2'))))",
		},
		{
			"Or with Not condition",
			Or(
				Field("stock").Gt(0),
				Not(Field("discontinued").Eq(true)),
			),
			"((stock > 0) OR !((discontinued = true)))",
		},
		{
			"Complex nested with Not",
			And(
				Field("price").Between(20.0, 200.0),
				Not(And(
					Field("brand").Contains("excluded"),
					Field("category").Eq("restricted"),
				)),
			),
			"(((price >= 20) and (price <= 200)) AND !(((brand contains 'excluded') AND (category contains 'restricted'))))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.ToYQL()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNotConditionChaining(t *testing.T) {
	// Test that Not condition can be combined with And/Or methods
	notCondition := Not(Field("brand").Contains("nike"))

	// Test And method
	andCondition := notCondition.And(Field("price").Gt(50))
	expectedAnd := "(!((brand contains 'nike')) AND (price > 50))"
	resultAnd := andCondition.ToYQL()
	if resultAnd != expectedAnd {
		t.Errorf("And method - Expected %q, got %q", expectedAnd, resultAnd)
	}

	// Test Or method
	orCondition := notCondition.Or(Field("stock").Lt(10))
	expectedOr := "(!((brand contains 'nike')) OR (stock < 10))"
	resultOr := orCondition.ToYQL()
	if resultOr != expectedOr {
		t.Errorf("Or method - Expected %q, got %q", expectedOr, resultOr)
	}
}

func TestNotConditionXORUseCase(t *testing.T) {
	// Test the original XOR use case
	query, err := NewQueryBuilder().
		Select("brand", "item_types").
		From("listings_sg_v1").
		Where(
			And(
				// At least one condition is true
				Or(
					Field("brand").Contains("nike"),
					Field("item_types").Contains("shoes"),
				),
				// But not both are true - using Not()
				Not(
					And(
						Field("brand").Contains("nike"),
						Field("item_types").Contains("shoes"),
					),
				),
			),
		).
		Build()

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expectedYQL := "select brand, item_types from sources listings_sg_v1 where (((brand contains 'nike') OR (item_types contains 'shoes')) AND !(((brand contains 'nike') AND (item_types contains 'shoes'))))"
	if query.YQL != expectedYQL {
		t.Errorf("Expected YQL %q, got %q", expectedYQL, query.YQL)
	}
}