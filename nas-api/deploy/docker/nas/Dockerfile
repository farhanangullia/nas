## BUILD STAGE
FROM golang:1.19-alpine as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.io,direct

WORKDIR /app/nas

COPY . .

RUN go build -o app ./cmd/nas

## DISTRIBUTION
FROM gcr.io/distroless/base:nonroot-amd64

COPY --from=builder /app/nas/configs/* .
COPY --from=builder /app/nas/app .

CMD ["./app"]
