# syntax=docker/dockerfile:1
FROM golang:1.23

WORKDIR /coordinator
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o /cmd/coordinator ./cmd/coordinator/main.go

ENTRYPOINT [ "/cmd/coordinator" ]
