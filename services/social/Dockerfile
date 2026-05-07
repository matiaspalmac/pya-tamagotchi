FROM golang:1.23-alpine AS build
WORKDIR /src
COPY services/social/go.mod services/social/go.sum* ./
RUN go mod download || true
COPY services/social ./
RUN CGO_ENABLED=0 go build -o /out/social ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/social /app/social
COPY services/social/migrations /app/migrations
EXPOSE 8083
ENTRYPOINT ["/app/social"]
