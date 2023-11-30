FROM golang:1.17.2-bullseye as builder

# Create app directory
WORKDIR /app

# A wildcard is used to ensure both package.json AND package-lock.json are copied
COPY . /app

ENV GO111MODULE=on

WORKDIR /app
RUN go build

FROM debian:bullseye-slim

COPY --from=builder /app/go-space-chat /app/go-space-chat
COPY --from=builder /app/web_resource /app/web_resource
COPY --from=builder /app/config /app/config