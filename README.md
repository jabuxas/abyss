# abyss
abyss is a basic http server made for uploading files (logs, images) and then sharing them to the internet

note: this is a project made for learning purposes, you should use other more mature projects if running in production

## running:
- edit consts in `main.go` to match your needs. (for example, on server, change `$url` so that the response will be nicely formatted)

- to run it, either build with `go build -o abyss && ./abyss` or run it directly with:
```bash
go run ./main.go
```

- then, simply upload your files with curl:
```bash
curl -X POST -F "file=@/path/to/file" http://localhost:8080/upload # default url:port
```
## todo:
- [x] add upload of logs funcionality (like 0x0.st)
- [ ] add docker easy setup
- [ ] add db for tracking of file names
- [ ] add file browser (like file://)
