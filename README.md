# CloudChallenge3 - golang server - lambda component

This project was generating to solve the cloud challenge number 3.

The proposed AWS architecture is:

![AWS architecture](https://joiners-challenge.s3.amazonaws.com/aws-challenge-3.jpeg)

## Site URL

    `https://challenge-3.goodwilli.com/`

## Features

- The back-end is built with Golang
- Important packages: 
    - AWS SDK for Go
    - Gin (github.com/gin-gonic/gin)
    - AWS lambda (github.com/aws/aws-lambda-go)
- This repository has continuous integration for code test and build
- This repository has continuous deployment that posts the generated files into a lambda function
    - lambda name: **sign-up (us-east-1)**

![AWS architecture](https://joiners-challenge.s3.amazonaws.com/go-lambda-pipeline.png)

## Go commons

Build application command:
    
    `GOOS=linux GOARCH=amd64 go build -o application application.go`

Zip application command:

    `zip application.zip application`

Go test command:

    `go test`

!!! Important
For local test comment the `lambda.Start(Handler)` and uncomment `setupGin().Run(":3000")` in the main method.