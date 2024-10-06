# syntax=docker/dockerfile:1
FROM golang:1.23

ENV PKG_CONFIG_PATH=/usr/local/lib/pkgconfig:/usr/lib/pkgconfig
# ENV GOPATH=/go
# ENV PATH=$PATH:/go/bin

WORKDIR /worker
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy

# Install necessary dependencies
RUN apt-get update && \
    apt-get install -y \
        pkg-config \
        cmake \
        ffmpeg \
        libavformat-dev \
        libavcodec-dev \
        libavutil-dev \
        libavfilter-dev \
        libopusfile-dev \
        libswscale-dev \
        libswresample-dev \
        libc6-dev \
        zlib1g-dev \
        libvpx-dev \
        libopus-dev \
        libdrm-dev \
        liblzma-dev \
        libx264-dev \
        build-essential \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Check for correct flags using pkg-config
RUN pkg-config --cflags --libs libavcodec libavformat libavutil vpx
RUN ldd --version


COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /cmd/worker ./cmd/worker/main.go

ENTRYPOINT [ "/cmd/worker" ]
