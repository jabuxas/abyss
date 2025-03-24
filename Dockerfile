FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./

# this is needed because we embed these files into the binary
COPY static/ ./static/
COPY templates/ ./templates

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /abyss

FROM scratch

COPY --from=builder /abyss /abyss

CMD ["/abyss"]
