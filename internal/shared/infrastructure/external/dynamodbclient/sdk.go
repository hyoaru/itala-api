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
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
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

type Decimal valueobjects.Decimal

func (d Decimal) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberN{
		Value: valueobjects.Decimal(d).String(),
	}, nil
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

func (c *SDKDynamoDBClient) TransactWriteItems(ctx context.Context, items []TransactWriteItem) error {
	parsedTransactItems := make([]types.TransactWriteItem, 0, len(items))

	for _, item := range items {
		parsedTransactItem, err := c.toTransactWriteItem(item)
		if err != nil {
			return err
		}

		parsedTransactItems = append(parsedTransactItems, parsedTransactItem)
	}

	_, err := c.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: parsedTransactItems})
	if err != nil {
		if canceled, ok := errors.AsType[*types.TransactionCanceledException](err); ok {
			for _, reason := range canceled.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return ErrItemExists
				}
			}
		}

		return err
	}

	return err
}

func (c *SDKDynamoDBClient) toTransactWriteItem(writeItem TransactWriteItem) (types.TransactWriteItem, error) {
	switch {
	case writeItem.Put != nil:
		return c.toTransactPut(writeItem.Put)
	case writeItem.Update != nil:
		return c.toTransactUpdate(writeItem.Update)
	case writeItem.Delete != nil:
		return c.toTransactDelete(writeItem.Delete)
	default:
		return types.TransactWriteItem{}, errors.New("write operation must contain put, update, or delete")
	}
}

func (c *SDKDynamoDBClient) toTransactPut(operation *TransactPut) (types.TransactWriteItem, error) {
	parsedItem, err := attributevalue.MarshalMap(operation.Item)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	item := &types.Put{
		TableName: aws.String(operation.TableName),
		Item:      parsedItem,
	}

	if operation.Condition != "" {
		item.ConditionExpression = aws.String(operation.Condition)
	}

	return types.TransactWriteItem{Put: item}, nil
}

func (c *SDKDynamoDBClient) toTransactUpdate(operation *TransactUpdate) (types.TransactWriteItem, error) {
	key, err := attributevalue.MarshalMap(operation.Key)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	values, err := attributevalue.MarshalMap(operation.ExpressionValues)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	item := &types.Update{
		TableName:                 aws.String(operation.TableName),
		Key:                       key,
		UpdateExpression:          aws.String(operation.UpdateExpression),
		ExpressionAttributeValues: values,
		ExpressionAttributeNames:  operation.ExpressionNames,
	}

	if operation.Condition != "" {
		item.ConditionExpression = aws.String(operation.Condition)
	}

	return types.TransactWriteItem{Update: item}, nil
}

func (c *SDKDynamoDBClient) toTransactDelete(operation *TransactDelete) (types.TransactWriteItem, error) {
	key, err := attributevalue.MarshalMap(operation.Key)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	item := &types.Delete{
		TableName: aws.String(operation.TableName),
		Key:       key,
	}

	if operation.Condition != "" {
		item.ConditionExpression = aws.String(operation.Condition)
	}

	return types.TransactWriteItem{Delete: item}, nil
}

func (c *SDKDynamoDBClient) Query(
	ctx context.Context,
	tableName string,
	limit int32,
	conditionExpression string,
	filterExpression string,
	expressionValues map[string]any,
	startKey map[string]any,
	output any,
) (QueryMetadata, error) {
	parsedExpressionValues, err := attributevalue.MarshalMap(expressionValues)
	if err != nil {
		return QueryMetadata{}, fmt.Errorf("marshal expression values: %w", err)
	}

	queryInput := &dynamodb.QueryInput{
		Limit:                     aws.Int32(limit),
		TableName:                 aws.String(tableName),
		KeyConditionExpression:    aws.String(conditionExpression),
		ExpressionAttributeValues: parsedExpressionValues,
	}

	if startKey != nil {
		parsedStartKey, err := attributevalue.MarshalMap(startKey)
		if err != nil {
			return QueryMetadata{}, fmt.Errorf("marshal start key: %w", err)
		}
		queryInput.ExclusiveStartKey = parsedStartKey
	}

	if len(filterExpression) > 0 {
		queryInput.FilterExpression = aws.String(filterExpression)
	}

	queryOutput, err := c.client.Query(ctx, queryInput)
	if err != nil {
		return QueryMetadata{}, fmt.Errorf("query items: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(queryOutput.Items, output); err != nil {
		return QueryMetadata{}, fmt.Errorf("unmarshal items: %w", err)
	}

	if queryOutput.LastEvaluatedKey == nil {
		return QueryMetadata{}, nil
	}

	var lastEvaluatedKey map[string]any
	if err = attributevalue.UnmarshalMap(queryOutput.LastEvaluatedKey, &lastEvaluatedKey); err != nil {
		return QueryMetadata{}, fmt.Errorf("unmarshal last evaluated key: %w", err)
	}

	return QueryMetadata{LastEvaluatedKey: lastEvaluatedKey}, nil
}
