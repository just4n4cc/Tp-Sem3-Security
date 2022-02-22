# syntax=docker/dockerfile:1
FROM golang:latest
WORKDIR /app
COPY ./ ./
CMD cp myproxy-ca.crt /usr/local/share/ca-certificates
CMD update-ca-certificates

RUN go mod download
RUN go build -o /myproxy
CMD ["/myproxy"]
