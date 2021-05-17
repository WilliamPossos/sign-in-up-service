package repository

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"os"
	"sign-in-up-service/model"
)

var QueueUrl = os.Getenv("QUEUE_URL")

type ISqsEmailService interface {
	Send(verification model.EmailVerification) error
}

type SqsEmailService struct {
	SqsClient sqsiface.SQSAPI
}

func (s SqsEmailService) Send(verification model.EmailVerification) error {
	verificationBytes, err := json.Marshal(verification)
	if err != nil {
		return err
	}

	_, err = s.SqsClient.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(string(verificationBytes)),
		QueueUrl:     &QueueUrl,
	})

	if err != nil {
		return err
	}

	return nil
}
