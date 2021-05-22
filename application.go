package main

import (
	"context"
	"errors"
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
	"sign-in-up-service/util"
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
	MaxAttempts               = 3
	SignInSuccess             = "SignInSuccess"
	SignUpSuccess             = "SignUpSuccess"
	InvalidUsernameOrPassword = "invalid username or password"
	VerificationSubject       = "Challenge 3 verification code"
	VerificationHtmlBody      = "<h1>Verification code</h1>	<p>%s</p>"
	UserLockSubject           = "Challenge 3 user locked"
	UserLockHtmlBody          = "<p>The user has been blocked. We are working on the unlock functionality.</p>"
)

func init() {

	dynamoClient = dynamodb.New(sess)
	sqsClient = sqs.New(sess)
	sqsRepository = repository.SqsEmailService{SqsClient: sqsClient}
	userRepository = repository.UserRepository{DbClient: dynamoClient}
	attemptRepository = repository.LoginAttemptRepository{DbClient: dynamoClient}
	ginLambda = ginadapter.New(setupGin())
}

func setupGin() *gin.Engine {
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://challenge-3.goodwilli.com", "http://localhost:4200"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Accept", "Origin", "Cache-Control", "X-Requested-With", "Access-Control-Allow-Origin", "Access-Control-Allow-Methods"},
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
		OkMessage: SignInSuccess,
	})
}

func signUpHandler(c *gin.Context) {
	var user model.SignUpDto
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	code := getRandomCode()

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

	verification := model.EmailConfig{
		Email:    user.Email,
		HtmlBody: fmt.Sprintf(VerificationHtmlBody, code),
		Subject:  VerificationSubject,
	}
	err = sqsRepository.Send(verification)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.OperationResult{OkMessage: SignUpSuccess})
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

	c.JSON(http.StatusOK, model.OperationResult{OkMessage: signInValidation})
}

func userSingIn(signInDto model.SignInDto) (string, error) {
	userFound, err := userRepository.Search(signInDto.Email)
	if err != nil {
		return "", err
	}

	if userFound == nil {
		return "", errors.New(InvalidUsernameOrPassword)
	}

	validation, err := attemptRepository.GetAttemptsValidation(signInDto.Email, MaxAttempts)
	if validation == repository.AllowedFailedAttempts {
		hashPassword := util.GetHashPassword(signInDto.Password)
		if hashPassword == userFound.Password {
			saveSuccessAttemptRecord(signInDto)
			return SignInSuccess, nil
		}
		saveFailedAttemptRecord(signInDto)
		return "", errors.New(InvalidUsernameOrPassword)
	}
	if validation == repository.MaxFailedAttempts {
		verification := model.EmailConfig{
			Email:    userFound.Email,
			HtmlBody: UserLockHtmlBody,
			Subject:  UserLockSubject,
		}
		err = sqsRepository.Send(verification)

	}

	return validation, nil
}

func getRandomCode() string {
	rand.Seed(time.Now().UnixNano())
	min := 1001
	max := 9998
	code := fmt.Sprintf("%d", rand.Intn(max-min+1)+min)
	return code
}

func saveFailedAttemptRecord(signInDto model.SignInDto) {
	saveAttemptRecord(signInDto, false)
}

func saveSuccessAttemptRecord(signInDto model.SignInDto) {
	saveAttemptRecord(signInDto, true)
}

func saveAttemptRecord(user model.SignInDto, isSuccessAttempt bool) {
	errAttempt := attemptRepository.Create(model.LoginAttempt{
		Email:  user.Email,
		Time:   fmt.Sprint(time.Now().Unix()),
		Status: isSuccessAttempt,
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
