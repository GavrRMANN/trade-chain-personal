FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /trade-chain-back ./cmd/app


FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /trade-chain-back .

EXPOSE 8080

CMD ["./trade-chain-back"]