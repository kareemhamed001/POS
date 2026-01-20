FROM golang:1.25.3-alpine as builder

WORKDIR /app

copy . .


RUN GOOS=linux CGO_ENABLED=0 go build -mod=vendor -o main ./cmd/api 

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
copy --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]