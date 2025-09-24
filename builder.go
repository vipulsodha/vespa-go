package vespa

import (
	"fmt"
	"strings"
)

// QueryBuilderImpl is the concrete implementation of QueryBuilder
type QueryBuilderImpl struct {
	selectFields    []string
	sources         []string
	whereConditions []WhereCondition
	rankExpression  RankExpression
	ranking         string
	hits            int
	offset          int
	defaultIndex    string
	inputParams     map[string]interface{}
	query           string
}

// NewQueryBuilder creates a new query builder instance.
// This is the main entry point for constructing Vespa YQL queries.
func NewQueryBuilder() QueryBuilder {
	return &QueryBuilderImpl{
		selectFields:    make([]string, 0),
		sources:         make([]string, 0),
		whereConditions: make([]WhereCondition, 0),
		inputParams:     make(map[string]interface{}),
	}
}

// Select specifies which fields to select
func (qb *QueryBuilderImpl) Select(fields ...string) QueryBuilder {
	qb.selectFields = append(qb.selectFields, fields...)
	return qb
}

// From specifies the data sources to query
func (qb *QueryBuilderImpl) From(sources ...string) QueryBuilder {
	qb.sources = append(qb.sources, sources...)
	return qb
}

// Where adds a where condition
func (qb *QueryBuilderImpl) Where(condition WhereCondition) QueryBuilder {
	qb.whereConditions = append(qb.whereConditions, condition)
	return qb
}

// Rank sets the ranking expression
func (qb *QueryBuilderImpl) Rank(rankExpression RankExpression) QueryBuilder {
	qb.rankExpression = rankExpression
	return qb
}

// WithRanking sets the ranking profile
func (qb *QueryBuilderImpl) WithRanking(profile string) QueryBuilder {
	qb.ranking = profile
	return qb
}

// WithHits sets the number of hits to return
func (qb *QueryBuilderImpl) WithHits(hits int) QueryBuilder {
	qb.hits = hits
	return qb
}

// WithOffset sets the number of initial results to skip (for pagination)
func (qb *QueryBuilderImpl) WithOffset(offset int) QueryBuilder {
	qb.offset = offset
	return qb
}

// WithDefaultIndex sets the default index for text queries
func (qb *QueryBuilderImpl) WithDefaultIndex(index string) QueryBuilder {
	qb.defaultIndex = index
	return qb
}

// WithInput adds an input parameter (e.g., for query vectors)
func (qb *QueryBuilderImpl) WithInput(key string, value interface{}) QueryBuilder {
	qb.inputParams[key] = value
	return qb
}

// WithQuery sets the text query
func (qb *QueryBuilderImpl) WithQuery(query string) QueryBuilder {
	qb.query = query
	return qb
}

// BuildYQL builds just the YQL string
func (qb *QueryBuilderImpl) BuildYQL() (string, error) {
	if err := qb.validate(); err != nil {
		return "", err
	}

	var yqlParts []string

	// SELECT clause
	selectClause := qb.buildSelectClause()
	yqlParts = append(yqlParts, selectClause)

	// FROM clause
	fromClause := qb.buildFromClause()
	yqlParts = append(yqlParts, fromClause)

	// WHERE clause
	whereClause := qb.buildWhereClause()
	if whereClause != "" {
		yqlParts = append(yqlParts, "where", whereClause)
	} else {
		// Vespa requires a WHERE clause, add default "true" condition
		yqlParts = append(yqlParts, "where", "true")
	}

	return strings.Join(yqlParts, " "), nil
}

// Build creates the complete VespaQuery
func (qb *QueryBuilderImpl) Build() (*VespaQuery, error) {
	yql, err := qb.BuildYQL()
	if err != nil {
		return nil, err
	}

	query := &VespaQuery{
		YQL: yql,
	}

	// Set optional fields
	if qb.ranking != "" {
		query.Ranking = qb.ranking
	}

	if qb.hits > 0 {
		query.Hits = qb.hits
	}

	if qb.offset > 0 {
		query.Offset = qb.offset
	}

	if qb.defaultIndex != "" {
		query.DefaultIndex = qb.defaultIndex
	}

	if len(qb.inputParams) > 0 {
		query.Input = make(map[string]interface{})
		for k, v := range qb.inputParams {
			query.Input[k] = v
		}
	}

	if qb.query != "" {
		query.Query = qb.query
	}

	return query, nil
}

// Helper methods for building query parts

func (qb *QueryBuilderImpl) buildSelectClause() string {
	if len(qb.selectFields) == 0 {
		return "select *"
	}
	return fmt.Sprintf("select %s", strings.Join(qb.selectFields, ", "))
}

func (qb *QueryBuilderImpl) buildFromClause() string {
	if len(qb.sources) == 0 {
		return "from sources *"
	}
	return fmt.Sprintf("from sources %s", strings.Join(qb.sources, ", "))
}

func (qb *QueryBuilderImpl) buildWhereClause() string {
	var conditions []string

	// Add regular where conditions
	for _, condition := range qb.whereConditions {
		if yql := condition.ToYQL(); yql != "" {
			conditions = append(conditions, yql)
		}
	}

	// Add rank expression if present
	if qb.rankExpression != nil {
		if rankYQL := qb.rankExpression.ToYQL(); rankYQL != "" {
			conditions = append(conditions, rankYQL)
		}
	}

	if len(conditions) == 0 {
		return ""
	}

	// Join conditions with AND
	return strings.Join(conditions, " and ")
}

// validate checks if the query builder state is valid
func (qb *QueryBuilderImpl) validate() error {
	// At minimum, we need a FROM clause or sources
	if len(qb.sources) == 0 {
		return &ValidationError{
			Field:   "sources",
			Message: "at least one source must be specified",
		}
	}

	// Validate that if we have input parameters, they follow the expected format
	for key := range qb.inputParams {
		if !strings.HasPrefix(key, "input.query(") {
			return &ValidationError{
				Field:   "input",
				Message: fmt.Sprintf("input parameter key '%s' should start with 'input.query('", key),
			}
		}
	}

	return nil
}
