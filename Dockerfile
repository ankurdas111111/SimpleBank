#converting this docker file to multistage
#Build stage
From golang:1.25.0-alpine3.22 AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go
RUN apk add curl
RUN curl -Ls https://github.com/golang-migrate/migrate/releases/download/v4.18.3/migrate.linux-amd64.tar.gz | tar -xz

# Run stage
FROM alpine:3.13
WORKDIR /app
RUN apk add --no-cache netcat-openbsd
COPY --from=builder /app/main .
COPY --from=builder /app/migrate ./migrate
COPY app.docker.env ./app.env
COPY start.sh ./start.sh
COPY wait-for.sh ./wait-for.sh
COPY db/migration ./migration


EXPOSE 8080
CMD ["/app/main"] 
ENTRYPOINT ["/app/start.sh"]