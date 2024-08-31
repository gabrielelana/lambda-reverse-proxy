package main

import (
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	mux.HandleFunc("/", Rie)

	return mux
}

func Rie(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Host: %s\n", r.Host)

	s, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String("http://lambda:8080"),
		Region:      aws.String("eu-central-1"),
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to create an AWS SDK session"))
		return
	}

	l := lambda.New(s)
	o, err := l.Invoke(&lambda.InvokeInput{
		// TODO: someting related to observability?
		// ClientContext:  new(string),
		FunctionName:   aws.String("function"),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        []byte("{}"),
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Unable to invoke AWS lambda"))
		return
	}
	fmt.Println(string(o.Payload))
	w.WriteHeader(http.StatusOK)
}
