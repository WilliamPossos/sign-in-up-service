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
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"log"
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
var attemptRepository repository.ILoginAttemptRepository
var sqsRepository repository.ISqsEmailService

const (
	MaxAttempts = 3
	AuthSuccess = "AuthSuccess"
	AuthError   = "AuthError"
)

func init() {
	sess = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	dynamoClient = dynamodb.New(sess)
	sqsClient = sqs.New(sess)
	sqsRepository = repository.SqsEmailService{SqsClient: sqsClient}
	userRepository = repository.UserRepository{DbClient: dynamoClient}
	attemptRepository = repository.LoginAttemptRepository{DbClient: dynamoClient}
	ginLambda = ginadapter.New(setupGin())
}

func setupGin() *gin.Engine {
	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{}
	router.Use(cors.New(config))

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://challenge-3.goodwilli.com", "http://localhost:4200"},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))

	router.POST("/sign-up", func(c *gin.Context) {
		signUpHandler(c)
	})

	router.POST("/sign-in", func(c *gin.Context) {
		signInHandler(c)
	})

	router.POST("/verify", func(c *gin.Context) {
		verifyHandler(c)
	})

	return router
}

func verifyHandler(c *gin.Context) {
	var user model.SignInDto
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	isSuccessVerified, err := userRepository.Verify(user.Email, user.VerificationCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if isSuccessVerified != true {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error validating code"})
		return
	}

	saveAttemptRecord(user, true)
	c.JSON(http.StatusOK, model.OperationResult{
		Message: AuthSuccess,
	})
}

func signUpHandler(c *gin.Context) {
	var user model.SignUpDto
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := GetRandomCode()

	err := userRepository.Create(model.UserDynamoDb{
		Username:         user.Username,
		Email:            user.Email,
		Password:         user.Password,
		VerificationCode: code,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

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
	var user model.SignInDto
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	signInValidation, err := userSingIn(model.SignInDto{
		Email:            user.Email,
		Password:         user.Password,
		VerificationCode: user.VerificationCode,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.OperationResult{Message: signInValidation})
}

func userSingIn(user model.SignInDto) (string, error) {
	validation, err := attemptRepository.GetAttemptsValidation(user.Email, MaxAttempts)
	if err != nil {
		return AuthError, err
	}

	if validation == repository.AllowedFailedAttempts {
		isLoginSuccess, err := userRepository.Search(user.Email, user.Password)
		saveAttemptRecord(user, isLoginSuccess)

		if err != nil {
			return AuthError, err
		}

		return AuthSuccess, nil
	}

	return validation, nil
}

func saveAttemptRecord(user model.SignInDto, isLoginSuccess bool) {
	errAttempt := attemptRepository.Create(model.LoginAttempt{
		Email:  user.Email,
		Time:   fmt.Sprint(time.Now().Unix()),
		Status: isLoginSuccess,
	})

	if errAttempt != nil {
		log.Print("Error saving attempt")
	}
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// If no name is provided in the HTTP request body, throw an error
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(Handler)
	//setupGin().Run(":3000")
}
