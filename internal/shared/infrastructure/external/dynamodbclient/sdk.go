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
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
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

type Decimal valueobject.Decimal

func (d Decimal) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberN{
		Value: valueobject.Decimal(d).String(),
	}, nil
}

func (c *SDKDynamoDBClient) PutItem(ctx context.Context, input *PutItemInput) error {
	av, err := attributevalue.MarshalMap(input.Item)
	if err != nil {
		return fmt.Errorf("marshal map: %w", err)
	}

	conditionExpression := "attribute_not_exists(PK)"
	if input.ConditionExpression != nil {
		conditionExpression = *input.ConditionExpression
	}

	putItemInput := &dynamodb.PutItemInput{
		TableName:                aws.String(input.TableName),
		Item:                     av,
		ConditionExpression:      aws.String(conditionExpression),
		ExpressionAttributeNames: input.ExpressionAttributeNames,
	}

	if input.ExpressionAttributeValues != nil {
		parsedExpressionValues, err := attributevalue.MarshalMap(input.ExpressionAttributeValues)
		if err != nil {
			return fmt.Errorf("marshal expression values: %w", err)
		}

		putItemInput.ExpressionAttributeValues = parsedExpressionValues
	}

	_, err = c.client.PutItem(ctx, putItemInput)
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return ErrItemExists
		}

		return fmt.Errorf("put item: %w", err)
	}

	return nil
}

func (c *SDKDynamoDBClient) TransactWriteItems(ctx context.Context, input *TransactWriteItemsInput) error {
	parsedTransactItems := make([]types.TransactWriteItem, 0, len(input.TransactItems))

	for _, item := range input.TransactItems {
		parsedTransactItem, err := c.toTransactWriteItem(item)
		if err != nil {
			return fmt.Errorf("marshal transact write item: %w", err)
		}

		parsedTransactItems = append(parsedTransactItems, parsedTransactItem)
	}

	_, err := c.client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{TransactItems: parsedTransactItems})
	if err != nil {
		if canceled, ok := errors.AsType[*types.TransactionCanceledException](err); ok {
			for _, reason := range canceled.CancellationReasons {
				if reason.Code != nil && *reason.Code == "ConditionalCheckFailed" {
					return ErrConditionFailed
				}
			}
		}

		return err
	}

	return nil
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
		TableName:           aws.String(operation.TableName),
		Item:                parsedItem,
		ConditionExpression: operation.ConditionExpression,
	}

	return types.TransactWriteItem{Put: item}, nil
}

func (c *SDKDynamoDBClient) toTransactUpdate(operation *TransactUpdate) (types.TransactWriteItem, error) {
	key, err := attributevalue.MarshalMap(operation.Key)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	values, err := attributevalue.MarshalMap(operation.ExpressionAttributeValues)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	item := &types.Update{
		TableName:                 aws.String(operation.TableName),
		Key:                       key,
		UpdateExpression:          aws.String(operation.UpdateExpression),
		ConditionExpression:       operation.ConditionExpression,
		ExpressionAttributeValues: values,
		ExpressionAttributeNames:  operation.ExpressionAttributeNames,
	}

	return types.TransactWriteItem{Update: item}, nil
}

func (c *SDKDynamoDBClient) toTransactDelete(operation *TransactDelete) (types.TransactWriteItem, error) {
	key, err := attributevalue.MarshalMap(operation.Key)
	if err != nil {
		return types.TransactWriteItem{}, err
	}

	item := &types.Delete{
		TableName:           aws.String(operation.TableName),
		Key:                 key,
		ConditionExpression: operation.ConditionExpression,
	}

	return types.TransactWriteItem{Delete: item}, nil
}

func (c *SDKDynamoDBClient) Query(ctx context.Context, input *QueryInput, output any) (QueryOutput, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:                aws.String(input.TableName),
		Limit:                    input.Limit,
		ScanIndexForward:         input.ScanIndexForward,
		KeyConditionExpression:   input.KeyConditionExpression,
		IndexName:                input.IndexName,
		FilterExpression:         input.FilterExpression,
		ExpressionAttributeNames: input.ExpressionAttributeNames,
	}

	if input.ExpressionAttributeValues != nil {
		parsedExpressionValues, err := attributevalue.MarshalMap(input.ExpressionAttributeValues)
		if err != nil {
			return QueryOutput{}, fmt.Errorf("marshal expression values: %w", err)
		}
		queryInput.ExpressionAttributeValues = parsedExpressionValues
	}

	if input.ExclusiveStartKey != nil {
		parsedStartKey, err := attributevalue.MarshalMap(input.ExclusiveStartKey)
		if err != nil {
			return QueryOutput{}, fmt.Errorf("marshal start key: %w", err)
		}
		queryInput.ExclusiveStartKey = parsedStartKey
	}

	queryOutput, err := c.client.Query(ctx, queryInput)
	if err != nil {
		return QueryOutput{}, fmt.Errorf("query items: %w", err)
	}

	if err := attributevalue.UnmarshalListOfMaps(queryOutput.Items, output); err != nil {
		return QueryOutput{}, fmt.Errorf("unmarshal items: %w", err)
	}

	if queryOutput.LastEvaluatedKey == nil {
		return QueryOutput{}, nil
	}

	var lastEvaluatedKey map[string]any
	if err = attributevalue.UnmarshalMap(queryOutput.LastEvaluatedKey, &lastEvaluatedKey); err != nil {
		return QueryOutput{}, fmt.Errorf("unmarshal last evaluated key: %w", err)
	}

	return QueryOutput{LastEvaluatedKey: lastEvaluatedKey}, nil
}

func (c *SDKDynamoDBClient) GetItem(ctx context.Context, input *GetItemInput, output any) error {
	parsedKey, err := attributevalue.MarshalMap(input.Key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	getItemOutput, err := c.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(input.TableName),
		Key:       parsedKey,
	})
	if err != nil {
		if _, ok := errors.AsType[*types.ResourceNotFoundException](err); ok {
			return ErrItemNotFound
		}
		return fmt.Errorf("get item: %w", err)
	}

	if getItemOutput.Item == nil {
		return ErrItemNotFound
	}

	if err := attributevalue.UnmarshalMap(getItemOutput.Item, output); err != nil {
		return fmt.Errorf("unmarshal item: %w", err)
	}

	return nil
}

func (c *SDKDynamoDBClient) UpdateItem(ctx context.Context, input *UpdateItemInput) error {
	parsedKey, err := attributevalue.MarshalMap(input.Key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	parsedExpressionValues, err := attributevalue.MarshalMap(input.ExpressionAttributeValues)
	if err != nil {
		return fmt.Errorf("marshal expression values: %w", err)
	}

	updateItemInput := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(input.TableName),
		Key:                       parsedKey,
		UpdateExpression:          aws.String(input.UpdateExpression),
		ConditionExpression:       input.ConditionExpression,
		ExpressionAttributeNames:  input.ExpressionAttributeNames,
		ExpressionAttributeValues: parsedExpressionValues,
	}

	_, err = c.client.UpdateItem(ctx, updateItemInput)
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return ErrConditionFailed
		}

		return fmt.Errorf("update item: %w", err)
	}

	return nil
}

func (c *SDKDynamoDBClient) DeleteItem(ctx context.Context, input *DeleteItemInput) error {
	parsedKey, err := attributevalue.MarshalMap(input.Key)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	parsedExpressionValues, err := attributevalue.MarshalMap(input.ExpressionAttributeValues)
	if err != nil {
		return fmt.Errorf("marshal expression values: %w", err)
	}

	deleteItemInput := &dynamodb.DeleteItemInput{
		TableName:                 aws.String(input.TableName),
		Key:                       parsedKey,
		ConditionExpression:       input.ConditionExpression,
		ExpressionAttributeNames:  input.ExpressionAttributeNames,
		ExpressionAttributeValues: parsedExpressionValues,
	}

	_, err = c.client.DeleteItem(ctx, deleteItemInput)
	if err != nil {
		if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); ok {
			return ErrConditionFailed
		}

		return fmt.Errorf("delete item: %w", err)
	}

	return nil
}
