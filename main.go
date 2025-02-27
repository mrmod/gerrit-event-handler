package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/buildkite/go-buildkite/buildkite"
	"github.com/mrmod/gerrit-buildkite/backend"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	flagStreamType = flag.String("stream-type", "ssh", "Stream type to use")

	flagGerritSshUrl     = flag.String("gerrit-ssh-url", "ssh://gerrit:29418/project", "Gerrit SSH URL")
	flagGerritSshKeyPath = flag.String("gerrit-ssh-key-path", "/path/to/credentials", "File with ssh private key authorized to Gerrit")

	flagEnableChangeReplication   = flag.Bool("enable-change-replication", false, "Enable change replication on 'patchset-created' event")
	flagReplicationDestinationUrl = flag.String("replication-destination-url", "file:///path/to/destination", "Destination URL for replication")
	flagReplicationSshKeyPath     = flag.String("replication-ssh-key-path", "/path/to/credentials", "File with ssh private key authorized to destination")
	flagReplicationClonePath      = flag.String("replication-clone-path", "", "Path to clone source repository to. If empty, will use system temp dir")

	flagEnableBuildkiteIntegration = flag.Bool("enable-buildkite-integration", false, "Enable Buildkite integration")
	flagBuildkiteOrgSlug           = flag.String("buildkite-org-slug", "org-slug", "Buildkite organization slug")
	flagBuildkitePipelineSlug      = flag.String("buildkite-pipeline-slug", "pipeline-slug", "Buildkite pipeline slug")
	flagBuildkiteApiUrl            = flag.String("buildkite-api-url", "https://api.buildkite.com/v2", "Buildkite API URL")
	flagBuildkiteApiTokenPath      = flag.String("buildkite-api-token-path", "/path/to/credentials", "File with an API token for Buildkite. Token should have write_builds permission")

	flagBuildkiteWebhookHandlerDisabled = flag.Bool("disable-buildkite-webhook-handler", true, "Disable Buildkite webhook handler when passed")
	flagWebhookHandlerPort              = flag.String("webhook-handler-port", "10005", "Port to listen for Buildkite webhook events. Ex: 8080")

	flagLoggingTraceEnabled = flag.Bool("enable-trace-logging", false, "Enable trace logging")
	flagLoggingDebugEnabled = flag.Bool("enable-debug-logging", false, "Enable debug logging")
)

func handleSSHEventStream() {
	// Buffer up to 16 events in the stream
	eventStream := make(chan Event, 16)

	_backend := backend.NewRedisBackend()
	client, err := NewGerritSSHClient(*flagGerritSshUrl, *flagGerritSshKeyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Gerrit SSH client")
	}

	if !*flagBuildkiteWebhookHandlerDisabled {
		log.Debug().
			Str("webhookHandlerPort", *flagWebhookHandlerPort).
			Msg("Starting Buildkite webhook handler")
		webhookStream := make(chan BuildkiteWebhook, 16)
		webhookHandler, err := NewBuildkiteWebhookHandler(*flagBuildkiteOrgSlug, *flagBuildkitePipelineSlug, *flagBuildkiteApiUrl, *flagBuildkiteApiTokenPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create Buildkite webhook handler")
		}
		webhookHandler.HookEvents = webhookStream

		go func() {
			log.Debug().Str("port", *flagWebhookHandlerPort).Msg("Listening for Buildkite webhook events")
			http.ListenAndServe(":"+*flagWebhookHandlerPort, webhookHandler)
		}()

		webhookHandler.Backend = _backend
		log.Info().Msg("Started Webhook event handler")
		go HandleWebhookEvents(webhookStream, client, _backend)
	}

	apiUrl, err := url.Parse(*flagBuildkiteApiUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse Buildkite Api Url")
	}
	apiToken, err := readToken(*flagBuildkiteApiTokenPath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to read api token: %s", *flagBuildkiteApiTokenPath)
	}

	apiTransport, err := buildkite.NewTokenConfig(apiToken, *flagLoggingTraceEnabled)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Buildkite Api client")
	}

	log.Debug().Str("host", apiUrl.Host).Msgf("Setting API host")

	pipeline := &Pipeline{
		OrgSlug:      *flagBuildkiteOrgSlug,
		PipelineSlug: *flagBuildkitePipelineSlug,
		ApiUrl:       apiUrl,
		ApiClient:    apiTransport.Client(),
	}
	// TODO: An EventHandler should have an Setup(EventRouter{}) sync Function
	// TODO: An EventHandler should have a Handle(InstrumentedEvent{Event{}, :TraceId}, ResultChan{:*Error, :TraceId}) async Function
	// TODO: An IntrumentedIntegration should have GetResult(:TraceId) {Done, Error, Running, Pending} sync Function. The order allows `> Done` guard.
	if *flagEnableBuildkiteIntegration {
		log.Debug().Msg("Buildkite integration enabled")
		eventRouter["patchset-created"] = append(eventRouter["patchset-created"], HandlePatchsetCreated)
		eventRouter["comment-added"] = append(eventRouter["comment-added"], HandleCommentAdded)
		eventRouter["ref-updated"] = append(eventRouter["ref-updated"], HandleRefUpdated)
	}

	if *flagEnableChangeReplication {
		log.Debug().Msg("Change replication enabled")
		destinationUrl, err := url.Parse(*flagReplicationDestinationUrl)
		if err != nil {
			log.
				Fatal().
				Err(err).
				Str("destinationUrl", *flagReplicationDestinationUrl).
				Msg("Failed to parse replication destination URL")
		}

		destinationRepository := &GitSSHRemote{
			URL:        destinationUrl,
			SshKeyPath: *flagReplicationSshKeyPath,
		}
		replicator := NewSSHReplicator(&client.GitSSHRemote, destinationRepository, *flagReplicationClonePath)
		handleReplication := func(event Event, p BuildPipeline, b backend.Backend) error {
			// Replicate from refs/changes/01/2/1 to change-3
			return replicator.Replicate(event.PatchSet.Ref, fmt.Sprintf("change-%d", event.Change.Number))
		}

		eventRouter["patchset-created"] = append(eventRouter["patchset-created"], handleReplication)
	}
	go client.Handle(eventStream, pipeline, _backend)
	log.Info().Msg("Listening for Gerrit events")
	client.Listen(eventStream)
}

func initFlags() {
	flag.Parse()
}
func main() {
	initFlags()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	if *flagLoggingDebugEnabled {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if *flagLoggingTraceEnabled {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	switch *flagStreamType {
	case "ssh":
		handleSSHEventStream()
	default:
		log.Fatal().Str("streamType", *flagStreamType).Msg("Unknown stream type")
	}

}
