The Gerrit code review and jGit implementation writes line-oriented JSON events to an [SSH event stream](https://gerrit-review.googlesource.com/Documentation/cmd-stream-events.html) when Changes are created or updated `patchset-created`, when they are merged `change-merged`, or receive comments `comment-added`. There are other [events too](https://gerrit-review.googlesource.com/Documentation/cmd-stream-events.html#_schema).

This is not a stable fork.

# Building

`./build.sh` will produce `gerrit-event-handler` for use. See the **Features** section below for usage examples.

# Development Environment

Services are defined in `docker-compose.yaml` and use `.env`.

## Gerrit Service

```
GERRIT_PORT=29418
GERRIT_WEB_PORT=8080

# Clean and docker-compose a gerrit instance up
./recreate_gerrit_instance.sh

# Generates an SSH Keypair `gerrit-local-ssh-keypair-$RANDOM`
# Authorizes the public key as `admin` in Gerrit
# Creates `test-project` and pushes Change 1
./setup_gerrit.sh
```

To put more changes in Gerrit, use the local keypair `GIT_SSH_COMMAND=ssh -i $(pwd)/$PRIVATE_KEY` from the `test-project/` path. Then create **Change-2**
```
C=Change-2 ; echo $C >> README.md ; git add -u ; git commit -m "$C" -m "Updated readme with $C" ; git push origin HEAD:refs/for/main
```

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
## Should have Opt-In Buildkite integration

Given the Gerrit Event Handler has multiple features
And we want to opt-out of all features by default
And we want a way to opt-in to using the BuildKite integration
Then `--enable-buildkite-integration` should enable it

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

----

![basic design](https://github.com/mrmod/gerrit-buildkite/blob/version-22/Design.png?raw=true)

Story of the yak
* It'd be nice to allow different backends to be used.
* What if Gerrit events could be accepted on different topologies like SSH, Kineses, or webhooks?
* Logging could be better.
* Could Gerrit event dispatch be driven by YAML and text templates?

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
