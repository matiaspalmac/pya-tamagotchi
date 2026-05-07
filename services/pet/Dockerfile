FROM golang:1.23-alpine AS build
WORKDIR /src
COPY services/pet/go.mod services/pet/go.sum* ./
RUN go mod download || true
COPY services/pet ./
RUN CGO_ENABLED=0 go build -o /out/pet ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/pet /app/pet
COPY services/pet/migrations /app/migrations
EXPOSE 8082
ENTRYPOINT ["/app/pet"]
