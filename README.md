# abyss

abyss is a basic and mostly single user http server written in go made for uploading files (logs, images) and then sharing them to the internet

<figure>
 <img src="https://github.com/user-attachments/assets/eae42368-d8b5-4c42-ac8a-0e1486fcd0d4" alt="homepage"/>
 <figcaption>this is abyss' default home page<figcaption/>
</figure>

## table of contents

- [features](#features)
- [running abyss](#running)
  - [installing with docker](#docker)
  - [installing manually](#manual)
- [uploading files](#uploading)
- [theming](#theming)
- [docs](#docs)
- [todo list](#todo)
- [more pictures](#pictures)

## features

- **file uploads**: supports uploading various file types, including images, videos, and documents.
- **flexible media display**: automatically renders uploaded files on a webpage based on their type (images, pdfs, videos, or plain text).
- **easily customizable interface**: allows for easy modification of color schemes and layout to suit specific design needs.
- **syntax highlighting for code**: syntax highlighting available by default for code files, with support for multiple programming languages. (can be tweaked/changed and even removed)
- **security considerations**: as it is single user, it's mostly secure but there are still some edges to sharpen

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

- you can upload both with the main key and with jwt tokens

##### main key

- to upload your files with main key:

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

  curl -F "file=@$file" -H "X-Auth: $(cat ~/.key)" http://localhost:3235/

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

    curl -F "file=@$file" -H "X-Auth: $(cat ~/.key)" http://localhost:3235/

    if command test -p /dev/stdin
        rm "$file"
    end
end
```

</details>

##### with jwt tokens

- you first need to generate them:

```bash
curl -u admin http://localhost:3235/token # you can also access the url in the browser directly
```

- the user will be the value of `$AUTH_USERNAME` and password the value of `$AUTH_PASSWORD`

- then you use the token in place of the main key:

```bash
curl -F"file=@/path/to/file.jpg" -H "X-Auth: your-token" http://localhost:3235/
```

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
- [x] custom file displaying!!
- [x] syntax highlighting
- [ ] add rate limits

## pictures

<figure>
  <img src="https://github.com/user-attachments/assets/32ce9b3a-8c0f-4bb5-bdcf-3a602e0c81e6"/>
  <figcaption>this is abyss' default directory list<figcaption/>
</figure>

<figure>
  <img src="https://github.com/user-attachments/assets/7072b227-9972-4c2a-a9f3-384d2efb4fe1"/>
  <figcaption>this is abyss' default file presentation<figcaption/>
</figure>
