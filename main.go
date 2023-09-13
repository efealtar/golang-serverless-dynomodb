package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Todo struct {
	ID   string `json:"id"`
	Task string `json:"task"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case "POST":
		return CreateHandler(request)
	case "GET":
		return ReadHandler(request)
	case "PUT":
		return UpdateHandler(request)
	case "DELETE":
		return DeleteHandler(request)
	default:
		return events.APIGatewayProxyResponse{StatusCode: 405, Body: "Method Not Allowed"}, nil
	}
}

func CreateHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	svc := dynamodb.New(sess)

	// Assuming the request body contains a JSON with the Todo task
	var todo Todo
	err := json.Unmarshal([]byte(request.Body), &todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid input"}, err
	}

	// Generate a unique ID for the Todo (for simplicity, using a UUID)//
	todo.ID = uuid.New().String()

	av, err := dynamodbattribute.MarshalMap(todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	body, err := json.Marshal(todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(body)}, nil
}

func ReadHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	svc := dynamodb.New(sess)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(request.PathParameters["id"]),
			},
		},
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
	}

	result, err := svc.GetItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	if result.Item == nil {
		return events.APIGatewayProxyResponse{StatusCode: 404, Body: "Todo not found"}, nil
	}

	var todo Todo
	err = dynamodbattribute.UnmarshalMap(result.Item, &todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	body, err := json.Marshal(todo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: string(body)}, nil
}

func UpdateHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	svc := dynamodb.New(sess)

	// Assuming the request body contains a JSON with the updated Todo task
	var updatedTodo Todo
	err := json.Unmarshal([]byte(request.Body), &updatedTodo)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid input: " + err.Error()}, nil
	}

	if updatedTodo.Task == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Task cannot be empty"}, nil
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":n": {
				S: aws.String(updatedTodo.Task),
			},
		},
		ExpressionAttributeNames: map[string]*string{
			"#N": aws.String("task"),
		},
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(request.PathParameters["id"]),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set #N = :n"),
	}

	_, err = svc.UpdateItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Update failed: " + err.Error()}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Todo updated successfully!"}, nil
}

func DeleteHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	}))

	svc := dynamodb.New(sess)

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(request.PathParameters["id"]),
			},
		},
		TableName: aws.String(os.Getenv("DYNAMODB_TABLE")),
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Todo deleted successfully!"}, nil
}

func main() {
	lambda.Start(handler)
}
