#!/bin/sh
echo "Press enter to use the default value '[value]' or use a custom value"

read -p "Server domain name - this is the end url of where abyss will be hosted []: " -e ABYSS_URL

read -p "Upload key - this is the key you will need to have to be able to make uploads to the server []: " -e UPLOAD_KEY

read -p "Auth username - this is the username to access /tree (show all uploaded files) [admin]: " -e AUTH_USERNAME
if [ -z $AUTH_USERNAME ]; then
    AUTH_USERNAME="admin"
fi

read -p "Auth password - this is the password to access /tree (show all uploaded files) [admin]: " -e AUTH_PASSWORD
if [ -z $AUTH_PASSWORD ]; then
    AUTH_PASSWORD="admin"
fi

cat << EOF > .env
# This is the full name of the final domain for the server. Example: paste.abyss.dev
ABYSS_URL=$ABYSS_URL

# This is the username of the user for accessing /tree
AUTH_USERNAME=$AUTH_USERNAME

# This is the password of the user for accessing /tree
AUTH_PASSWORD=$AUTH_PASSWORD

# This is the key needed to make uploads. Include it as X-Auth in curl.
# Tip: Save it somewhere and use it in curl with \$(cat /path/to/key)
UPLOAD_KEY=$UPLOAD_KEY
EOF
