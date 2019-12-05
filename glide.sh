#!/bin/bash
set -e

mkdir -p $GOPATH/bin
if [[ -f "/usr/bin/glide" || -f "/usr/local/bin/glide" ]]
then
	echo "glide found"
else
	echo "glide not found in /usr/bin/glide or /usr/local/bin/glide"
        curl https://glide.sh/get | sh
fi
glide install
if [ $? != 0 ];
then
	echo "Glide Install failed"
    exit 1
fi
