package repository

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"golang.org/x/crypto/bcrypt"
	"sign-in-up-service/model"
)

var tableName = "User"

type IUserRepository interface {
	Exist(email string) (bool, error)
	Search(email string, password string) (bool, error)
	Create(user model.User) error
}

type UserRepository struct {
	DbClient dynamodbiface.DynamoDBAPI
}

func (ur UserRepository) Exist(email string) (bool, error) {
	input := getEmailItemInput(email)
	return GetItem(ur, input)
}

func (ur UserRepository) Search(email string, password string) (bool, error) {
	hashedPassword, err := getHashPassword(password)
	if err != nil {
		return false, err
	}
	input := getSignInItemInput(email, hashedPassword)
	return GetItem(ur, input)
}

func (ur UserRepository) Create(user model.User) error {
	userToPut, err := getUserWithHashPassword(user)
	av, err := dynamodbattribute.MarshalMap(userToPut)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = ur.DbClient.PutItem(input)

	return err
}

func GetItem(ur UserRepository, input *dynamodb.GetItemInput) (bool, error) {
	result, err := ur.DbClient.GetItem(input)
	if err != nil {
		return false, err
	}

	if result.Item == nil {
		return false, errors.New("could not find the item")
	}

	item := model.User{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	if err != nil {
		return false, err
	}

	return true, nil
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

func getSignInItemInput(email string, password string) *dynamodb.GetItemInput {
	return &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
			"password": {
				S: aws.String(password),
			},
		},
	}
}

func getUserWithHashPassword(user model.User) (*model.User, error) {
	hashedPassword, err := getHashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	return &model.User{
		Username: user.Username,
		Email:    user.Email,
		Password: hashedPassword,
	}, nil
}

func getHashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
