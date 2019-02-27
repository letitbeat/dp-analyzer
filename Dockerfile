FROM golang:1.9 as builder

RUN apt-get update && apt-get install graphviz -y
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/github.com/letitbeat/dp-analyzer
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only

COPY . ./
RUN go test -v ./...

RUN go build -v main.go

EXPOSE 5000

ENTRYPOINT ["./main"]