FROM golang:1.23-alpine AS builder

RUN apk update && apk add --no-cache git build-base

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o salesforce-splunk-migration .

FROM alpine:3.20
RUN adduser -D -s /bin/false gouser
WORKDIR /app
COPY --from=builder /app/salesforce-splunk-migration /app/
COPY --from=builder /app/credentials.json /app/
COPY --from=builder /app/resources /app/resources
USER gouser
ENTRYPOINT ["./salesforce-splunk-migration"]