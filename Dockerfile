FROM golang:1.19-alpine

WORKDIR /app

COPY . .
RUN go build -o /news-feed .

EXPOSE 80

ENTRYPOINT ["/news-feed"]
