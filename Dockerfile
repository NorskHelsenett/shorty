FROM golang:1.23-alpine
WORKDIR /app
COPY . .  
RUN go mod download
RUN go build -o /app/shortyapi cmd/shorty/main.go
CMD ["/app/shortyapi"]