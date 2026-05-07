FROM golang:1.23-alpine AS build
WORKDIR /src
COPY services/notif/go.mod services/notif/go.sum* ./
RUN go mod download || true
COPY services/notif ./
RUN CGO_ENABLED=0 go build -o /out/notif ./cmd

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/notif /app/notif
EXPOSE 8084
ENTRYPOINT ["/app/notif"]
