//go:build integration_tests
// +build integration_tests

package main

import (
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSSHReplicationShouldReplaceExistingRef(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	fsLocalGitRoot, err := os.MkdirTemp("", "local-git-root")
	if err != nil {
		t.Fatalf("Failed to create local-git-root temp dir %s", err)
	}
	defer os.RemoveAll(fsLocalGitRoot)

	if err := exec.Command("git", "init", fsLocalGitRoot).Run(); err != nil {
		t.Fatalf("Failed to init local fs root %s", err)
	}

	fsGerritCloneTemp, err := os.MkdirTemp("", "gerrit-clone")
	if err != nil {
		t.Fatalf("Failed to create gerrit-clone temp dir %s", err)
	}
	defer os.RemoveAll(fsGerritCloneTemp)

	gerritRemoteUrl, err := url.Parse("ssh://admin@localhost:29418/test-project")
	if err != nil {
		t.Fatalf("Failed to parse Gerrit remote URL %s", err)
	}
	localRemoteUrl, err := url.Parse("file://" + fsLocalGitRoot + "/.git")
	if err != nil {
		t.Fatalf("Failed to parse local remote URL %s", err)
	}

	sshKeyPath, _ := filepath.Abs("./local.private.key")

	sshReplicator := &SSHReplicator{
		SourceRepository: &GitSSHRemote{
			URL:        gerritRemoteUrl,
			SshKeyPath: sshKeyPath,
		},
		DestinationRepository: &GitSSHRemote{
			URL: localRemoteUrl,
			// TODO: Unused in file URL
			SshKeyPath: sshKeyPath,
		},
		CloneOutputPath:  fsGerritCloneTemp,
		useExistingClone: true,
	}
	// Push inital patch (nothing exists in remote)
	if err := sshReplicator.Replicate("refs/changes/48/48/1", "change-48"); err != nil {
		t.Fatalf("expected to replication `refs/changes/48/48/1` to `change-48` but failed, %s", err)

	}
	fh, err := os.Open(fsLocalGitRoot + "/.git/refs/heads/change-48")
	if err != nil {
		t.Fatalf("expected to find `change-48` destination repository but failed, %s", err)
	}
	defer fh.Close()

	commitSha, err := io.ReadAll(fh)
	if err != nil {
		t.Fatalf("Failed to read commit sha %s", err)
	}

	if err := sshReplicator.Replicate("refs/changes/48/48/2", "change-48"); err != nil {
		t.Fatalf("expected to replication `refs/changes/48/48/2` to `change-48` but failed, %s", err)

	}
	fhNext, err := os.Open(fsLocalGitRoot + "/.git/refs/heads/change-48")
	if err != nil {
		t.Fatalf("expected to find `change-48` destination repository but failed, %s", err)
	}
	commitShaNext, err := io.ReadAll(fhNext)
	if err != nil {
		t.Fatalf("Failed to read commit sha %s", err)
	}
	if string(commitSha) == string(commitShaNext) {
		t.Fatalf("Expected to find different commit sha %s", err)
	}

}

func TestSSHReplicationShouldPushRefToRemote(t *testing.T) {
	wd, _ := os.Getwd()
	defer os.Chdir(wd)
	fsLocalGitRoot, err := os.MkdirTemp("", "local-git-root")
	if err != nil {
		t.Fatalf("Failed to create local-git-root temp dir %s", err)
	}
	defer os.RemoveAll(fsLocalGitRoot)

	if err := exec.Command("git", "init", fsLocalGitRoot).Run(); err != nil {
		t.Fatalf("Failed to init local fs root %s", err)
	}

	fsGerritCloneTemp, err := os.MkdirTemp("", "gerrit-clone")
	if err != nil {
		t.Fatalf("Failed to create gerrit-clone temp dir %s", err)
	}
	defer os.RemoveAll(fsGerritCloneTemp)

	gerritRemoteUrl, err := url.Parse("ssh://admin@localhost:29418/test-project")
	if err != nil {
		t.Fatalf("Failed to parse Gerrit remote URL %s", err)
	}
	localRemoteUrl, err := url.Parse("file://" + fsLocalGitRoot + "/.git")
	if err != nil {
		t.Fatalf("Failed to parse local remote URL %s", err)
	}

	sshKeyPath, _ := filepath.Abs("local.private.key")

	sshReplicator := &SSHReplicator{
		SourceRepository: &GitSSHRemote{
			URL:        gerritRemoteUrl,
			SshKeyPath: sshKeyPath,
		},
		DestinationRepository: &GitSSHRemote{
			URL: localRemoteUrl,
			// TODO: Unused in file URL
			SshKeyPath: sshKeyPath,
		},
		CloneOutputPath: fsGerritCloneTemp,
	}

	if err := sshReplicator.Replicate("refs/changes/48/48/1", "change-48"); err != nil {
		t.Fatalf("expected to replication `refs/changes/48/48/1` to `change-48` but failed, %s", err)

	}
	_, err = os.Stat(fsLocalGitRoot + "/.git/refs/heads/change-48")

	if err != nil {
		t.Fatalf("expected to find `change-48` destination repository but failed, %s", err)
	}
}

func TestPushToRemote(t *testing.T) {
	fsLocalGitRoot, err := os.MkdirTemp("", "local-git-root")
	if err != nil {
		t.Fatalf("Failed to create local-git-root temp dir %s", err)
	}
	defer os.RemoveAll(fsLocalGitRoot)

	if err := exec.Command("git", "init", fsLocalGitRoot).Run(); err != nil {
		t.Fatalf("Failed to init local fs root %s", err)
	}

	fsGerritCloneTemp, err := os.MkdirTemp("", "gerrit-clone")
	if err != nil {
		t.Fatalf("Failed to create gerrit-clone temp dir %s", err)
	}
	defer os.RemoveAll(fsGerritCloneTemp)

	gerritRemoteUrl := "ssh://localhost:29418/test-project"
	gerritBranch := "refs/changes/48/48/1"
	// remoteBranch := "refs/changes-10"

	identityFile, _ := filepath.Abs("./local.private.key")
	_cloneCommandArgs := buildCloneCommandArgs(gerritRemoteUrl, fsGerritCloneTemp)
	_cloneCommand := createGitCommand(_cloneCommandArgs, []string{"GIT_SSH_COMMAND=ssh -l admin -i " + identityFile})

	if err := execGitCommand(_cloneCommand); err != nil {
		t.Fatalf("Failed to clone branch %s", err)
	}
	if err := os.Chdir(fsGerritCloneTemp); err != nil {
		t.Fatalf("Failed to change to temp clone directory directory %s", err)
	}
	fetchCmd := createGitCommand(buildFetchCommandArgs(gerritBranch), []string{"GIT_SSH_COMMAND=ssh -l admin -i " + identityFile})
	if err := execGitCommand(fetchCmd); err != nil {
		t.Fatalf("Failed to fetch ref %s", err)
	}

	fsLocalUrl := "file://" + fsLocalGitRoot + "/.git"
	forcePushCmdArgs := buildForcePushCommandArgs(fsLocalUrl, "change-48")
	forcePushCmd := createGitCommand(forcePushCmdArgs, []string{"GIT_SSH_COMMAND=ssh -l admin -i " + identityFile})
	if err := execGitCommand(forcePushCmd); err != nil {
		t.Fatalf("Failed to push ref %s", err)
	}
	_, err = os.Stat(fsLocalGitRoot + "/.git/refs/heads/change-48")

	if err != nil {
		t.Fatalf("Failed to find branch 'change-48' %s", err)
	}

}
