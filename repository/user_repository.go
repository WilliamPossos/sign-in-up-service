package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"golang.org/x/crypto/bcrypt"
	"sign-in-up-service/model"
	"sign-in-up-service/util"
)

const tableName = "User"

type IUserRepository interface {
	Exist(email string) (bool, error)
	Search(email string, password string) (bool, error)
	Verify(email string, code string) (bool, error)
	Create(user model.UserDynamoDb) error
}

type UserRepository struct {
	DbClient dynamodbiface.DynamoDBAPI
}

func (ur UserRepository) Exist(email string) (bool, error) {
	input := getEmailItemInput(email)
	user, err := util.GetItem(ur.DbClient, input)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

func (ur UserRepository) Verify(email string, code string) (bool, error) {

	input := getEmailItemInput(email)
	user, err := util.GetItem(ur.DbClient, input)
	if err != nil {
		return false, err
	}

	return user != nil, nil
}

func (ur UserRepository) Search(email string, password string) (bool, error) {
	hashedPassword, err := getHashPassword(password)
	if err != nil {
		return false, err
	}

	input := getSignInItemInput(email, hashedPassword)
	user, err := util.GetItem(ur.DbClient, input)
	if err != nil {
		return false, err
	}

	return user != nil, nil
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
			"username": {
				S: aws.String("stibent"),
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

func getUserWithHashPassword(user model.UserDynamoDb) (*model.UserDynamoDb, error) {
	hashedPassword, err := getHashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	return &model.UserDynamoDb{
		Username:         user.Username,
		Email:            user.Email,
		Password:         hashedPassword,
		VerificationCode: user.VerificationCode,
	}, nil
}

func getHashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
