package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/dgrijalva/jwt-go"
	"github.com/mailru/easyjson"
	"log"
	"monzo-transactions/monzo"
	"os"
	"sort"
)

type LambdaFunction struct {
	l      *log.Logger
	s      *session.Session
	sns    *sns.SNS
	secret []byte
	topic  string
}

func (f *LambdaFunction) validateToken(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return f.secret, nil
}

func (f *LambdaFunction) accountClaims(token *jwt.Token) ([]string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !token.Valid || !ok {
		return nil, errors.New("invalid JWT token")
	}
	accountClaims, ok := claims["act"]
	if !ok {
		return nil, errors.New("no act claims in JWT token")
	}

	accountClaimsSlice, ok := accountClaims.([]interface{})
	if !ok {
		return nil, errors.New("act claims is not a list")
	}

	accounts := make([]string, len(accountClaimsSlice))
	for i, x := range accountClaimsSlice {
		s, ok := x.(string)
		if !ok {
			return nil, errors.New("act claims contain a non-string")
		}
		accounts[i] = s
	}

	sort.Strings(accounts)
	return accounts, nil
}

func (f *LambdaFunction) checkIsValidAccount(accounts []string, e *monzo.TransactionCreated) error {
	for _, a := range accounts {
		if a == e.AccountID {
			return nil
		}
	}
	return errors.New("transaction account not valid for the supplied token")
}

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (f *LambdaFunction) HandleRequest(ctx context.Context, e events.APIGatewayProxyRequest) error {
	reqToken, ok := e.QueryStringParameters["jwt"]
	if !ok {
		return errors.New("no JWT token present in request")
	}

	token, err := jwt.Parse(reqToken, f.validateToken)
	if err != nil {
		return err
	}

	accounts, err := f.accountClaims(token)
	if err != nil {
		return err
	}

	var body []byte
	if e.IsBase64Encoded {
		body, err = base64.StdEncoding.DecodeString(e.Body)
		if err != nil {
			return err
		}
	} else {
		body = []byte(e.Body)
	}

	data, err := monzo.GetEvent(body)
	if err != nil {
		return err
	}

	transaction, ok := data.(*monzo.TransactionCreated)
	if !ok {
		return errors.New("expected a transaction event type")
	}

	err = f.checkIsValidAccount(accounts, transaction)
	if err != nil {
		return err
	}

	b, err := easyjson.Marshal(transaction)
	if err != nil {
		return err
	}

	_, err = f.sns.PublishWithContext(ctx, &sns.PublishInput{
		Message:  aws.String(string(b)),
		Subject:  aws.String(transaction.ID),
		TopicArn: aws.String(f.topic),
	})
	if err != nil {
		return err
	}

	return nil
}

func (f *LambdaFunction) Invoke(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	err := f.HandleRequest(ctx, e)
	if err != nil {
		f.l.Println(err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `"error"`,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `"ok"`,
	}, nil
}

func main() {
	s := session.New()
	f := &LambdaFunction{
		l:      log.New(os.Stderr, "lambda:transaction-hook", log.LstdFlags|log.Lshortfile),
		s:      s,
		sns:    sns.New(s),
		secret: []byte(os.Getenv("LAMBDA_JWT_SECRET")),
		topic:  os.Getenv("LAMBDA_SNS_ARN"),
	}
	lambda.Start(f.Invoke)
}
