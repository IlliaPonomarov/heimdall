FROM golang:1.25-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /heimdall .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates

COPY --from=build /heimdall /heimdall
COPY --from=build /app/app.yaml /app.yaml

EXPOSE 8080

ENTRYPOINT ["/heimdall"]
