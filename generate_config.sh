#!/bin/bash

# script exits if any command fails
set -e

echo "--- abyss configuration setup ---"
echo "press Enter to use the default value shown in brackets."
echo

> .env

read -p "Server URL (e.g., files.example.com) []: " -e ABYSS_URL
if [[ -n "$ABYSS_URL" ]]; then
    echo "# The full URL where the application is accessible. Used for generating response URLs." >> .env
    echo "ABYSS_URL=${ABYSS_URL}" >> .env
    echo "" >> .env
fi

# optional. 
read -p "Files directory [./files]: " -e ABYSS_FILEDIR
if [[ -n "$ABYSS_FILEDIR" ]]; then
    echo "# The local directory where Abyss will store uploaded files." >> .env
    echo "ABYSS_FILEDIR=${ABYSS_FILEDIR}" >> .env
    echo "" >> .env
fi

# optional. 
read -p "Server port [3235]: " -e ABYSS_PORT
if [[ -n "$ABYSS_PORT" ]]; then
    echo "# The port the server will listen on." >> .env
    echo "ABYSS_PORT=${ABYSS_PORT}" >> .env
    echo "" >> .env
fi

read -p "Basic-Auth username [admin]: " -e AUTH_USERNAME
read -p "Basic-Auth password [changeme]: " -e AUTH_PASSWORD

: "${AUTH_USERNAME:=admin}"
: "${AUTH_PASSWORD:=changeme}"

echo "# credentials for accessing protected endpoints like /all." >> .env
echo "AUTH_USERNAME=${AUTH_USERNAME}" >> .env
echo "AUTH_PASSWORD=${AUTH_PASSWORD}" >> .env
echo "" >> .env


while true; do
    read -p "Upload key (required for uploads): " -e UPLOAD_KEY
    if [[ -n "$UPLOAD_KEY" ]]; then
        echo "# the secret key that must be sent in the 'X-Auth' header for uploads." >> .env
        echo "UPLOAD_KEY=${UPLOAD_KEY}" >> .env
        break
    else
        echo "the upload key cannot be empty. please enter a value."
    fi
done

echo
echo "✅ .env file has been generated successfully!"
