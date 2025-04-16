FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
COPY . .

RUN go mod download

RUN go build -o main ./cmd/main.go

EXPOSE 3000
