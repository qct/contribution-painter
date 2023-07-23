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
	config *configs.Config
	repo   *git.Repository
}

func NewSquasher(cfg *configs.Config) *Squasher {
	r, err := repo.CloneRepo(cfg.RepoUrl, cfg.GhToken)
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

	// Create a new tree by applying changes from each commit
	//newTree, err := createSquashedTree(repo, commits)
	//if err != nil {
	//	return fmt.Errorf("failed to create squashed tree: %w", err)
	//}
	//
	//err = s.updateRef(startCommit)
	//if err != nil {
	//	return fmt.Errorf("failed to update reference: %w", err)
	//}
	//
	//err = s.pushChanges()
	//if err != nil {
	//	return fmt.Errorf("failed to push changes: %w", err)
	//}
	//
	//logrus.Info("Repository pushed successfully.")
	return nil
}

// Helper function to create a new tree by applying changes from each commit
//func createSquashedTree(repo *git.Repository, commits []*object.Commit) (*object.Tree, error) {
//	// Create an empty tree
//	tree := object.Tree{}
//
//	// Apply changes from each commit to the new tree
//	for _, commit := range commits {
//		commitTree, err := commit.Tree()
//		if err != nil {
//			return nil, fmt.Errorf("failed to get commit tree: %w", err)
//		}
//
//		// Apply changes from the commit's tree to the new tree builder
//		err = commitTree.Iterate(func(name string, entry *object.TreeEntry) error {
//			// Add or update the entry in the new tree builder
//			err := treeBuilder.Insert(name, entry)
//			if err != nil {
//				return fmt.Errorf("failed to insert entry: %w", err)
//			}
//			return nil
//		})
//		if err != nil {
//			return nil, fmt.Errorf("failed to iterate commit tree: %w", err)
//		}
//	}
//
//	// Build the final tree
//	newTreeID, err := treeBuilder.Write()
//	if err != nil {
//		return nil, fmt.Errorf("failed to build tree: %w", err)
//	}
//
//	// Get the new tree object
//	newTree, err := repo.TreeObject(plumbing.NewHash(newTreeID.String()))
//	if err != nil {
//		return nil, fmt.Errorf("failed to get tree object: %w", err)
//	}
//
//	return newTree, nil
//}

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

// Update references to only keep the target commit
//func (s *Squasher) updateRef(targetCommit *object.Commit) error {
//	refName := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", s.config.Squash.TargetBranch))
//	ref, err := s.repo.Reference(refName, true)
//	if err != nil {
//		return fmt.Errorf("failed to get reference: %w", err)
//	}
//
//	refHash := ref.Hash()
//	if refHash != targetCommit.Hash {
//		// Squash all commits into a new commit
//		newCommit, err := s.squashCommits(targetCommit)
//		if err != nil {
//			return fmt.Errorf("failed to squash commits: %w", err)
//		}
//
//		// Update the reference to point to the new squashed commit
//		err = s.repo.Storer.SetReference(plumbing.NewReferenceFromStrings(refName.String(), newCommit.Hash.String()))
//		if err != nil {
//			return fmt.Errorf("failed to update reference: %w", err)
//		}
//	}
//	return nil
//}

// Helper function to squash all commits into a new commit
//func (s *Squasher) squashCommits(targetCommit *object.Commit) (*object.Commit, error) {
//	// Get the commit history starting from the target commit
//	commits, err := s.getCommitsUntil(targetCommit)
//	if err != nil {
//		return nil, err
//	}
//
//	// Create a new tree by applying changes from each commit in reverse order
//	newTree, err := s.createSquashedTree(commits)
//	if err != nil {
//		return nil, err
//	}
//
//	// Create a new commit with the squashed changes
//	newCommit, err := s.createSquashedCommit(targetCommit, newTree)
//	if err != nil {
//		return nil, err
//	}
//
//	return newCommit, nil
//}

// Helper function to create a new tree by applying changes from each commit
//func (s *Squasher) createSquashedTree(commits []*object.Commit) (*object.Tree, error) {
//	// Start with an empty tree
//	newTree, err := s.repo.EmptyTree()
//	if err != nil {
//		return nil, fmt.Errorf("failed to create empty tree: %w", err)
//	}
//
//	// Apply changes from each commit to the new tree
//	for _, commit := range commits {
//		tree, err := commit.Tree()
//		if err != nil {
//			return nil, fmt.Errorf("failed to get commit tree: %w", err)
//		}
//
//		// Apply changes from the commit's tree to the new tree
//		newTree, err = s.applyChanges(newTree, tree)
//		if err != nil {
//			return nil, fmt.Errorf("failed to apply changes: %w", err)
//		}
//	}
//
//	return newTree, nil
//}
//
//// Helper function to apply changes from a source tree to a target tree
//func (s *Squasher) applyChanges(targetTree *object.Tree, sourceTree *object.Tree) (*object.Tree, error) {
//	changes, err := object.DiffTree(targetTree, sourceTree)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get tree diff: %w", err)
//	}
//
//	for _, change := range changes {
//		if change.From.Name == "" {
//			// New file or directory, add it to the target tree
//			err := targetTree.Add(change.To.Name, change.To.Mode, change.To.Hash)
//			if err != nil {
//				return nil, fmt.Errorf("failed to add file or directory: %w", err)
//			}
//		} else {
//			// Modified or deleted file or directory, update or remove it from the target tree
//			if change.To == nil {
//				err := targetTree.Remove(change.From.Name)
//				if err != nil {
//					return nil, fmt.Errorf("failed to remove file or directory: %w", err)
//				}
//			} else {
//				err := targetTree.Add(change.To.Name, change.To.Mode, change.To.Hash)
//				if err != nil {
//					return nil, fmt.Errorf("failed to add file or directory: %w", err)
//				}
//			}
//		}
//	}
//
//	return targetTree, nil
//}
//
//// Helper function to create a new commit with the squashed changes
//func (s *Squasher) createSquashedCommit(targetCommit *object.Commit, newTree *object.Tree) (*object.Commit, error) {
//	parents := []*object.Commit{targetCommit}
//
//	// Create the new commit with the squashed changes
//	newCommit, err := s.repo.CommitObject(newTree.ID(), parents...)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create commit: %w", err)
//	}
//
//	return newCommit, nil
//}

// Helper function to get the commit history until the target commit
func (s *Squasher) getCommitsUntil(targetCommit *object.Commit) ([]*object.Commit, error) {
	var commits []*object.Commit
	commitIter, err := s.repo.Log(&git.LogOptions{From: targetCommit.Hash})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit history: %w", err)
	}
	defer commitIter.Close()

	err = commitIter.ForEach(func(c *object.Commit) error {
		commits = append(commits, c)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to iterate commits: %w", err)
	}

	// Reverse the order of commits to squash them in the correct order
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}

	return commits, nil
}

func (s *Squasher) findTargetCommit(commits []*object.Commit) (start, end *object.Commit) {
	earliestCommit := commits[len(commits)-1]
	latestCommit := commits[0]
	for _, commit := range commits {
		if strings.HasPrefix(commit.Hash.String(), s.config.Squash.StartCommit) {
			start = commit
		}
		if strings.HasPrefix(commit.Hash.String(), s.config.Squash.EndCommit) {
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
	err := repo.ForcePush(s.repo, s.config.GhToken)
	if err != nil {
		return fmt.Errorf("failed to force push: %w", err)
	}

	return nil
}
