# مرحلة البناء
FROM golang:1.25-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/api

# مرحلة التشغيل — صورة نهائية صغيرة بدون أدوات البناء
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /server /app/server
EXPOSE 4000
ENTRYPOINT ["/app/server"]
