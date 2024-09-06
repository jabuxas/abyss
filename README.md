# abyss

abyss is a basic single user http server made for uploading files (logs, images) and then sharing them to the internet

note: this is a project made for learning purposes, you should use other more mature projects if running in production. probably.

## table of contents

- [running abyss](#running)
  - [installing with docker](#docker)
  - [installing manually](#manual)
- [uploading files](#uploading)
- [docs](#docs)
- [todo list](#todo)

## running:

- run `./generate_config.sh` to setup the necessary environment variables

### docker

- to run with docker, you can use docker compose:

```bash
docker compose up -d # might be docker-compose depending on distro
```

- you can optionally use the [docker image](https://git.jabuxas.xyz/jabuxas/-/packages/container/abyss/latest) directly and set it up how you want

- dont change inside port of 8999 unless you know what you're doing

### manual

- to run it manually, build it with `go build -o abyss` and run:

```bash
./abyss
```

## uploading

- then, simply upload your files with curl:

```bash
curl -F "file=@/path/to/file" -H "X-Auth: "$(cat /path/to/.key) http://localhost:8999/
```

## docs

- `ABYSS_URL`: this is used for the correct formatting of the response of `curl`.
- `AUTH_USERNAME | AUTH_PASSWORD`: this is used to access `http://localhost:8999/tree`, which shows all uploaded files
- `UPLOAD_KEY`: this is key checked when uploading files. if the key doesn't match with server's one, then it refuses uploading.

## todo:

- [x] add upload of logs funcionality (like 0x0.st)
- [x] add docker easy setup
- ~~add db for tracking of file names~~ (dont need that)
- [x] add file browser (like file://)
- [x] add file extension in its name
- [x] login prompt when accessing /tree
- [ ] add rate limits
