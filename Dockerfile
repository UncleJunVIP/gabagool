FROM --platform=linux/arm64 golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    libsdl2-dev \
    libsdl2-ttf-dev

WORKDIR /build

# Copy only go.mod and go.sum first for better caching
COPY go.mod go.sum* ./
# Use go mod download instead of go get
RUN go mod download

# Then copy the rest of the code
COPY . .
RUN go build -v

CMD ["/bin/bash"]