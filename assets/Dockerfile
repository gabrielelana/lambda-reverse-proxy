FROM node:22-alpine

# Go back to autoconf=2.71-r2; the upgrade to 2.72-r0 breaks the awslambdaric install
RUN echo https://dl-cdn.alpinelinux.org/alpine/v3.19/main >> /etc/apk/repositories
RUN echo https://dl-cdn.alpinelinux.org/alpine/v3.19/community >> /etc/apk/repositories
RUN apk add autoconf=2.71-r2 curl automake bash binutils cmake g++ gcc libtool make lsof openssl-dev unzip python3 py3-setuptools

# it has to be 3.16, too, after that it was removed :'(
RUN apk add --no-cache --update --repository=https://dl-cdn.alpinelinux.org/alpine/v3.16/main/ libexecinfo-dev

WORKDIR function

# Install aws-lambda-rie
RUN curl -Lo aws-lambda-rie https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie && \
    chmod +x aws-lambda-rie

RUN npm install -g pnpm
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
COPY package.json pnpm-lock.yaml .
RUN pnpm install --frozen-lockfile

# Install aws-lambda-ric
RUN pnpm add -g aws-lambda-ric

COPY . .

# Run the lambda with NodeJS
CMD ./aws-lambda-rie pnpm aws-lambda-ric "lambda.handler" || exit 1
