FROM golang:1.26.0-alpine

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main main.go

EXPOSE 8080

CMD ["./main", "start"]