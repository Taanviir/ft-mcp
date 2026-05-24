FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ft-mcp .

FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/ft-mcp /ft-mcp
CMD ["/ft-mcp", "--transport", "http"]
