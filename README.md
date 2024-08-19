# abyss
abyss is a basic http server made for uploading images and then sharing them to the internet

note: this is a project made for learning purposes, you should use other more mature projects if running in production

## running:
- edit consts in `main.go` to match your needs. (for example, on my server, change `$url` so that the response will be nicely formatted)

- to run it, either build with `go build -o abyss` or run it directly with:
```
go run ./main.go
```
## todo:
- add upload of logs funcionality (like 0x0.st)
- add docker easy setup
