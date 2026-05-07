FROM golang:1.23-alpine AS build
WORKDIR /src
COPY services/auth/go.mod services/auth/go.sum* ./
RUN go mod download || true
COPY services/auth ./
RUN CGO_ENABLED=0 go build -o /out/auth ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/auth /app/auth
COPY services/auth/migrations /app/migrations
EXPOSE 8081
ENTRYPOINT ["/app/auth"]
