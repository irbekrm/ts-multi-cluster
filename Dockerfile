FROM golang AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o region-server main.go

FROM alpine:latest
COPY --from=builder /app/region-server .
EXPOSE 8080

CMD ["./region-server"]