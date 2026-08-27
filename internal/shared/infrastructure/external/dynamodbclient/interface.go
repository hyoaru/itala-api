package dynamodbclient

import "context"

type TransactPut struct {
	TableName string
	Item      map[string]any
	Condition string
}

type TransactUpdate struct {
	TableName        string
	Key              map[string]any
	UpdateExpression string
	ExpressionValues map[string]any
	ExpressionNames  map[string]string
	Condition        string
}

type TransactDelete struct {
	TableName string
	Key       map[string]any
	Condition string
}

type TransactWriteItem struct {
	Put    *TransactPut
	Update *TransactUpdate
	Delete *TransactDelete
}

type QueryMetadata struct {
	LastEvaluatedKey map[string]any
}

type DynamoDBClient interface {
	PutItem(ctx context.Context, tableName string, item map[string]any) error
	TransactWriteItems(ctx context.Context, items []TransactWriteItem) error
	GetItem(ctx context.Context, tableName string, key map[string]any, output any) error
	Query(
		ctx context.Context,
		tableName string,
		indexName string,
		limit int32,
		scanIndexForward bool,
		conditionExpression string,
		filterExpression string,
		expressionNames map[string]string,
		expressionValues map[string]any,
		startKey map[string]any,
		output any,
	) (QueryMetadata, error)
	UpdateItem(
		ctx context.Context,
		tableName string,
		key map[string]any,
		expression string,
		condition string,
		expressionNames map[string]string,
		expressionValues map[string]any,
	) error
	DeleteItem(ctx context.Context, tableName string, key map[string]any, condition string) error
}
