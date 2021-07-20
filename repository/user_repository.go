package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"os"
	"sign-in-up-service/model"
	"sign-in-up-service/util"
)

var tableName = os.Getenv("USER_TABLE_NAME")

type IUserRepository interface {
	Verify(email string, code string) (bool, error)
	Search(email string) (*model.UserDynamoDb, error)
	Create(user model.UserDynamoDb) error
}

type UserRepository struct {
	DbClient dynamodbiface.DynamoDBAPI
}

func (ur UserRepository) Verify(email string, code string) (bool, error) {

	user, err := ur.Search(email)
	if err != nil {
		return false, err
	}

	isValidCode := user != nil && user.VerificationCode == code
	return isValidCode, err
}

func (ur UserRepository) Search(email string) (*model.UserDynamoDb, error) {
	input := getEmailItemInput(email)
	user, err := util.GetItem(ur.DbClient, input)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ur UserRepository) Create(user model.UserDynamoDb) error {
	userToPut, err := getUserWithHashPassword(user)
	av, err := dynamodbattribute.MarshalMap(userToPut)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String(tableName),
		ConditionExpression: aws.String("attribute_not_exists (email)"),
	}

	_, err = ur.DbClient.PutItem(input)

	return err
}

func getEmailItemInput(email string) *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
	}
}

func getUserWithHashPassword(user model.UserDynamoDb) (*model.UserDynamoDb, error) {
	hashedPassword := util.GetHashPassword(user.Password)
	return &model.UserDynamoDb{
		Username:         user.Username,
		Email:            user.Email,
		Password:         hashedPassword,
		VerificationCode: user.VerificationCode,
	}, nil
}
