FROM golang:1.23-alpine AS build
WORKDIR /src
COPY services/gateway/go.mod services/gateway/go.sum* ./
RUN go mod download || true
COPY services/gateway ./
RUN CGO_ENABLED=0 go build -o /out/gateway ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/gateway /app/gateway
EXPOSE 8080
ENTRYPOINT ["/app/gateway"]
