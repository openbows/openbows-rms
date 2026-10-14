FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/server /app/server
COPY configs/config.example.yaml /app/configs/config.yaml
COPY migrations /app/migrations
EXPOSE 8080
CMD ["/app/server", "-config", "/app/configs/config.yaml"]
