# Start from the official Go image
FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

# Command to run the application
CMD ["./main"]