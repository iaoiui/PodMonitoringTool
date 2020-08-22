FROM golang:latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./app/main.go .
RUN go build -o main .
CMD ["./main"]