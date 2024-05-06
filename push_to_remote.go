package main

import (
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

type SSHReplicator struct {
	// SourceRepository Get the Gerrit URL and an SSH Key for cloning from Gerrit. Ex: ssh://user@gerrit.site:29418/repo/.git
	SourceRepository *GitSSHRemote
	// CloneOutputPath Local path where Gerrit is cloned to
	CloneOutputPath string
	// DestinationRepository Get SSH URL for a destination repository. Ex: ssh://user@remote.com:23456/repo/.git
	DestinationRepository *GitSSHRemote
	useCloneOutputTemp    bool
	// When false will fail when cloning into an existing directory
	// This is permitted by default as a replicator may handle the replication
	// of several branches and want to re-use the cloned repo for each branch
	useExistingClone bool
}

type ChangeReplicator interface {
	Replicate(srcRef, destRef string) error
}

func (s *SSHReplicator) Close() error {
	if s.useCloneOutputTemp {
		log.Trace().Msgf("Removing cloneOutputTemp directory %s", s.CloneOutputPath)
		return os.RemoveAll(s.CloneOutputPath)
	}
	return nil
}

func NewSSHReplicator(sourceRepository, destinationRepository *GitSSHRemote, cloneOutputPath string) *SSHReplicator {

	sshReplicator := &SSHReplicator{
		SourceRepository:      sourceRepository,
		CloneOutputPath:       cloneOutputPath,
		DestinationRepository: destinationRepository,
		useExistingClone:      true,
	}

	if cloneOutputPath == "" {
		_cloneOutputPath := os.TempDir()
		log.Trace().Msgf("Using temp directory for clone output: %s", _cloneOutputPath)
		sshReplicator.useCloneOutputTemp = true
		sshReplicator.CloneOutputPath = _cloneOutputPath
	}
	log.Debug().
		Str("sourceRepository", sourceRepository.String()).
		Str("destinationRepository", destinationRepository.String()).
		Str("cloneOutputPath", sshReplicator.CloneOutputPath).
		Bool("useExistingClone", sshReplicator.useExistingClone).
		Msg("Created SSHReplicator")
	return sshReplicator
}
func (s *SSHReplicator) Replicate(srcRef, destRef string) error {

	cloneArgs := buildCloneCommandArgs(s.SourceRepository.String(), s.CloneOutputPath)
	gitSshCommand := ""
	if s.SourceRepository.SshKeyPath != "" {
		gitSshCommand += " -i " + s.SourceRepository.SshKeyPath
	}
	if s.SourceRepository.URL.User != nil {
		gitSshCommand += " -l" + s.SourceRepository.URL.User.Username()
	}

	cloneCmd := createGitCommand(cloneArgs, []string{"GIT_SSH_COMMAND=ssh " + gitSshCommand})

	if err := execGitCommand(cloneCmd); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return err
		}
		log.Err(err).Msgf("exec git clone failed %d", exitError.ExitCode())
		cloneExists := strings.HasSuffix("already exists and is not an empty directory.\n", string(exitError.Stderr))

		if cloneExists && !s.useExistingClone {
			log.Err(err).
				Str("cloneOutputPath", s.CloneOutputPath).
				Msg("Found existing clone directory, but useExistingClone is false")
			return err
		}
		log.Warn().
			Str("cloneOutputPath", s.CloneOutputPath).
			Msg("Clone destination already exists and is not an empty directory")

		if !cloneExists && err != nil {
			return err
		}

	}

	if err := os.Chdir(s.CloneOutputPath); err != nil {
		return err
	}

	fetchCmd := createGitCommand(buildFetchCommandArgs(srcRef), []string{"GIT_SSH_COMMAND=ssh " + gitSshCommand})
	if err := execGitCommand(fetchCmd); err != nil {
		return err
	}

	resetFetchHeadCmd := createGitCommand(buildResetFetchHeadCommandArgs(), []string{})
	if err := execGitCommand(resetFetchHeadCmd); err != nil {
		return err
	}

	forcePushCmd := createGitCommand(buildForcePushCommandArgs(s.DestinationRepository.String(), destRef), []string{"GIT_SSH_COMMAND=ssh " + gitSshCommand})
	if err := execGitCommand(forcePushCmd); err != nil {
		return err
	}

	return nil
}

func buildCloneCommandArgs(remoteUrl, destPath string) []string {
	return []string{
		"clone",
		remoteUrl,
		destPath,
	}
}
func buildFetchCommandArgs(ref string) []string {
	return []string{
		"fetch",
		"origin",
		ref,
	}
}
func buildResetFetchHeadCommandArgs() []string {
	return []string{
		"reset",
		"--hard",
		"FETCH_HEAD",
	}
}

func buildForcePushCommandArgs(remoteUrl, ref string) []string {
	return []string{
		"push",
		"--force",
		remoteUrl,
		"HEAD:" + ref,
	}
}

func createGitCommand(args []string, env []string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Env = append(cmd.Env, env...)
	log.Debug().Str("cmd", cmd.String()).Msg("Created git command")
	return cmd
}
func execGitCommand(cmd *exec.Cmd) error {
	log.Debug().Str("cmd", cmd.String()).Msg("Executing git command")
	o, err := cmd.CombinedOutput()
	if err != nil {
		log.Err(err).Msgf("Failed to clone branch %s", string(o))
		return err
	}
	return nil
}
