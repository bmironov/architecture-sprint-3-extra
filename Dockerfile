FROM golang:1.23-alpine3.20
WORKDIR /app
COPY src/go.mod src/go.sum ./
COPY src/main/*.go  ./main/
COPY src/hvac/*.go  ./hvac/
COPY src/light/*.go ./light/
RUN go mod download && cd main; CGO_ENABLED=0 GOOS=linux go build -o /warm_home ./... && ls -r /

EXPOSE 8080/tcp

# Run
ENTRYPOINT ["/warm_home"]
