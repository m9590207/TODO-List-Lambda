package todo

import (
	"encoding/json"
	"errors"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/google/uuid"
)

var (
	ErrorFailedToUnmarshalRecord = "failed to unmarshal record"
	ErrorFailedToFetchRecord     = "failed to fetch record"
	ErrorInvalidData             = "invalid data"
	ErrorCouldNotMarshalItem     = "could not marshal item"
	ErrorCouldNotDeleteItem      = "could not delete item"
	ErrorCouldNotDynamoPutItem   = "could not dynamo put item"
	ErrorDoesNotExist            = "does not exist"
)

type Todo struct {
	Id        string `json:"id"`
	Item      string `json:"item"`
	CreatedBy string `json:"createdBy"`
	State     int    `json:"state"`
}

func ListTodo(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*[]Todo, error) {
	createdBy := req.QueryStringParameters["createdBy"]
	state, err := strconv.Atoi(req.QueryStringParameters["state"])
	if len(createdBy) == 0 || err != nil {
		return nil, errors.New(ErrorInvalidData)
	}

	cond1 := expression.Name("createdBy").Equal(expression.Value(createdBy))
	cond2 := expression.Name("state").Equal(expression.Value(state))
	proj := expression.NamesList(
		expression.Name("id"), expression.Name("item"), expression.Name("createdBy"), expression.Name("state"))
	expr, _ := expression.NewBuilder().WithFilter(cond1.And(cond2)).WithProjection(proj).Build()
	input := &dynamodb.ScanInput{
		TableName:                 aws.String(tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
	}

	result, err := dynaClient.Scan(input)
	if err != nil {
		log.Printf("Scan err: %s", err.Error())
		return nil, errors.New(ErrorFailedToFetchRecord)
	}
	item := new([]Todo)
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, item)
	return item, err
}

func GetTodo(id, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*Todo, error) {
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(tableName),
	}

	result, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	item := new(Todo)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New(ErrorFailedToUnmarshalRecord)
	}
	return item, nil
}

func CreateTodo(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*Todo, error) {
	var t Todo
	if err := json.Unmarshal([]byte(req.Body), &t); err != nil {
		return nil, errors.New(ErrorInvalidData)
	}
	t.Id = uuid.New().String()
	av, err := dynamodbattribute.MarshalMap(t)

	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		log.Printf("PutItem err: %s", err.Error())
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}
	return &t, nil
}

func UpdateTodo(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*Todo, error) {
	var t Todo
	if err := json.Unmarshal([]byte(req.Body), &t); err != nil {
		return nil, errors.New(ErrorInvalidData)
	}

	if len(t.Id) == 0 {
		return nil, errors.New(ErrorInvalidData)
	}

	curTodo, _ := GetTodo(t.Id, tableName, dynaClient)
	if curTodo != nil && len(curTodo.CreatedBy) == 0 {
		return nil, errors.New(ErrorDoesNotExist)
	}

	t.CreatedBy = curTodo.CreatedBy
	if len(t.Item) == 0 {
		t.Item = curTodo.Item
	}
	av, err := dynamodbattribute.MarshalMap(t)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		log.Printf("PutItem err: %s", err.Error())
		return nil, errors.New(ErrorCouldNotDynamoPutItem)
	}
	return &t, nil
}

func DeleteTodo(req events.APIGatewayProxyRequest, tableName string, dynaClient dynamodbiface.DynamoDBAPI) error {
	id := req.QueryStringParameters["id"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(id),
			},
		},
		TableName: aws.String(tableName),
	}
	_, err := dynaClient.DeleteItem(input)
	if err != nil {
		return errors.New(ErrorCouldNotDeleteItem)
	}
	return nil
}
