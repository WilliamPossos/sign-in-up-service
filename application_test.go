package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events/test"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"sign-in-up-service/model"
	"sign-in-up-service/repository"
	"sign-in-up-service/util"
	"strings"
	"testing"
)

type MockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

type MockUserRepository struct {
	mock.Mock
}

type MockSqsRepository struct {
	mock.Mock
}

func (m MockSqsRepository) Send(verification model.EmailConfig) error {
	args := m.Called(verification)
	return args.Error(0)
}

func (m MockUserRepository) Verify(email string, code string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m MockUserRepository) Search(email string) (*model.UserDynamoDb, error) {
	args := m.Called(email)
	return args.Get(0).(*model.UserDynamoDb), args.Error(1)
}

func (m MockUserRepository) Create(user model.UserDynamoDb) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockDynamoDBClient) Scan(*dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	return &dynamodb.ScanOutput{
		Items: []map[string]*dynamodb.AttributeValue{
			{
				"email": {
					S: aws.String("email"),
				},
				"code": {
					S: aws.String("1010"),
				}},
		},
	}, nil
}

func (m *MockDynamoDBClient) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func (m *MockDynamoDBClient) GetItem(*dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	itemOutput := dynamodb.GetItemOutput{Item: map[string]*dynamodb.AttributeValue{
		"email": {
			S: aws.String("possos@unicauca.edu.co"),
		},
		"password": {
			S: aws.String("9df0890d7e64d6cd56eb09086bc054dd718c9ecc608e27a03524a677b064982d"),
		},
		"code": {
			S: aws.String("1010"),
		}}}
	return &itemOutput, nil
}

func TestSignUp(t *testing.T) {
	ts := httptest.NewServer(setupGin())
	defer ts.Close()

	userRepositoryMock := new(MockUserRepository)
	sqsRepositoryMock := new(MockSqsRepository)
	sqsRepositoryMock.On("Send", mock.Anything).Return(nil)
	userRepositoryMock.On("Create", mock.MatchedBy(func(user model.UserDynamoDb) bool {
		return user.Email == "test@test.com" && user.Username == "test" && user.Password == "pass"
	})).Return(nil)

	userRepository = userRepositoryMock
	sqsRepository = sqsRepositoryMock

	inputJson := test.ReadJSONFromFile(t, "sign-up.json")
	resp, err := http.Post(fmt.Sprintf("%s/sign-up", ts.URL), "application/json", strings.NewReader(string(inputJson)))

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestSignIn(t *testing.T) {
	ts := httptest.NewServer(setupGin())
	defer ts.Close()

	userRepository = repository.UserRepository{DbClient: &MockDynamoDBClient{}}
	attemptRepository = repository.LoginAttemptRepository{DbClient: &MockDynamoDBClient{}}

	inputJson := test.ReadJSONFromFile(t, "sign-in.json")
	resp, err := http.Post(fmt.Sprintf("%s/sign-in", ts.URL), "application/json", strings.NewReader(string(inputJson)))

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestVerify(t *testing.T) {
	ts := httptest.NewServer(setupGin())
	defer ts.Close()

	userRepository = repository.UserRepository{DbClient: &MockDynamoDBClient{}}
	attemptRepository = repository.LoginAttemptRepository{DbClient: &MockDynamoDBClient{}}

	inputJson := test.ReadJSONFromFile(t, "verify.json")
	resp, err := http.Post(fmt.Sprintf("%s/verify", ts.URL), "application/json", strings.NewReader(string(inputJson)))

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestSaveUser(t *testing.T) {

	repository := repository.UserRepository{DbClient: &MockDynamoDBClient{}}
	var user = model.UserDynamoDb{
		Username:         "a",
		Email:            "b",
		Password:         "c",
		VerificationCode: "d",
	}
	err := repository.Create(user)

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}
}

func TestSearchUser(t *testing.T) {

	repository := repository.UserRepository{DbClient: &MockDynamoDBClient{}}
	var user = model.SignUpDto{
		Username: "stibent",
		Email:    "possos@unicauca.edu.co",
		Password: "possos",
	}
	found, err := repository.Search(user.Email)

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}

	if found.Password != util.GetHashPassword("possos") {
		t.Fatalf("Expected password match")
	}
}
