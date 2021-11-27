FROM golang:1.17.1
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD go run main.go --start-service