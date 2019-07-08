package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/mailru/easyjson"
	"monzo-transactions/monzo"
	"os"
)

type LambdaFunction struct {
	table *string
	db    *dynamodb.DynamoDB
}

func (f *LambdaFunction) Invoke(ctx context.Context, req events.SNSEvent) error {
	for _, e := range req.Records {

		var tx monzo.TransactionCreated
		err := easyjson.Unmarshal([]byte(e.SNS.Message), &tx)
		if err != nil {
			return err
		}

		attr, err := dynamodbattribute.MarshalMap(&tx)
		if err != nil {
			return err
		}

		_, err = f.db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
			TableName: f.table,
			Item:      attr,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	s := session.New()
	f := &LambdaFunction{
		table: aws.String(os.Getenv("LAMBDA_DYNAMODB_TABLE")),
		db:    dynamodb.New(s),
	}
	lambda.Start(f.Invoke)
}
