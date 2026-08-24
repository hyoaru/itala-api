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

type DynamoDBClient interface {
	PutItem(ctx context.Context, tableName string, item map[string]any) error
	TransactWriteItems(ctx context.Context, items []TransactWriteItem) error
	Query(
		ctx context.Context,
		tableName string,
		conditionExpression string,
		filterExpression string,
		expressionValues map[string]any,
		result any,
	) error
}
