FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o migrate ./cmd/migrate/main.go
RUN go build -o initialization ./cmd/initialization/main.go
RUN go build -o core ./cmd/core/main.go

EXPOSE 3000