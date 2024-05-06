The Gerrit code review and jGit implementation writes line-oriented JSON events to an [SSH event stream](https://gerrit-review.googlesource.com/Documentation/cmd-stream-events.html) when Changes are created or updated `patchset-created`, when they are merged `change-merged`, or receive comments `comment-added`. There are other [events too](https://gerrit-review.googlesource.com/Documentation/cmd-stream-events.html#_schema).

# Building

`./build.sh` will produce `gerrit-event-handler` for use. See the **Features** section below for usage examples.

# Features

## Should Listen for Gerrit Events

Given we want to listen to Gerrit Event Stream events over SSH
Then it should be enabled by the `--stream-type=ssh` configuration
And `--gerrit-ssh-url` should point to the Gerrit SSH url
And `--gerrit-ssh-key-path` should accept the Gerrit SSH identity key path

```
gerrit-event-handler \
    --stream-type=ssh \
    --gerrit-ssh-url 'ssh://user@gerrit:29418/my-project' \
    --gerrit-ssh-key-path key-in-current-directory
```

## Should Create BuildKite Build Jobs

**Experimental** This hasn't been tested against a BuildKite with a real API token

Given we want to create BuildKite Builds of Changes and Patches from Gerrit
And we want to trigger builds each time a Chanage is created/updated
And we want to configure the BuildKite Organization
And we want to configure the BuildKite Pipeline
And we want to configure the BuildKite API Token
And we want to configure the Buildkite API Url
Then it should be a default behavior
And `--buildkite-org-slug` configures the BuildKite Organization
And `--buildkite-pipeline-slug` configures the BuildKite Pipeline
And `--buildkite-api-token-path` configures the file containing the buildkite api token
And `--buildkite-api-url` configures the BuildKite api url

```
gerrit-event-handler \
    --buildkite-org-slug my-org \
    --buildkite-pipeline-slug my-pipeline \
    --buildkite-api-token-path file-with-api-token \
    --buildkite-api-url https://real-or-fake-api-url \
```

Given we create BuildKite builds
And we want to relate them to Gerrit Change-patches
And we should be able to disable it
Then there should be a BuildKite Webhook handler
And `--disable-buildkite-webhook-handler` disables it

```
gerrit-event-handler \
    --disable-buildkite-webhook-handler
```

## Should Replicate Changes to SSH Remotes

Given we want to replicate Gerrit Changes
And they should replicate when Changes are created
And should replicate to the same location when Changes are updated
Then `--enable-change-replication` should adds this handler on `patchset-created` events
And use `--replication-clone-path` for staging the Gerrit Change before replication
And use `--replication-destination-url` for the repository to send changes to
And use `--replication-ssh-key-path` as the destination repository ssh identity


# Development Notes

# Buildkite Gerrit bridge

Buildkite doesn't support gerrit integration out of the box but provides all
the hooks to implement it pretty easily.

The basic flow is as follow:
 1) gerrit-buildkite uses gerrit stream-events over ssh to watch for events
 2) When an event which should trigger a verification is found, a build is started in Buildkite using the REST API.
 3) The gerrit event and the newly created Buildkite build event are tracked together in a map so the results can be correlated back to the review.
 4) gerrit-buildkite runs a small webserver which listens for the webhooks back from Buildkite
 5) When a response comes back, we look in the map, and if there is an associated review, we publish the results back to gerrit.

If you reply to a review in gerrit with 'retest' on a line, it will re-trigger a verification.
Change-


# Version 2

![basic design](https://github.com/mrmod/gerrit-buildkite/blob/version-22/Design.png?raw=true)

* It'd be nice to allow different backends to be used.
* What if Gerrit events could be accepted on different topologies like SSH, Kineses, or webhooks?
* Logging could be better.
* Could Gerrit event dispatch be driven by YAML and text templates?


## Running in Development

```
# Start Redis and Gerrit
docker-compose up -d

go run *.go \
    --buildkite-api-token-path ./local.buildkite_api_token \
    --buildkite-api-url "http://localhost:9999" \
    --disable-buildkite-webhook-handler=false \
    --gerrit-ssh-key-path ./local.private.key \
    --gerrit-ssh-url "ssh://admin@localhost:29418/test-project" \
    --enable-debug-logging \
    --stream-type ssh \
     2>&1 | tee events.log
```