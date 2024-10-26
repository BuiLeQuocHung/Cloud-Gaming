# syntax=docker/dockerfile:1
FROM golang:1.23

WORKDIR /worker
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y \
        pkg-config \
        libavcodec-dev \
        libavutil-dev \
        libswscale-dev \
        libc6-dev \
        libvpx-dev \
        libopus-dev \
        libopusfile-dev \
        libx264-dev \
        build-essential \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /cmd/worker ./cmd/worker/main.go

ENTRYPOINT [ "/cmd/worker" ]
