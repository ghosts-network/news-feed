FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o /news-feed ./app/api

EXPOSE 3001

CMD ["/news-feed"]
