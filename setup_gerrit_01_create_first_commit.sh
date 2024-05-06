#!/bin/bash

set -ex
echo Creating the test-project in Gerrit...
PRIVATE_KEY=$1
PUBLIC_KEY=$2
source .env
GERRIT_ROOT="${GERRIT_ROOT:-gerrit-root}"
GERRIT_PORT="${GERRIT_PORT:-29418}"
GERRIT_WEB_PORT="${GERRIT_WEB_PORT:-8080}"
TEST_PROJECT="${TEST_PROJECT:-test-project}"
# Need an author for a commit
# git config --global user.email "test-project@localhost"
# git config --global user.name "Test user"
ssh -l admin \
    -i $PRIVATE_KEY \
    -p ${GERRIT_PORT} localhost \
     -o UserKnownHostsFile=/dev/null \
     -o StrictHostKeyChecking=no \
    "gerrit create-project --branch main --empty-commit ${TEST_PROJECT}"

[ -d ${TEST_PROJECT} ] && rm -rf ${TEST_PROJECT}
# Add an initial change to the `${TEST_PROJECT}`
export GIT_SSH_COMMAND="ssh -i $(pwd)/$PRIVATE_KEY -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
git clone ssh://admin@localhost:${GERRIT_PORT}/${TEST_PROJECT}.git

pushd ${TEST_PROJECT}
echo Installing commit-msg hook for Gerrit Change-Id information...
f="$(git rev-parse --git-dir)/hooks/commit-msg"

while [ 1 ] ; do 

    curl -o "$f" http://localhost:${GERRIT_WEB_PORT}/tools/hooks/commit-msg
    chmod +x "$f"
    HOOK_SIZE=$(du -s "$f" | cut -f1 )
    # On occasion, the web service is not ready to serve the hook just yet
    if [[ $HOOK_SIZE -gt 0 ]] ; then
        break
    fi

    sleep 0.2
done    


CHANGE=1

echo $CHANGE >> README.md
git add README.md
git commit -m "Change $CHANGE" 
git push origin HEAD:refs/for/main

popd
echo Done
set +x
echo GIT_SSH_COMMAND="ssh -i $PRIVATE_KEY"
echo TEST_PROJECT=${TEST_PROJECT}
echo GERRIT_PORT=${GERRIT_PORT}
echo GERRIT_WEB_PORT=${GERRIT_WEB_PORT}
echo GERRIT_ROOT=${GERRIT_ROOT}