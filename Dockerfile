# syntax=docker/dockerfile:1

FROM golang:1.23

WORKDIR /app

COPY . ./

RUN go mod download

RUN go build -o /sms-service

EXPOSE 8080

CMD ["/stock"]