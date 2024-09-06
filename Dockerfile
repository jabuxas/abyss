FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY go.mod ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /abyss

FROM scratch

COPY --from=builder /abyss /abyss

CMD ["/abyss"]
