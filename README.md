# abyss

abyss is a basic (mostly) single user http server made for uploading files (logs, images) and then sharing them to the internet

note: this is a project made for learning purposes, you should use other more mature projects if running in production. probably.

## table of contents

- [running abyss](#running)
  - [installing with docker](#docker)
  - [installing manually](#manual)
- [uploading files](#uploading)
- [theming](#theming)
- [docs](#docs)
- [todo list](#todo)

## running:

- run `./generate_config.sh` to setup the necessary environment variables

### docker

- to run with docker, you can use the `docker-compose.yml` file available in this repo. to do so, run:

```bash
docker compose up -d # might be docker-compose depending on distro
```

- you can optionally use the [docker image](https://git.jabuxas.xyz/jabuxas/-/packages/container/abyss/latest) directly and set it up how you want

### manual

- to run it manually, build it with `go build -o abyss` and run:

```bash
./abyss
```

## uploading

#### with curl

- to upload your files with curl:

```bash
curl -F "file=@/path/to/file" -H "X-Auth: "$(cat /path/to/.key) http://localhost:3235/
```

- you should probably create an `alias` or a `function` to do this automatically for you.
<details>
<summary>click for an example for bash/zsh:</summary>

```bash
pst() {
  local file

  if [[ -p /dev/stdin ]]; then
    file=$(mktemp)
    cat > "$file"
  elif [[ -n $1 ]]; then
    file="$1"
  else
    echo "Usage: pst [file]"
    return 1
  fi

  curl -F "file=@$file" -H "X-Auth: $(cat ~/.key)" http://localhost:3235

  if [[ -p /dev/stdin ]]; then
    rm "$file"
  fi
}
```

</details>

<details>
<summary>click for an example for fish shell:</summary>

```bash
function pst
    set -l file

    if command test -p /dev/stdin
        set file "/tmp/tmp.txt"
        cat > $file
    else if test -n "$argv[1]"
        set file "$argv[1]"
    end

    curl -F "file=@$file" -H "X-Auth: $(cat ~/.key)" http://localhost:3235

    if command test -p /dev/stdin
        rm "$file"
    end
end
```

</details>

#### through the browser

- you can only upload text through the browser, to do so, simply write text in the form in the default webpage and click upload.
- this upload can be restricted to need authentication or not, controlled by an environment variable.

## theming

- there is an example homepage in `static/` you can edit directly, which the server will serve automatically
- if running with docker, it's also possible to override `/static` inside the container with your own page.
  - otherwise you will need to clone this repository and edit `static/` and `templates/` manually, or recreate the structure.
- same thing with templates in `templates/`
  - it is preferred to use `dev/` for that reason, since it is git-ignored and that way makes it easier if wanting to update regularly without making changes to the tree

## docs

- `ABYSS_URL`: this is used for the correct formatting of the response of `curl`.
- `AUTH_USERNAME | AUTH_PASSWORD`: this is used to access `/tree`, which shows all uploaded files
- `UPLOAD_KEY`: this is key checked when uploading files. if the key doesn't match with server's one, then it refuses uploading.
- `ABYSS_FILEDIR`: this points to the directory where abyss will save the uploads to. defaults to `./files`
- `ABYSS_PORT`: this is the port the server will run on. safe to leave empty. defaults to 3235
- `SHOULD_AUTH`: if it is `yes`, then to upload text through the browser you will need authentication (same auth as `/tree`), any value other than that and upload is auth-less

## todo:

- [x] add upload of logs funcionality (like 0x0.st)
- [x] add docker easy setup
- ~~add db for tracking of file names~~ (dont need that)
- [x] add file browser (like file://)
- [x] add file extension in its name
- [x] login prompt when accessing /tree
- [x] home page
- [ ] add rate limits
