#converting this docker file to multistage
#Build stage
From golang:1.25.0-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

# Run stage
FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .
COPY app.docker.env ./app.env

EXPOSE 8080
CMD ["/app/main"] 