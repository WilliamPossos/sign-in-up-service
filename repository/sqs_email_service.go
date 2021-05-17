package repository

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"sign-in-up-service/model"
)

var QueueUrl = "QueueURL"

type ISqsEmailService interface {
	Send(verification model.EmailVerification) error
}

type SqsEmailService struct {
	SqsClient sqsiface.SQSAPI
}

func (s SqsEmailService) Send(verification model.EmailVerification) error {
	_, err := s.SqsClient.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageAttributes: map[string]*sqs.MessageAttributeValue{
			"Email": {
				DataType:    aws.String("String"),
				StringValue: aws.String(verification.Email),
			},
			"Code": {
				DataType:    aws.String("String"),
				StringValue: aws.String(verification.Code),
			},
		},
		MessageBody: aws.String("user verification code"),
		QueueUrl:    &QueueUrl,
	})

	if err != nil {
		fmt.Println("Error", err)
	}

	return err
}
