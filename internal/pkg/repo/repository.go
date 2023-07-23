package repo

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/sirupsen/logrus"
)

// CloneRepo Clones the given repository, creating the remote, the local branches
// and fetching the objects, everything in memory
func CloneRepo(repoUrl, ghToken string) (*git.Repository, error) {
	logrus.Infof("Cloning repo: %s", repoUrl)
	r, err := git.CloneContext(context.Background(), memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: repoUrl,
		Auth: &http.BasicAuth{
			Username: "token",
			Password: ghToken,
		},
	})
	return r, err
}

func ForcePush(r *git.Repository, ghToken string) error {
	logrus.Info("Force pushing changes")
	err := r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: "token",
			Password: ghToken,
		},
		Force: true,
	})
	return err
}

func GetCommits(repo *git.Repository, opts *git.LogOptions) ([]*object.Commit, error) {
	// Get the commit history
	options := opts
	if options == nil {
		options = &git.LogOptions{}
	}

	commitIter, err := repo.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve commit history: %w", err)
	}

	// Iterate over the commits
	var commits []*object.Commit
	err = commitIter.ForEach(func(commit *object.Commit) error {
		commits = append(commits, commit)
		return nil
	})
	commitIter.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to iterate over commits: %w", err)
	}
	return commits, nil
}
