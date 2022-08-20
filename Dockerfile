# syntax=docker/dockerfile:1

## Build
FROM golang:1.18-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o /trigger-jenkins

## Deploy
FROM alpine:3.16

RUN addgroup trigger && \
    adduser -S -G trigger -h / trigger

USER trigger
WORKDIR /

COPY --from=build /trigger-jenkins /trigger-jenkins

ENTRYPOINT ["/trigger-jenkins"]