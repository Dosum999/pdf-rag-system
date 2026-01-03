FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Ensure go.sum is up to date
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o pdf-rag-server cmd/server/main.go

# Runtime image
FROM alpine:latest

RUN apk --no-cache add ca-certificates python3 py3-pip poppler-utils

WORKDIR /app

# Install Python dependencies for PDF rendering
RUN pip3 install --no-cache-dir pdf2image==1.16.3 Pillow==10.1.0 --break-system-packages

# Copy binary from builder
COPY --from=builder /app/pdf-rag-server .

# Copy PDF renderer script
COPY pdf_renderer.py /app/pdf_renderer.py

# Create uploads directory
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["./pdf-rag-server"]
