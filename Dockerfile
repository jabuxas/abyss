FROM golang:1.25-alpine AS builder

WORKDIR /app

ARG APP_VERSION="dev"
ARG GIT_COMMIT
ARG BUILD_DATE
ARG BUILT_BY

RUN test -n "${BUILD_DATE}" || BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") && \
    echo "Build date set to ${BUILD_DATE}"

RUN test -n "${GIT_COMMIT}" || GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
    echo "Git commit set to ${GIT_COMMIT}"

RUN test -n "${BUILT_BY}" || BUILT_BY=$(whoami 2>/dev/null || echo "docker") && \
    echo "Built by ${BUILT_BY}"

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN echo "building docker image with version: ${APP_VERSION}, commit: ${GIT_COMMIT}, date: ${BUILD_DATE}, builtBy: ${BUILT_BY}" && \
    export CGO_ENABLED=0 && \
    export GOOS=linux && \
    export GOARCH=amd64 && \
    go build -v -a \
      -ldflags="-s -w \
      -X main.version=${APP_VERSION} \
      -X main.commit=${GIT_COMMIT} \
      -X main.date=${BUILD_DATE} \
      -X main.builtBy=${BUILT_BY}" \
      -o /app/abyss \
      ./cmd/abyss

FROM alpine:latest

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/abyss .
COPY --from=builder /app/assets ./assets

ENV GIN_MODE=release

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 3235

CMD ["./abyss"]
