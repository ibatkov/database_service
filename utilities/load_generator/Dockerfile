FROM golang:1.22 as builder

WORKDIR /app

COPY ../go.mod go.sum ./

RUN go mod download

COPY .. .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o load_generator ./utilities/load_generator

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/load_generator .
