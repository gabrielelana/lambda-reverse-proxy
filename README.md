## Lambda Reverse Proxy

HTTP reverse proxy for lambda functions. Use lambdas as if they were HTTP
services both locally and on AWS.

### Why

On AWS lambdas are easier to deploy and manage than containers if you don't have
a mature IDP (Internal Developer Platform) at your disposal. Therefore for
something like mock servers or real microservices (with _real_ I mean really
micro, as in their dimension, service) a lambda can be a pragmatic choice.

If you want to call these services implemented with lambdas via HTTP (especially
if they are mocks of services invocable via HTTP) you will have two problems:

1. To run them locally you need to use the [Lambda Runtime
   Interface](https://github.com/aws/aws-lambda-runtime-interface-emulator) to
   beign able to run your lambda code and invoke it via HTTP, but the HTTP
   request must be compatible with the format of the [lambda
   events](https://docs.aws.amazon.com/lambda/latest/api/API_Invoke.html#API_Invoke_RequestSyntax),
   plus the path of your request must have the
   `/2015-03-31/functions/function/invocations` prefix otherwise it will not
   work 🤷
2. To run them on AWS you need an API Gateway, you can have one for a bunch of
   lambdas, but you cannot route requests based on the service domain but only
   on paths, therefore you need allocate a path in the gateway for every lambda
   and use that path in your services as prefix to call those services. On top
   of being awkward to use it costs good money.

### What

This is a reverse proxy able to do the following things:

- Proxy requests based on the domain of the request.
- Convert HTTP requests in HTTP requests in the format expected by the [Lambda
  Runtime Interface](https://github.com/aws/aws-lambda-runtime-interface-emulator) so
  that locally you can start your lambdas and talk to them as they were normal
  HTTP services.
- Convert HTTP requests in lambda invocations so that on AWS, by giving this
  reverse proxy multiple domain aliases, you can use lambdas as normal HTTP
  services.

TLDR: with this reverse proxy you can invoke lambdas via HTTP both locally and
on AWS.

### How

You have a couple of simple HTTP services and implementing them with lambdas
seems a good choice

#### Locally

You want use/test your implementation locally, fortunately you can run lambdas
with Docker using [Lambda Runtime
Interface](https://github.com/aws/aws-lambda-runtime-interface-emulator) to
invoke them as you would call a normal HTTP service you need to use this reverse
proxy.

```yaml
# docker-compose.yaml
services:
  service-1-lambda:
    # ...whatever

  service-2-lambda:
    # ...whatever

  proxy:
    image: gabrielelana/lambda-reverse-proxy:v0.0.2
    entrypoint: /go/bin/lrp /etc/lrp.yaml
    healthcheck:
      test: ["CMD", "curl", "--fail", "http://localhost:8080/healthz"]
      interval: 3s
      timeout: 2s
      retries: 3
      start_period: 1s
    volumes:
      - ./lrp.yaml:/etc/lrp.yaml
    networks:
      default:
        aliases:
          - service-1
          - service-2
    depends_on:
      - service-1-lambda
      - service-2-lambda
```

With the following configuration file

```yaml
# lrp.yaml
local:
  - hostname: service-1
    endpoint: http://service-1-lambda:8080
  - hostname: service-2
    endpoint: http://service-2-lambda:8080
```

With this every request to `http://service-1:8080` will be transalted and
redirected to `service-1-lambda` and all the requests to `http://service-2:8080`
will be transalted and redirected to `service-2-lambda`

All the lambdas listen on port `8080`

By default the proxy listen on port `8080`

### Configuration

```yaml
# Amazon region, useful only when used in AWS
region: string # default: eu-central-1

# Host and port where proxy will listen
host: string # default: 0.0.0.0
port: string # default: 8080

# Prefix of all the internal routes to avoid clashing with routes to proxy to lambdas
internal_route_prefix: string # default: empty, no prefix

# Rules to proxy requests locally
local:
    # Host of the request that should be routed
  - hostname: string
    # Http URL where to route the request
    # NOTE: path prefixes are not supported
    endpoint: string

# Rules to proxy requests on AWS
aws:
    # Host of the request that should be routed
  - hostname: string
    # Lambda function unique name
    function_name: string
```

### Internal Routes

The proxy implements a couple of service routes:
- `/ping` which will always reply with `200 OK` and body `pong`
- `/healthz` to be used as health check route

Both can be customized to avoid clashing with routes implemented by lambdas, by
setting `internal_route_proxy` with somting like `__@@__` in the configuration
file all the internal routes will be prefixed, therefore ping route from `/ping`
will become `/__@@__/ping` and `/healthz` will become `/__@@__/healthz`
