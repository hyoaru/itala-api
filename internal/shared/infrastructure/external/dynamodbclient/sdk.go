package dynamodbclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type SDKDynamoDBClient struct {
	client *dynamodb.Client
}

func NewSDKDynamoDBClient() *SDKDynamoDBClient {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(fmt.Errorf("load AWS config: %w", err))
	}

	return &SDKDynamoDBClient{
		client: dynamodb.NewFromConfig(cfg),
	}
}

func (c *SDKDynamoDBClient) PutItem(ctx context.Context, tableName string, item map[string]any) error {
	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("marshal map: %w", err)
	}

	_, err = c.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return ErrItemExists
		}

		return err
	}

	return nil
}
