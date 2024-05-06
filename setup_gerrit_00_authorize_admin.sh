#!/bin/bash

set -ex

PRIVATE_KEY=$1
PUBLIC_KEY=$2
source .env
GERRIT_ROOT="${GERRIT_ROOT:-gerrit-root}"
GERRIT_PORT="${GERRIT_PORT:-29418}"
GERRIT_WEB_PORT="${GERRIT_WEB_PORT:-8080}"

# Need an author for a bare repository
# git config --global user.email "test-project@localhost"
# git config --global user.name "Test user"
# Add the Admin public key to the gerrit instance

while [ 1 ] ; do
    if [ -f $GERRIT_ROOT/git/All-Users.git/refs/users/00/1000000 ] ; then
        break
    fi
    echo Waiting for gerrit to start up...
    sleep 3
done

pushd $GERRIT_ROOT/git/All-Users.git
echo Creating authorized keys file
PARENT_TREE=$(git show-ref -s refs/users/00/1000000 ) 
OBJECT_HASH=$(git hash-object -w ../../../${PUBLIC_KEY})
git read-tree -im $PARENT_TREE
git update-index --add --cacheinfo 100644 $OBJECT_HASH authorized_keys
CHILD_TREE=$(git write-tree)
COMMIT=$(git commit-tree -p $PARENT_TREE -m "Add ssh key" $CHILD_TREE )
echo $COMMIT > refs/users/00/1000000 

popd
set +e
while [ 1 ] ; do
    ssh -l admin -i $PRIVATE_KEY -p $GERRIT_PORT  -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no localhost 'gerrit version'
    if [ $? == 0 ] ; then
        break
    fi
    echo Waiting for gerrit ssh reciever to launch...
    sleep 5
done
