FROM golang:1.25

RUN apt-get update && apt-get upgrade -y

WORKDIR /app

COPY apps/backend/ ./
COPY packages/go-pkg /packages/go-pkg

RUN go mod tidy
RUN go build -o /web.x ./cmd/web/main.go

CMD ["/web.x"]
