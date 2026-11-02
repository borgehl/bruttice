FROM golang:1.25 AS build
WORKDIR /src
COPY packages/go-pkg ./
RUN go build -o /demo-seed ./examples/demo

FROM debian:stable-slim
COPY --from=build /demo-seed /demo-seed
CMD ["/demo-seed"]
