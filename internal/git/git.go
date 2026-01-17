package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func CurrentBranch() (string, error) {
	out, err := run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "main", nil
	}
	return strings.TrimSpace(out), nil
}

func CreateBranch(name, base string) error {
	if _, err := run("checkout", base); err != nil {
		return err
	}
	if _, err := run("pull", "origin", base); err != nil {
		// ignore pull errors for local-only repos
	}
	_, err := run("checkout", "-b", name)
	if err != nil {
		// branch might exist, try checkout
		_, err = run("checkout", name)
	}
	return err
}

func Checkout(branch string) error {
	_, err := run("checkout", branch)
	return err
}

func StashPush(message string) (bool, error) {
	before, _ := run("stash", "list", "-1", "--format=%gd %s")
	if _, err := run("stash", "push", "-m", message); err != nil {
		return false, err
	}
	after, _ := run("stash", "list", "-1", "--format=%gd %s")
	if after != before && strings.Contains(after, message) {
		return true, nil
	}
	return false, nil
}

func StashPop() error {
	_, err := run("stash", "pop")
	return err
}

func Push(branch string) error {
	_, err := run("push", "-u", "origin", branch)
	return err
}

func Merge(branch string) error {
	_, err := run("merge", "--no-edit", branch)
	return err
}

func MergeAbort() error {
	_, _ = run("merge", "--abort")
	return nil
}

func DeleteBranch(branch string) error {
	_, err := run("branch", "-d", branch)
	return err
}

func HasUncommittedChanges() bool {
	out, err := run("status", "--porcelain")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

func ConflictedFiles() ([]string, error) {
	out, err := run("diff", "--name-only", "--diff-filter=U")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

func CommitCount(base string) (int, error) {
	out, err := run("rev-list", "--count", base+"..HEAD")
	if err != nil {
		return 0, err
	}
	var count int
	fmt.Sscanf(strings.TrimSpace(out), "%d", &count)
	return count, nil
}

func WorktreeAdd(path, branch string) error {
	_, err := run("worktree", "add", path, branch)
	return err
}

func WorktreeRemove(path string) error {
	_, err := run("worktree", "remove", "-f", path)
	return err
}

func WorktreePrune() error {
	_, err := run("worktree", "prune")
	return err
}

func RemoteURL() (string, error) {
	out, err := run("remote", "get-url", "origin")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}

func RunInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String(), nil
}
