package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"sign-in-up-service/model"
	"sign-in-up-service/repository"
	"time"
)

var ginLambda *ginadapter.GinLambda

var sess *session.Session
var dynamoClient dynamodbiface.DynamoDBAPI
var sqsClient sqsiface.SQSAPI
var userRepository repository.IUserRepository
var sqsRepository repository.SqsEmailService

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	dynamoClient = dynamodb.New(sess)
	sqsClient = sqs.New(sess)
	sqsRepository = repository.SqsEmailService{SqsClient: sqsClient}
	userRepository = repository.UserRepository{DbClient: dynamoClient}
	ginLambda = ginadapter.New(setupGin())
}

func setupGin() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.POST("/sign-up", func(c *gin.Context) {
		signUpHandler(c)
	})

	router.POST("/sign-in", func(c *gin.Context) {
		signInHandler(c)
	})

	return router
}

func signUpHandler(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := userRepository.Create(user)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := GetRandomCode()
	verification := model.EmailVerification{
		Email: user.Email,
		Code:  code,
	}
	err = sqsRepository.Send(verification)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func GetRandomCode() string {
	rand.Seed(time.Now().UnixNano())
	min := 1001
	max := 9998
	code := fmt.Sprintf("%d", rand.Intn(max-min+1)+min)
	return code
}

func signInHandler(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := userRepository.Search(user.Email, user.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
	//setupGin().Run(":3000")
}
