package vespa

import "fmt"

// Operator represents comparison operators for where conditions
type Operator string

const (
	// Equality operators
	EQ  Operator = "="
	NEQ Operator = "!="

	// Comparison operators
	GT  Operator = ">"
	GTE Operator = ">="
	LT  Operator = "<"
	LTE Operator = "<="

	// Collection operators
	IN     Operator = "in"
	NOT_IN Operator = "not in"

	// Text operators
	CONTAINS     Operator = "contains"
	NOT_CONTAINS Operator = "not contains"
	MATCHES      Operator = "matches"
)

// ContainsType represents different types of text matching
type ContainsType string

const (
	ExactMatch  ContainsType = "exact"
	PhraseMatch ContainsType = "phrase"
	FuzzyMatch  ContainsType = "fuzzy"
)

// QueryBuilder is the main interface for building YQL queries
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

// WhereCondition represents a condition in the WHERE clause
type WhereCondition interface {
	ToYQL() string
	And(condition WhereCondition) WhereCondition
	Or(condition WhereCondition) WhereCondition
}

// RankExpression represents a ranking expression
type RankExpression interface {
	ToYQL() string
	AddCondition(condition WhereCondition) RankExpression
}

// FieldBuilder provides fluent API for building field conditions
type FieldBuilder struct {
	field string
}

// VespaQuery represents the final query structure
type VespaQuery struct {
	YQL          string                 `json:"yql"`
	Ranking      string                 `json:"ranking,omitempty"`
	Hits         int                    `json:"hits,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	DefaultIndex string                 `json:"defaultIndex,omitempty"`
	Input        map[string]interface{} `json:"input,omitempty"`
	Query        string                 `json:"query,omitempty"`
}

// ValidationError represents validation errors in query building
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// Option types for functional options pattern

// NearestNeighborOption represents options for nearest neighbor operations
type NearestNeighborOption func(*NearestNeighborConfig)

// NearestNeighborConfig holds configuration for nearest neighbor operations
type NearestNeighborConfig struct {
	Label             string
	DistanceThreshold *float64
	Approximate       *bool
}

// WithLabel adds a label to the nearest neighbor operation
func WithLabel(label string) NearestNeighborOption {
	return func(config *NearestNeighborConfig) {
		if config != nil {
			config.Label = label
		}
	}
}

// WithThreshold adds a distance threshold to the nearest neighbor operation
func WithThreshold(threshold float64) NearestNeighborOption {
	return func(config *NearestNeighborConfig) {
		if config != nil {
			config.DistanceThreshold = &threshold
		}
	}
}

// WithApproximate sets whether to use approximate or exact nearest neighbor search
func WithApproximate(approximate bool) NearestNeighborOption {
	return func(config *NearestNeighborConfig) {
		if config != nil {
			config.Approximate = &approximate
		}
	}
}

// ContainsOption represents options for contains operations
type ContainsOption func(*ContainsConfig)

// ContainsConfig holds configuration for contains operations
type ContainsConfig struct {
	Type ContainsType
}

// WithExactMatching sets contains to use exact matching (default)
func WithExactMatching() ContainsOption {
	return func(config *ContainsConfig) {
		if config != nil {
			config.Type = ExactMatch
		}
	}
}

// WithPhraseMatching sets contains to use phrase matching
func WithPhraseMatching() ContainsOption {
	return func(config *ContainsConfig) {
		if config != nil {
			config.Type = PhraseMatch
		}
	}
}

// WithFuzzyMatching sets contains to use fuzzy matching
func WithFuzzyMatching() ContainsOption {
	return func(config *ContainsConfig) {
		if config != nil {
			config.Type = FuzzyMatch
		}
	}
}

// NotCondition represents a negated condition (!condition)
type NotCondition struct {
	Condition WhereCondition
}
