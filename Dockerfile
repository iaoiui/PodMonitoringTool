FROM golang:latest as builder
WORKDIR /go
COPY ./vendor/ ./vendor/
COPY go.mod go.sum ./
ENV GOPATH=""
COPY ./app ./app/
RUN go build -o main app/*
CMD ["./main"]


# FROM alpine:3.10
# ENV GOTRACEBACK=single
# # COPY main main
# COPY --from=builder /go/main .
# CMD ["./main"]
