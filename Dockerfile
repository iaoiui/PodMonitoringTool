FROM golang:alpine as builder
WORKDIR /go
COPY ./vendor/ ./vendor/
COPY go.mod go.sum ./
ENV GOPATH=""
COPY ./app ./app/
RUN GOOS=linux GOARCH=amd64 go build -o main app/*

FROM alpine:latest
ENV GOTRACEBACK=single

COPY --from=builder /go/main .
CMD ["./main"]