package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"sign-in-up-service/model"
)

var LoginAttemptTableName = "LoginAttempt"

const (
	EmptyAttempts         = "EmptyAttempts"
	MaxFailedAttempts     = "MaxFailedAttempts"
	AllowedFailedAttempts = "AllowedFailedAttempts"
)

type ILoginAttemptRepository interface {
	Create(user model.LoginAttempt) error
	GetAttemptsValidation(email string, limit int) (string, error)
}

type LoginAttemptRepository struct {
	DbClient dynamodbiface.DynamoDBAPI
}

func (ar LoginAttemptRepository) GetAttemptsValidation(email string, limit int) (string, error) {
	filter := expression.Name("email").Equal(expression.Value(email))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()

	if err != nil {
		return "", err
	}

	// Build the query input parameters
	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		TableName:                 aws.String(LoginAttemptTableName),
	}

	// Make the DynamoDB Query API call
	result, err := ar.DbClient.Scan(params)

	if err != nil {
		return "", err
	}

	if len(result.Items) == 0 {
		return EmptyAttempts, nil
	}

	failedAttempts := 0

	for i := 0; i < limit; i++ {
		attempt := model.LoginAttempt{}

		err = dynamodbattribute.UnmarshalMap(result.Items[i], &attempt)
		if err != nil {
			return "", err
		}

		if attempt.Status != true {
			failedAttempts++
		}
	}

	if failedAttempts == limit {
		return MaxFailedAttempts, nil
	}

	return AllowedFailedAttempts, nil
}

func (ar LoginAttemptRepository) Create(attempt model.LoginAttempt) error {
	av, err := dynamodbattribute.MarshalMap(attempt)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(LoginAttemptTableName),
	}

	_, err = ar.DbClient.PutItem(input)

	return err
}
