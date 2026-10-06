FROM golang:alpine

WORKDIR /app

# Install dependensi
COPY go.mod go.sum ./
RUN go mod download

# Salin kode dan build
COPY . .
RUN go build -o server main.go

# Buka port 8080
EXPOSE 8080

# Jalankan server
CMD ["./server"]
