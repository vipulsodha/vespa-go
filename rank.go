package vespa

import (
	"fmt"
	"strings"
)

// =============================================================================
// Utility Functions
// =============================================================================

// NewRank creates a new rank expression builder for constructing ranking logic.
// Rank expressions are used to score and order search results.
func NewRank() RankExpression {
	return &RankExpressionImpl{
		conditions: make([]WhereCondition, 0),
	}
}

// Custom creates a custom ranking condition with arbitrary expressions.
// This allows for complex ranking logic that cannot be expressed through the builder API.
func Custom(expression string) WhereCondition {
	return &CustomFeature{
		Expression: expression,
	}
}

// =============================================================================
// RankExpressionImpl
// =============================================================================

// RankExpressionImpl is the concrete implementation of RankExpression
type RankExpressionImpl struct {
	conditions []WhereCondition
}

// AddCondition adds a condition to the rank expression
func (r *RankExpressionImpl) AddCondition(condition WhereCondition) RankExpression {
	r.conditions = append(r.conditions, condition)
	return r
}

// ToYQL converts the rank expression to YQL format
func (r *RankExpressionImpl) ToYQL() string {
	if len(r.conditions) == 0 {
		return ""
	}

	var expressions []string

	// Add condition expressions
	for _, condition := range r.conditions {
		if yql := condition.ToYQL(); yql != "" {
			expressions = append(expressions, yql)
		}
	}

	if len(expressions) == 0 {
		return ""
	}

	return fmt.Sprintf("rank(%s)", strings.Join(expressions, ", "))
}

// =============================================================================
// CustomFeature
// =============================================================================

// CustomFeature allows for arbitrary ranking expressions
type CustomFeature struct {
	Expression string
}

func (cf *CustomFeature) ToYQL() string {
	return cf.Expression
}

func (cf *CustomFeature) And(condition WhereCondition) WhereCondition {
	return And(cf, condition)
}

func (cf *CustomFeature) Or(condition WhereCondition) WhereCondition {
	return Or(cf, condition)
}
