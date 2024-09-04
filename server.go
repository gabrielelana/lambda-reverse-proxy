package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
)

var ErrRouteNotFound = errors.New("Route not found")

type Config struct {
	Region              string          `yaml:"region" default:"eu-central-1"`
	Host                string          `yaml:"host" default:"0.0.0.0"`
	Port                string          `yaml:"port" default:"8080"`
	InternalRoutePrefix string          `yaml:"internal_route_prefix"`
	Local               []ProxyToHost   `yaml:"local"`
	AWS                 []ProxyToLambda `yaml:"aws"`
}

type ProxyToLambda struct {
	Hostname     string `yaml:"hostname"`
	FunctionName string `yaml:"function_name"`
}

type ProxyToHost struct {
	Hostname string `yaml:"hostname"`
	Endpoint string `yaml:"endpoint"`
}

type Destination interface {
	Invoke(payload []byte) (*lambda.InvokeOutput, error)
}

type LocalDestination struct {
	mu       *sync.Mutex
	Endpoint string
	Region   string
}

type LambdaDestination struct {
	FunctionName string
	Region       string
}

func (d LocalDestination) Invoke(payload []byte) (*lambda.InvokeOutput, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	s, err := session.NewSession(&aws.Config{
		Endpoint:    aws.String(d.Endpoint),
		Region:      aws.String(d.Region),
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		return nil, err
	}
	l := lambda.New(s)
	return l.Invoke(&lambda.InvokeInput{
		// TODO: someting related to observability?
		// ClientContext:  new(string),
		FunctionName:   aws.String("function"),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        payload,
	})
}

func (d LambdaDestination) Invoke(payload []byte) (*lambda.InvokeOutput, error) {
	s, err := session.NewSession(&aws.Config{
		Region:      aws.String(d.Region),
		Credentials: credentials.NewEnvCredentials(),
	})
	if err != nil {
		return nil, err
	}
	l := lambda.New(s)
	return l.Invoke(&lambda.InvokeInput{
		// TODO: someting related to observability?
		// ClientContext:  new(string),
		FunctionName:   aws.String(d.FunctionName),
		InvocationType: aws.String(lambda.InvocationTypeRequestResponse),
		Payload:        payload,
	})
}

func NewServer(config Config) http.Handler {
	mux := http.NewServeMux()

	prefix := ""
	if len(config.InternalRoutePrefix) > 0 {
		prefix = "/" + config.InternalRoutePrefix
	}

	mux.HandleFunc(
		fmt.Sprintf("GET %s/ping", prefix),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("pong"))
		})

	mux.HandleFunc(
		fmt.Sprintf("GET %s/healthz", prefix),
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(""))
		})

	var mu sync.Mutex
	mux.HandleFunc("/", Rie(&config, &mu))

	return mux
}

func Route(config *Config, hostname string, mu *sync.Mutex) (Destination, error) {
	for _, host := range config.Local {
		if host.Hostname == hostname {
			return LocalDestination{
				mu:       mu,
				Endpoint: host.Endpoint,
				Region:   config.Region,
			}, nil
		}
	}
	for _, host := range config.AWS {
		if host.Hostname == hostname {
			return LambdaDestination{
				FunctionName: host.FunctionName,
				Region:       config.Region,
			}, nil
		}
	}
	return nil, fmt.Errorf("Cannot route hostname %s: %w", hostname, ErrRouteNotFound)
}

func Rie(config *Config, mu *sync.Mutex) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		event, err := ToEvent(r)
		if err != nil {
			// TODO: better error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Cannot convert HTTP request to lambda event"))
			return
		}

		payload, err := json.Marshal(event)
		if err != nil {
			// TODO: better error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Cannot encode lambda event to JSON payload"))
			return
		}

		d, err := Route(config, event.RequestContext.DomainName, mu)
		if err != nil {
			if errors.Is(err, ErrRouteNotFound) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(err.Error()))
				return
			}
			// TODO: better error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Unable to create an AWS SDK session"))
			return
		}

		o, err := d.Invoke(payload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Lambda invocation error: " + err.Error()))
			return
		}

		if len(o.Payload) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Missing payload in lambda response"))
			return
		}

		if o.FunctionError != nil && len(*o.FunctionError) > 0 {
			// TODO: better error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Lambda invocation error"))
			return
		}

		response := events.APIGatewayV2HTTPResponse{}
		if err := json.Unmarshal(o.Payload, &response); err != nil {
			// TODO: better error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Invalid lambda response"))
			return
		}

		FromEvent(&response, w)
	}
}

func FromEvent(event *events.APIGatewayV2HTTPResponse, w http.ResponseWriter) {
	for key, value := range event.Headers {
		w.Header().Set(key, value)
	}
	for _, cookie := range event.Cookies {
		w.Header().Set("Set-Cookie", cookie)
	}
	w.WriteHeader(event.StatusCode)
	w.Write([]byte(event.Body))
}

func ConvertCookies(cookies []*http.Cookie) []string {
	var res []string
	for _, cookie := range cookies {
		res = append(res, strings.Join([]string{cookie.Name, cookie.Value}, "="))
	}
	return res
}

func ConvertHeaders(headers map[string][]string) map[string]string {
	res := make(map[string]string, len(headers))
	for name, values := range headers {
		res[name] = strings.Join(values, ",")
	}
	return res
}

func ConvertQueryString(qs string) (map[string]string, error) {
	res := make(map[string]string)
	elements, err := url.ParseQuery(qs)
	if err != nil {
		return res, err
	}
	for name, values := range elements {
		res[name] = strings.Join(values, ",")
	}
	return res, nil
}

func ToEvent(r *http.Request) (*events.APIGatewayV2HTTPRequest, error) {
	cookies := ConvertCookies(r.Cookies())
	headers := ConvertHeaders(r.Header)
	queryStringParameters, err := ConvertQueryString(r.URL.RawQuery)
	if err != nil {
		return nil, err
	}
	userAgent := r.Header.Get("User-Agent")

	domainName := r.Host
	if tokens := strings.Split(domainName, ":"); len(tokens) > 1 {
		domainName = tokens[0]
	}

	domainPrefix := ""
	domainComponents := strings.Split(domainName, ".")
	if len(domainComponents) > 2 {
		domainPrefix = domainComponents[0]
	}

	bufferBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	body := string(bufferBody)

	shouldBeBase64Encoded := strings.Contains(body, "Content-Disposition: form-data")
	if shouldBeBase64Encoded {
		body = base64.StdEncoding.EncodeToString(bufferBody)
	}

	lambdaPayload := events.APIGatewayV2HTTPRequest{
		Version:               "2.0",
		RouteKey:              "$default",
		RawPath:               r.URL.Path,
		RawQueryString:        r.URL.RawQuery,
		Cookies:               cookies,
		Headers:               headers,
		QueryStringParameters: queryStringParameters,
		PathParameters:        map[string]string{},
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			RouteKey:  "$default",
			AccountID: "123456789012",
			Stage:     "$default",
			// TODO: observability
			RequestID:    "",
			Authorizer:   &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{},
			APIID:        "api-id",
			DomainName:   domainName,
			DomainPrefix: domainPrefix,
			Time:         time.Now().Format(time.RFC3339),
			TimeEpoch:    time.Now().Unix(),
			HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{
				Method:    r.Method,
				Path:      r.URL.Path,
				Protocol:  r.Proto,
				SourceIP:  r.RemoteAddr,
				UserAgent: userAgent,
			},
		},
		StageVariables:  map[string]string{},
		Body:            body,
		IsBase64Encoded: shouldBeBase64Encoded,
	}

	return &lambdaPayload, nil
}
