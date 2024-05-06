#!/bin/bash

PRIVATE_KEY="gerrit-local-ssh-keypair-$RANDOM"
PUBLIC_KEY="$PRIVATE_KEY.pub"
ssh-keygen -t ed25519 -f $PRIVATE_KEY -N ''

./setup_gerrit_00_authorize_admin.sh $PRIVATE_KEY $PUBLIC_KEY
./setup_gerrit_01_create_first_commit.sh $PRIVATE_KEY $PUBLIC_KEY