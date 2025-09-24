package vespa

import (
	"fmt"
	"reflect"
	"strings"
)

// =============================================================================
// Utility Functions
// =============================================================================

// Field creates a new FieldBuilder for the given field name.
// This is the primary entry point for building field-based conditions.
//
// Example:
//   Field("price").Gte(100)
//   Field("category").In("electronics", "gadgets")
func Field(name string) FieldBuilder {
	return FieldBuilder{field: name}
}

// Eq creates an equality condition (= operator for numeric values, contains for strings).
func (f FieldBuilder) Eq(value interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: EQ, Value: value}
}

// NotEq creates a not-equal condition (!= operator for numeric values, not contains for strings).
func (f FieldBuilder) NotEq(value interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: NEQ, Value: value}
}

// Gt creates a greater-than condition (> operator).
func (f FieldBuilder) Gt(value interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: GT, Value: value}
}

// Gte creates a greater-than-or-equal condition (>= operator).
func (f FieldBuilder) Gte(value interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: GTE, Value: value}
}

// Lt creates a less-than condition (< operator).
func (f FieldBuilder) Lt(value interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: LT, Value: value}
}

// Lte creates a less-than-or-equal condition (<= operator).
func (f FieldBuilder) Lte(value interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: LTE, Value: value}
}

// In creates an IN condition that checks if the field value is in the provided list.
func (f FieldBuilder) In(values ...interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: IN, Value: values}
}

// NotIn creates a NOT IN condition that checks if the field value is not in the provided list.
func (f FieldBuilder) NotIn(values ...interface{}) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: NOT_IN, Value: values}
}

// Contains creates a CONTAINS condition for text matching with optional matching type.
// Supports exact, phrase, and fuzzy matching through functional options.
func (f FieldBuilder) Contains(value interface{}, opts ...ContainsOption) WhereCondition {
	// Apply options to config (default to ExactMatch)
	config := &ContainsConfig{Type: ExactMatch}
	for _, opt := range opts {
		opt(config)
	}

	return &FieldCondition{
		Field:        f.field,
		Operator:     CONTAINS,
		Value:        value,
		ContainsType: config.Type,
	}
}

// NotContains creates a NOT CONTAINS condition for text matching.
func (f FieldBuilder) NotContains(value string) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: NOT_CONTAINS, Value: value}
}

// Matches creates a MATCHES condition for regular expression matching.
func (f FieldBuilder) Matches(pattern string) WhereCondition {
	return &FieldCondition{Field: f.field, Operator: MATCHES, Value: pattern}
}

// Between creates a range condition (field >= min AND field <= max).
// This is a convenience method that generates two conditions combined with AND.
func (f FieldBuilder) Between(min, max interface{}) WhereCondition {
	return &RangeCondition{Field: f.field, Min: min, Max: max}
}

// NearestNeighbor creates a nearest neighbor condition for vector search.
// Can be used in both WHERE clauses and rank expressions with functional options.
func (f FieldBuilder) NearestNeighbor(queryVector string, targetHits int, opts ...NearestNeighborOption) WhereCondition {
	// Apply options to config
	config := &NearestNeighborConfig{}
	for _, opt := range opts {
		opt(config)
	}

	return &NearestNeighbor{
		Field:             f.field,
		QueryVector:       queryVector,
		TargetHits:        targetHits,
		Label:             config.Label,
		DistanceThreshold: config.DistanceThreshold,
		Approximate:       config.Approximate,
	}
}

// And combines multiple conditions with the AND boolean operator.
// Returns a single condition if only one is provided, nil if none are provided.
func And(conditions ...WhereCondition) WhereCondition {
	if len(conditions) == 0 {
		return nil
	}
	if len(conditions) == 1 {
		return conditions[0]
	}

	// Build left-associative tree: ((A AND B) AND C) AND D
	result := conditions[0]
	for i := 1; i < len(conditions); i++ {
		result = &BooleanCondition{
			Left:     result,
			Right:    conditions[i],
			Operator: "AND",
		}
	}
	return result
}

// Or combines multiple conditions with the OR boolean operator.
// Returns a single condition if only one is provided, nil if none are provided.
func Or(conditions ...WhereCondition) WhereCondition {
	if len(conditions) == 0 {
		return nil
	}
	if len(conditions) == 1 {
		return conditions[0]
	}

	// Build left-associative tree: ((A OR B) OR C) OR D
	result := conditions[0]
	for i := 1; i < len(conditions); i++ {
		result = &BooleanCondition{
			Left:     result,
			Right:    conditions[i],
			Operator: "OR",
		}
	}
	return result
}

// UserQuery creates a user query condition that can be used in both WHERE clauses and rank expressions.
// The optional defaultIndex parameter specifies which field to search in.
// The actual query text is provided via WithQuery() method on the builder.
func UserQuery(defaultIndex ...string) WhereCondition {
	var index string
	if len(defaultIndex) > 0 {
		index = defaultIndex[0]
	}
	return &UserQueryFeature{
		DefaultIndex: index,
	}
}

// Not creates a negated condition that wraps the given condition with !().
// This allows for complex boolean logic including XOR patterns.
func Not(condition WhereCondition) WhereCondition {
	return &NotCondition{Condition: condition}
}

// =============================================================================
// UserQueryFeature
// =============================================================================

// UserQueryFeature represents a userQuery condition for text search
type UserQueryFeature struct {
	DefaultIndex string
}

func (uq *UserQueryFeature) ToYQL() string {
	if uq.DefaultIndex != "" {
		return fmt.Sprintf("{defaultIndex:\"%s\"}userQuery()", uq.DefaultIndex)
	}
	return "userQuery()"
}

func (uq *UserQueryFeature) And(condition WhereCondition) WhereCondition {
	return And(uq, condition)
}

func (uq *UserQueryFeature) Or(condition WhereCondition) WhereCondition {
	return Or(uq, condition)
}

// =============================================================================
// FieldCondition
// =============================================================================

// FieldCondition represents a condition on a specific field
type FieldCondition struct {
	Field        string
	Operator     Operator
	Value        interface{}
	ContainsType ContainsType // Used only for CONTAINS operations
}

func (fc *FieldCondition) ToYQL() string {
	switch fc.Operator {
	case EQ:
		// For string values, use 'contains' for exact matching in Vespa
		if _, isString := fc.Value.(string); isString {
			return fmt.Sprintf("(%s contains %s)", fc.Field, formatValue(fc.Value))
		}
		return fmt.Sprintf("(%s = %s)", fc.Field, formatValue(fc.Value))
	case NEQ:
		// For string values, use 'not contains' for exact matching in Vespa
		if _, isString := fc.Value.(string); isString {
			return fmt.Sprintf("!(%s contains %s)", fc.Field, formatValue(fc.Value))
		}
		return fmt.Sprintf("(%s != %s)", fc.Field, formatValue(fc.Value))
	case GT:
		return fmt.Sprintf("(%s > %s)", fc.Field, formatValue(fc.Value))
	case GTE:
		return fmt.Sprintf("(%s >= %s)", fc.Field, formatValue(fc.Value))
	case LT:
		return fmt.Sprintf("(%s < %s)", fc.Field, formatValue(fc.Value))
	case LTE:
		return fmt.Sprintf("(%s <= %s)", fc.Field, formatValue(fc.Value))
	case IN:
		return fmt.Sprintf("(%s in %s)", fc.Field, formatInValues(fc.Value))
	case NOT_IN:
		return fmt.Sprintf("(%s not in %s)", fc.Field, formatInValues(fc.Value))
	case CONTAINS:
		switch fc.ContainsType {
		case ExactMatch:
			return fmt.Sprintf("(%s contains %s)", fc.Field, formatValue(fc.Value))
		case PhraseMatch:
			// Handle phrase matching for arrays of keywords
			if keywords, ok := fc.Value.([]string); ok {
				var quotedKeywords []string
				for _, kw := range keywords {
					quotedKeywords = append(quotedKeywords, fmt.Sprintf("'%s'", escapeString(kw)))
				}
				return fmt.Sprintf("(%s contains phrase(%s))", fc.Field, strings.Join(quotedKeywords, ", "))
			}
			// Handle phrase matching for single string values
			if str, ok := fc.Value.(string); ok {
				return fmt.Sprintf("(%s contains phrase(%s))", fc.Field, formatValue(str))
			}
			// Fallback to regular contains for other types
			return fmt.Sprintf("(%s contains %s)", fc.Field, formatValue(fc.Value))
		case FuzzyMatch:
			// For fuzzy matching, we can use a custom implementation
			return fmt.Sprintf("(%s contains fuzzy(%s))", fc.Field, formatValue(fc.Value))
		default:
			return fmt.Sprintf("(%s contains %s)", fc.Field, formatValue(fc.Value))
		}
	case NOT_CONTAINS:
		return fmt.Sprintf("(%s not contains %s)", fc.Field, formatValue(fc.Value))
	case MATCHES:
		return fmt.Sprintf("(%s matches %s)", fc.Field, formatValue(fc.Value))
	default:
		return ""
	}
}

func (fc *FieldCondition) And(condition WhereCondition) WhereCondition {
	return And(fc, condition)
}

func (fc *FieldCondition) Or(condition WhereCondition) WhereCondition {
	return Or(fc, condition)
}

// =============================================================================
// BooleanCondition
// =============================================================================

// BooleanCondition represents AND/OR combinations of conditions
type BooleanCondition struct {
	Left     WhereCondition
	Right    WhereCondition
	Operator string // "AND" or "OR"
}

func (bc *BooleanCondition) ToYQL() string {
	return fmt.Sprintf("(%s %s %s)", bc.Left.ToYQL(), bc.Operator, bc.Right.ToYQL())
}

func (bc *BooleanCondition) And(condition WhereCondition) WhereCondition {
	return And(bc, condition)
}

func (bc *BooleanCondition) Or(condition WhereCondition) WhereCondition {
	return Or(bc, condition)
}

// =============================================================================
// RangeCondition
// =============================================================================

// RangeCondition represents a BETWEEN condition (convenience for min <= field <= max)
type RangeCondition struct {
	Field string
	Min   interface{}
	Max   interface{}
}

func (rc *RangeCondition) ToYQL() string {
	minCondition := fmt.Sprintf("(%s >= %s)", rc.Field, formatValue(rc.Min))
	maxCondition := fmt.Sprintf("(%s <= %s)", rc.Field, formatValue(rc.Max))
	return fmt.Sprintf("(%s and %s)", minCondition, maxCondition)
}

func (rc *RangeCondition) And(condition WhereCondition) WhereCondition {
	return And(rc, condition)
}

func (rc *RangeCondition) Or(condition WhereCondition) WhereCondition {
	return Or(rc, condition)
}

// =============================================================================
// NearestNeighbor
// =============================================================================

// NearestNeighbor represents a nearestNeighbor operation that can be used in both WHERE clauses and RANK expressions
type NearestNeighbor struct {
	Field             string
	QueryVector       string
	TargetHits        int
	Label             string
	DistanceThreshold *float64
	Approximate       *bool
}

func (nn *NearestNeighbor) ToYQL() string {
	var params []string

	// Always include targetHits
	params = append(params, fmt.Sprintf("targetHits:%d", nn.TargetHits))

	// Add label if specified
	if nn.Label != "" {
		params = append(params, fmt.Sprintf("label:'%s'", nn.Label))
	}

	// Add distance threshold if specified
	if nn.DistanceThreshold != nil {
		params = append(params, fmt.Sprintf("distanceThreshold:%f", *nn.DistanceThreshold))
	}

	// Add approximate if specified
	if nn.Approximate != nil {
		params = append(params, fmt.Sprintf("approximate:%t", *nn.Approximate))
	}

	paramString := strings.Join(params, ",")
	return fmt.Sprintf("({%s}nearestNeighbor(%s, %s))", paramString, nn.Field, nn.QueryVector)
}

func (nn *NearestNeighbor) And(condition WhereCondition) WhereCondition {
	return And(nn, condition)
}

func (nn *NearestNeighbor) Or(condition WhereCondition) WhereCondition {
	return Or(nn, condition)
}

// =============================================================================
// Helper Functions
// =============================================================================

// Value formatting helpers
func formatValue(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", escapeString(v))
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		// Handle other types by converting to string
		return fmt.Sprintf("'%v'", v)
	}
}

func formatInValues(value interface{}) string {
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Slice {
		return fmt.Sprintf("(%s)", formatValue(value))
	}

	var formatted []string
	for i := 0; i < rv.Len(); i++ {
		val := rv.Index(i).Interface()
		formatted = append(formatted, formatValue(val))
	}

	return fmt.Sprintf("(%s)", strings.Join(formatted, ", "))
}

func escapeString(s string) string {
	// Escape single quotes in strings for YQL
	return strings.ReplaceAll(s, "'", "\\'")
}

// =============================================================================
// NotCondition
// =============================================================================

// NotCondition represents a negated condition (!condition)
func (nc *NotCondition) ToYQL() string {
	return fmt.Sprintf("!(%s)", nc.Condition.ToYQL())
}

func (nc *NotCondition) And(condition WhereCondition) WhereCondition {
	return And(nc, condition)
}

func (nc *NotCondition) Or(condition WhereCondition) WhereCondition {
	return Or(nc, condition)
}
