# Step 1: Modules caching
FROM golang:1.20.6-alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

# Step 2: Tests
FROM golang:1.20.6-alpine
COPY --from=modules /go/pkg /go/pkg
COPY . /app
WORKDIR /app

RUN go env -w CGO_ENABLED=0 && \
    go env -w GOOS=linux && \
    go env -w GOARCH=amd64 && \
    apk add --update-cache make && \
    make


CMD ["./vapp", "-c", "integration/config.yml"]

