package util

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"sign-in-up-service/model"
)

func GetItem(dbClient dynamodbiface.DynamoDBAPI, input *dynamodb.GetItemInput) (*model.UserDynamoDb, error) {
	result, err := dbClient.GetItem(input)
	if err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, errors.New("could not find the item")
	}

	item := model.UserDynamoDb{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func GetHashPassword(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	return fmt.Sprintf("%x", hash.Sum(nil))

}
