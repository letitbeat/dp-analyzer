# build stage
FROM golang:1.12 as builder

RUN apt-get update && apt-get install graphviz -y

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build main.go

# final stage using from scratch to reduce image size
FROM alpine:3.10
RUN apk add --update --no-cache \
           graphviz \
           ttf-freefont \
           python \
           py3-z3 \
           alpine-sdk \
            #python \
            #python-dev \
            py-pip \
            #&& pip install virtualenv \
            && pip install -v z3-solver

COPY --from=builder /app/main /app/
COPY config.yml /app/
COPY templates /app/templates
COPY scripts /app/scripts

EXPOSE 5000
ENTRYPOINT ["/app/main"]