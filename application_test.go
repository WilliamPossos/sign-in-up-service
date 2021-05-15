package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events/test"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"sign-in-up-service/model"
	"sign-in-up-service/repository"
	"strings"
	"testing"
)

type MockDynamoDBClient struct {
	dynamodbiface.DynamoDBAPI
}

type MockUserRepository struct {
	mock.Mock
}

func (m MockUserRepository) Exist(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m MockUserRepository) Search(email string, password string) (bool, error) {
	args := m.Called(email, password)
	return args.Bool(0), args.Error(1)
}

func (m MockUserRepository) Create(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockDynamoDBClient) PutItem(*dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	return nil, nil
}

func TestSignUp(t *testing.T) {
	ts := httptest.NewServer(setupGin())
	defer ts.Close()

	userRepositoryMock := new(MockUserRepository)
	userRepositoryMock.On("Create", mock.MatchedBy(func(user model.User) bool {
		return user.Email == "test@test.com" && user.Username == "test" && user.Password == "pass"
	})).Return(nil)
	userRepository = userRepositoryMock

	inputJson := test.ReadJSONFromFile(t, "sign-up.json")
	resp, err := http.Post(fmt.Sprintf("%s/sign-up", ts.URL), "application/json", strings.NewReader(string(inputJson)))

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
	}
}

func TestSaveUser(t *testing.T) {

	repository := repository.UserRepository{DbClient: &MockDynamoDBClient{}}
	var user = model.User{
		Username: "1",
		Email:    "2",
		Password: "3",
	}
	err := repository.Create(user)

	if err != nil {
		t.Fatalf("Expected no error got %s", err)
	}

}
