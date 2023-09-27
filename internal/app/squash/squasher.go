package squash

import (
	"errors"
	"fmt"
	"io"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/repo"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

type Squasher struct {
	config *configs.Configuration
	repo   *git.Repository
}

func NewSquasher(cfg *configs.Configuration) *Squasher {
	r, err := repo.CloneRepo(cfg.GitInfo.RepoUrl, cfg.GitInfo.GhToken)
	if err != nil {
		logrus.Fatalf("Clone repo failed: %v", err)
	}
	return &Squasher{config: cfg, repo: r}
}

func (s *Squasher) Squash() error {
	commits, err := repo.GetCommits(s.repo, nil)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	startCommit, endCommit := s.findTargetCommit(commits)
	logrus.Infof("start: %s(%s), end: %s(%s)", startCommit.Hash, startCommit.Message, endCommit.Hash, endCommit.Message)

	canSquashCommits, err := s.canSquashCommits(startCommit, endCommit)
	if err != nil {
		return fmt.Errorf("failed to check squash commits: %w", err)
	}
	logrus.Infof("squash commits: %v", canSquashCommits)

	logrus.Info("Repository pushed successfully.")
	return nil
}

// canSquashCommits Check if commits between commit A and commit B can be squashed
func (s *Squasher) canSquashCommits(commitFrom *object.Commit, commitTo *object.Commit) (bool, error) {
	// Check if commitFrom is an ancestor of commitTo
	isAncestor, err := commitFrom.IsAncestor(commitTo)
	if err != nil {
		return false, fmt.Errorf("failed to check ancestor relationship: %w", err)
	}
	if !isAncestor {
		return false, nil // commitFrom is not an ancestor of commitTo, cannot be squashed
	}

	// Retrieve the commit history from commitFrom to commitTo
	commitIter, err := s.repo.Log(&git.LogOptions{
		From: commitTo.Hash,
	})
	if err != nil {
		return false, fmt.Errorf("failed to retrieve commit history: %w", err)
	}
	defer commitIter.Close()

	// Check if any commits have multiple parents or merges
	for {
		commit, err := commitIter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break // No more commits to process
			} else {
				return false, fmt.Errorf("failed to iterate commits: %w", err)
			}
		}
		logrus.Debugf("commit hash: %s, time: %s", commit.Hash, commit.Author.When)

		if commit.NumParents() > 1 {
			return false, nil // Found a merge commit, cannot be squashed
		}

		if strings.HasPrefix(commit.Hash.String(), commitFrom.Hash.String()) {
			break
		}
	}

	return true, nil // All commits can be squashed
}

func (s *Squasher) findTargetCommit(commits []*object.Commit) (start, end *object.Commit) {
	earliestCommit := commits[len(commits)-1]
	latestCommit := commits[0]
	for _, commit := range commits {
		if s.config.Squash.StartCommit != "" && strings.HasPrefix(commit.Hash.String(), s.config.Squash.StartCommit) {
			start = commit
		}
		if s.config.Squash.EndCommit != "" && strings.HasPrefix(commit.Hash.String(), s.config.Squash.EndCommit) {
			end = commit
		}
	}
	if start == nil {
		start = earliestCommit
	}
	if end == nil {
		end = latestCommit
	}
	return
}

func (s *Squasher) pushChanges() error {
	// Force push the updated history
	err := repo.ForcePush(s.repo, s.config.GitInfo.GhToken)
	if err != nil {
		return fmt.Errorf("failed to force push: %w", err)
	}

	return nil
}
