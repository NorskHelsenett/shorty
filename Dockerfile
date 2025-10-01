FROM golang:1.25-alpine
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /app/shortyapi cmd/shorty/main.go
CMD ["/app/shortyapi"]