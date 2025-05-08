FROM golang:1.24.2
WORKDIR /app
COPY . .  
RUN go mod download
RUN go build -o /app/shortyapi cmd/shorty/main.go
CMD ["/app/shortyapi"]