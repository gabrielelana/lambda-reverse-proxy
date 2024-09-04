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

TLDR: with this reverse proxy you can use to invoke lambdas via HTTP both
locally and on AWS.


### How

TODO
