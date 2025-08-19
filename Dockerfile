FROM golang:1.25-alpine AS builder

WORKDIR /app

ARG APP_VERSION="dev-docker"
ARG GIT_COMMIT="none-docker"
ARG BUILD_DATE="unknown-docker"
ARG BUILT_BY="docker"

COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the entire source code and assets

COPY . .

RUN echo "building docker image with version: ${APP_VERSION}, Commit: ${GIT_COMMIT}, Date: ${BUILD_DATE}, BuiltBy: ${BUILT_BY}" && \
    export CGO_ENABLED=0 && \
    export GOOS=linux && \
    export GOARCH=amd64 && \
    LDFLAGS="-s -w -extldflags '-static'" && \
    LDFLAGS="$LDFLAGS -X main.version=${APP_VERSION}" && \
    LDFLAGS="$LDFLAGS -X main.commit=${GIT_COMMIT}" && \
    LDFLAGS="$LDFLAGS -X main.date=${BUILD_DATE}" && \
    LDFLAGS="$LDFLAGS -X main.builtBy=${BUILT_BY}" && \
    go build -v -a \
      -ldflags="$LDFLAGS" \
      -o /app/abyss \
      ./cmd/abyss

FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/abyss .

EXPOSE 3235

CMD ["./abyss"]
