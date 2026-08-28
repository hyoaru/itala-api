package dynamodbclient

import "context"

type PutItemInput struct {
	TableName           string
	Item                map[string]any
	ConditionExpression *string
}

type GetItemInput struct {
	TableName string
	Key       map[string]any
	Output    any
}

type QueryInput struct {
	TableName                 string
	IndexName                 *string
	Limit                     *int32
	ScanIndexForward          *bool
	KeyConditionExpression    *string
	FilterExpression          *string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]any
	ExclusiveStartKey         map[string]any
	Output                    any
}

type QueryOutput struct {
	LastEvaluatedKey map[string]any
}

type UpdateItemInput struct {
	TableName                 string
	Key                       map[string]any
	UpdateExpression          string
	ConditionExpression       *string
	ExpressionAttributeNames  map[string]string
	ExpressionAttributeValues map[string]any
}

type DeleteItemInput struct {
	TableName           string
	Key                 map[string]any
	ConditionExpression *string
}

type TransactWriteItemsInput struct {
	TransactItems []TransactWriteItem
}

type TransactPut struct {
	TableName           string
	Item                map[string]any
	ConditionExpression *string
}

type TransactUpdate struct {
	TableName                 string
	Key                       map[string]any
	UpdateExpression          string
	ExpressionAttributeValues map[string]any
	ExpressionAttributeNames  map[string]string
	ConditionExpression       *string
}

type TransactDelete struct {
	TableName           string
	Key                 map[string]any
	ConditionExpression *string
}

type TransactWriteItem struct {
	Put    *TransactPut
	Update *TransactUpdate
	Delete *TransactDelete
}

type DynamoDBClient interface {
	PutItem(ctx context.Context, input *PutItemInput) error
	TransactWriteItems(ctx context.Context, input *TransactWriteItemsInput) error
	GetItem(ctx context.Context, input *GetItemInput) error
	Query(ctx context.Context, input *QueryInput) (QueryOutput, error)
	UpdateItem(ctx context.Context, input *UpdateItemInput) error
	DeleteItem(ctx context.Context, input *DeleteItemInput) error
}
