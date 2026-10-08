FROM golang:1.26-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /app .
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=build /app ./app
COPY config.json ./config.json
ENTRYPOINT ["./app"]
