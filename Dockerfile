FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.23 AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app
COPY . .
RUN go get && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /go/bin/lrp

FROM scratch
COPY --from=builder /go/bin/lrp /go/bin/lrp

# Copy statically compiled curl to enable healthcheck
COPY --from=ghcr.io/tarampampam/curl:8.6.0 /bin/curl /bin/curl

# NOTE: you will need to override the healthcheck if you customize the default
# port or the default prefix in the configuration file. Unfortunately here it's
# not possible to use environment variables because in HEALTHCHECK environment
# variables are replaced by the shell not by the Docker builder, and in this
# container we don't have a shell

# Docs: <https://docs.docker.com/engine/reference/builder/#healthcheck>
HEALTHCHECK --interval=3s --timeout=2s --retries=3 --start-period=1s CMD [ \
    "curl", "--fail", "http://127.0.0.1:8080/healthz" \
]

ENTRYPOINT ["/go/bin/lrp"]
