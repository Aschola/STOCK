# syntax=docker/dockerfile:1

FROM golang:1.23

WORKDIR /app

COPY . ./

RUN go mod download

RUN go build -o /stock

EXPOSE 80

CMD ["/stock"]