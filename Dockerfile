FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app
COPY . .
RUN go get
RUN go build -o /go/bin/lrp

FROM scratch
COPY --from=builder /go/bin/lrp /go/bin/lrp
ENTRYPOINT ["/go/bin/lrp"]
