FROM golang:1.26-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN go build -ldflags "-X main.version=${VERSION}" -o referee-dashboard ./cmd

FROM scratch
COPY --from=builder /build/referee-dashboard /referee-dashboard
COPY --from=builder /build/db/migrations /db/migrations
COPY --from=builder /build/static /static
EXPOSE 3000
CMD ["/referee-dashboard", "serve"]
