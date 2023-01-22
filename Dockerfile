FROM golang:1.19-alpine as build

WORKDIR /app

COPY . .
RUN go build -o /news-feed .

ENTRYPOINT ["/news-feed"]

FROM alpine:latest

WORKDIR /app

COPY --from=build /news-feed /news-feed

EXPOSE 80

ENTRYPOINT ["/news-feed"]
