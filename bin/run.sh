#!/usr/bin/env bash

user="root"
remote_path="/root/"
remote_host="47.106.129.170"
port="22"

#VERSION=`git describe --tags`
VERSION=`git rev-parse HEAD`
BUILD_TIME=`date +%FT%T%z`

log()
{
	now=`date "+%Y-%m-%d %H:%M:%S"`
	echo  [INFO] "\033[34;49;1m--------------------------- $now $1 \033[39;49;0m"
}

ssh -t -t $user@$remote_host << EOF
         cd /root/code/vinesai; sh restart.sh; \
         exit
EOF