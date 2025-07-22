FROM golang:1.24.5-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && \
    go mod tidy

COPY . .

RUN go build -o app main.go

# 預設執行程式
CMD ["./app"]
