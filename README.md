# abyss

abyss is a basic single user http server made for uploading files (logs, images) and then sharing them to the internet

note: this is a project made for learning purposes, you should use other more mature projects if running in production

## table of contents

- [running abyss](#running)
  - [installing with docker](#docker)
  - [installing manually](#manual)
- [uploading files](#uploading)
- [todo list](#todo)

## running:

- change URL environment variable to your end url. example: `URL=paste.abyss.dev` if you your files will be accessed through `paste.abyss.dev/name-of-file`
- add your password (key) to `.key` in the root directory of the project - it will be used for authentication for uploads.
- add AUTH_USERNAME and AUTH_PASSWORD environment variables for access to `/tree/`

### docker

- to run with docker, you can use docker compose:

```bash
docker compose up -d # might be docker-compose depending on distro
```

- dont change inside port of 8080 unless you know what you're doing
- when updating, run with `--build` instead:

```bash
docker compose up --build -d
```

### manual

- to run it, either build with `go build -o abyss` or run it directly with:

```bash
URL="your-domain" AUTH_USERNAME=admin AUTH_PASSWORD=admin go run ./main.go
```

## uploading

- then, simply upload your files with curl:

```bash
curl -F "file=@/path/to/file" -H "X-Auth: "$(cat /path/to/.key) http://localhost:8080/
```

## todo:

- [x] add upload of logs funcionality (like 0x0.st)
- [x] add docker easy setup
- ~~add db for tracking of file names~~ (dont need that)
- [x] add file browser (like file://)
- [x] add file extension in its name
- [x] login prompt when accessing /tree
- [ ] add rate limits
