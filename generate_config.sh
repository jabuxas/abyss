#!/bin/bash
echo "Press enter to use the default value '[value]' or use a custom value"

read -p "Server domain name - this is the end url of where abyss will be hosted []: " -e ABYSS_URL

read -p "Upload key - this is the key you will need to have to be able to make uploads to the server []: " -e UPLOAD_KEY

read -p "Files Directory - this is the dir location for storing the files [./files]: " -e ABYSS_FILEDIR

read -p "Server port - this is the port the server will run on; type just the port number. [3235]: " -e ABYSS_PORT

read -p "Auth username - this is the username to access /tree (show all uploaded files) [admin]: " -e AUTH_USERNAME
if [ -z $AUTH_USERNAME ]; then
    AUTH_USERNAME="admin"
fi

read -p "Auth password - this is the password to access /tree (show all uploaded files) [admin]: " -e AUTH_PASSWORD
if [ -z $AUTH_PASSWORD ]; then
    AUTH_PASSWORD="admin"
fi

read -p "Auth for upload form - should password be needed to upload text through the browser? [yes]: " -e SHOULD_AUTH
if [ -z $SHOULD_AUTH ]; then
    SHOULD_AUTH="yes"
fi

cat << EOF > .env
# This is the full name of the final domain for the server. Example: paste.abyss.dev
ABYSS_URL=$ABYSS_URL

# Where abyss will store files. It's fine to leave it empty. Defaults to "./files"
ABYSS_FILEDIR=$ABYSS_FILEDIR

# The port the server will run on, it's fine to leave it empty. Defaults to :3235
ABYSS_PORT=$ABYSS_PORT

# This is the username of the user for accessing /tree
AUTH_USERNAME=$AUTH_USERNAME

# This is the password of the user for accessing /tree
AUTH_PASSWORD=$AUTH_PASSWORD

# This is whether you need a password to upload text (through browser or curl)
SHOULD_AUTH=$SHOULD_AUTH

# This is the key needed to make uploads. Include it as X-Auth in curl.
# Tip: Save it somewhere and use it in curl with \$(cat /path/to/key)
UPLOAD_KEY=$UPLOAD_KEY
EOF
