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
FROM scratch
COPY --from=builder /app/main /app/
EXPOSE 5000
ENTRYPOINT ["/app/main"]