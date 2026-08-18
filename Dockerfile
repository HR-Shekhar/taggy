# Builds the Go API from the monorepo root (Railway default).
FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY taggy-backend/go.mod taggy-backend/go.sum ./
RUN go mod download

COPY taggy-backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /server /server

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/server"]
