FROM golang:1.25.6-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd

FROM gcr.io/distroless/static-debian11 AS build-release-stage

WORKDIR /

COPY --from=build-stage /api /api

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/api"]
