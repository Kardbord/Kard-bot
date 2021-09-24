# syntax=docker/dockerfile:1

FROM golang:1.17-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
COPY main.go ./
COPY kardbot/ ./kardbot
COPY pasta/ ./pasta

RUN go mod download
RUN go build -o ./Kard-bot

CMD [ "./Kard-bot" ]
