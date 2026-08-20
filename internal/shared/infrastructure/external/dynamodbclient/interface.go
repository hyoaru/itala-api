package dynamodbclient

import "context"

type DynamoDBClient interface {
	PutItem(ctx context.Context, tableName string, item map[string]any) error
}
