package vespa

import (
	"fmt"
	"strings"
)

// SameElementCondition represents a sameElement condition for complex fields (arrays of structs/maps)
type SameElementCondition struct {
	Field      string
	Conditions []WhereCondition
}

// ContainsSameElement creates a sameElement condition for complex fields (arrays of structs/maps).
// This ensures that all conditions match within the same element of an array or map structure.
//
// IMPORTANT VESPA LIMITATIONS:
// - IN operators are NOT supported inside sameElement conditions (will cause 400 errors)
// - OR conditions are NOT supported inside sameElement conditions (will cause 400 errors)
// - Use top-level OR with multiple sameElement conditions as a workaround:
//
//   // ❌ This doesn't work:
//   sizes.ContainsSameElement(
//     Field("family").Eq("US"),
//     Field("size_value").In("10", "11")  // Fails
//   )
//
//   // ✅ Use this instead:
//   Or(
//     sizes.ContainsSameElement(Field("family").Eq("US"), Field("size_value").Eq("10")),
//     sizes.ContainsSameElement(Field("family").Eq("US"), Field("size_value").Eq("11"))
//   )
func (f FieldBuilder) ContainsSameElement(conditions ...WhereCondition) WhereCondition {
	return &SameElementCondition{
		Field:      f.field,
		Conditions: conditions,
	}
}

// ToYQL for SameElementCondition
func (se *SameElementCondition) ToYQL() string {
	if len(se.Conditions) == 0 {
		return ""
	}

	// Convert all conditions to their YQL representation for sameElement context
	var conditionStrings []string
	for _, condition := range se.Conditions {
		if conditionYQL := condition.ToYQL(); conditionYQL != "" {
			conditionStrings = append(conditionStrings, conditionYQL)
		}
	}

	if len(conditionStrings) == 0 {
		return ""
	}

	return fmt.Sprintf("(%s contains sameElement(%s))", se.Field, strings.Join(conditionStrings, ", "))
}

// And/Or methods for SameElementCondition
func (se *SameElementCondition) And(condition WhereCondition) WhereCondition {
	return And(se, condition)
}

func (se *SameElementCondition) Or(condition WhereCondition) WhereCondition {
	return Or(se, condition)
}
