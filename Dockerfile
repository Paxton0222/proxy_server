FROM golang:1.24.5-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod tidy

COPY . .

RUN go build -o app main.go

FROM alpine:3.21

WORKDIR /app

COPY --from=build /app /app/

# 預設執行程式
CMD ["./app"]
