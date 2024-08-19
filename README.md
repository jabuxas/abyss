# abyss
abyss is a basic http server made for uploading files (logs, images) and then sharing them to the internet

note: this is a project made for learning purposes, you should use other more mature projects if running in production

## table of contents
- [running abyss](#running)
    - [installing with docker](#docker)
    - [installing manually](#manual)
- [todo list](#todo)

## running:
- change URL env variable in to your domain. example: `URL=paste.abyss.dev`
### docker
- to run with docker, you can use either docker compose or just straight docker.
- then run the docker compose command:
```bash
docker compose up -d # might be docker-compose depending on distro
```
- dont change inside port of 8080 unless you know what you're doing

### manual

- to run it, either build with `go build -o abyss` or run it directly with:
```bash
URL="your-domain" go run ./main.go
```

- then, simply upload your files with curl:
```bash
curl -X POST -F "file=@/path/to/file" http://localhost:8080/upload # default url:port
```
## todo:
- [x] add upload of logs funcionality (like 0x0.st)
- [x] add docker easy setup
- [ ] add db for tracking of file names
- [ ] add file browser (like file://)
