#!/bin/bash

set -e

export CGO_ENABLED=0

go mod download
go build -o gerrit-event-handler \
    buildkite_webhook_handler.go \
    buildkite.go \
    gerrit_event_handlers.go \
    gerrit_ssh_client.go \
    gerrit.go \
    main.go \
    pipeline.go \
    push_to_remote.go
