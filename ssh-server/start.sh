#!/bin/bash


function signal_handler {
    echo "Caught SIGINT signal! $1"
    exit 0
}

trap signal_handler SIGINT
git init --bare /git-root
chown -R git:git /git-root
chown -R git:git /git

ls -la /git/
ls -la /git/.ssh

/etc/init.d/ssh start
sleep 1
while [ 1 ]; do
    tail -f /var/log/*log
done