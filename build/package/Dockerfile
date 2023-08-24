FROM golang:alpine as builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/cmd/tinyurl
RUN go build -o /tinyurl

FROM alpine:latest as release

WORKDIR /app/

COPY --from=builder /tinyurl .

EXPOSE 1378

ENTRYPOINT ["./tinyurl"]

CMD ["server"]
