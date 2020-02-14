FROM golang:1.12.5-alpine

WORKDIR /app
RUN apk add git

COPY . .
RUN go build -o app

CMD ["/app/app"]
