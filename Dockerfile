# syntax=docker/dockerfile:1
FROM golang:1.17-alpine
WORKDIR /app
COPY ./ ./

RUN go mod download

#COPY ./*/*.go .
#RUN ls -a
#RUN go get ./pkg/myhttp

RUN go build -o /myproxy
CMD ["/myproxy"]
