package rewriter

import (
	"fmt"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/graphql"
	"rewriting-history/internal/pkg/repo"
	"time"

	"github.com/go-git/go-git/v5"
	. "github.com/go-git/go-git/v5/_examples"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

const (
	minWidth = 3
	appender = "appender.txt"
)

type Rewriter struct {
	config    *configs.Config
	repo      *git.Repository
	startDate time.Time
	endDate   time.Time

	ghGraphql *graphql.GhGraphql
}

func NewRewriter(cfg *configs.Config) *Rewriter {
	ghGraphql := graphql.NewGhGraphql(cfg)

	startDate, err := getSunday(cfg.StartWeek)
	if err != nil {
		logrus.Fatalf("Get first Saturday failed: %v", err)
	}
	endDate := latestSunday(nil)
	logrus.Infof("startDate: %v, endDate: %v", startDate, endDate)

	return &Rewriter{
		config:    cfg,
		startDate: startDate,
		endDate:   endDate,
		ghGraphql: ghGraphql,
	}
}

func (r *Rewriter) Run() error {
	err := r.sanityCheck()
	if err != nil {
		return fmt.Errorf("sanity check failed: %w", err)
	}

	err = r.prepare()
	if err != nil {
		return fmt.Errorf("prepare repo failed: %w", err)
	}
	logrus.Infof("prepare repo success")

	err = r.drawBackground()
	if err != nil {
		return fmt.Errorf("draw backgroud failed: %w", err)
	}
	logrus.Infof("draw backgroud success")

	err = r.drawForeground()
	if err != nil {
		return fmt.Errorf("draw foreground failed: %w", err)
	}
	logrus.Infof("draw foreground success")

	return nil
}

func (r *Rewriter) GenerateCommit() plumbing.Hash {
	w, err := r.repo.Worktree()
	CheckIfError(err)

	// git add $appender
	_, err = w.Add(appender)
	CheckIfError(err)

	// git commit -m $message
	h, err := w.Commit("New content", &git.CommitOptions{})
	CheckIfError(err)
	return h
}

func (r *Rewriter) sanityCheck() error {
	logrus.Info("sanity checking...")
	//weeks between start and end date
	weeks := int(r.endDate.Sub(r.startDate).Hours()/24/7) + 1
	if weeks < minWidth {
		return fmt.Errorf("weeks between start and end date is less than %d", minWidth)
	}

	//commitsByDay, err := r.ghGraphql.CommitsByDay()
	//if err != nil {
	//	return fmt.Errorf("get commits by day failed: %w", err)
	//}
	//maxDailyCommit, err := graphql.MaxCommits(r.startDate, r.endDate, commitsByDay)
	//if err != nil {
	//	return fmt.Errorf("get max daily commit failed: %w", err)
	//}
	//if maxDailyCommit.Commits > r.config.BackgroundCommitsPerDay {
	//	return fmt.Errorf("max daily commit %d is greater than %d", maxDailyCommit.Commits, r.config.BackgroundCommitsPerDay)
	//}

	return nil
}

func (r *Rewriter) prepare() error {
	logrus.Info("preparing...")
	repository, err := repo.CloneRepo(r.config.RepoUrl, r.config.GhToken)
	if err != nil {
		logrus.Fatalf("Clone repo failed: %v", err)
	}
	r.repo = repository

	return nil
}

func (r *Rewriter) drawBackground() error {
	logrus.Info("drawing background...")
	commits, err := repo.GetCommits(r.repo, nil)
	if err != nil {
		return fmt.Errorf("get commits failed: %w", err)
	}

	// Compute the total number of days between the start date and today
	totalDays := int(time.Since(r.startDate).Hours()/24) + 1

	// Compute the target commit count per day
	targetCount := totalDays * r.config.BackgroundCommitsPerDay

	// Check if the current commit count matches the target count
	currentCount := len(commits)
	if currentCount < targetCount {
		// Create arbitrary commits to match the target count
		r.createCommits(targetCount - currentCount)
	}

	r.redistributeCommits(commits, targetCount)

	err = repo.ForcePush(r.repo, r.config.GhToken)
	if err != nil {
		return fmt.Errorf("force push failed: %w", err)
	}
	logrus.Info("force push success")
	return nil
}

func (r *Rewriter) drawForeground() error {
	logrus.Info("drawing foreground...")
	return nil
}

// Create arbitrary commits starting from
func (r *Rewriter) createCommits(extraCount int) {
	commitDate := r.startDate
	for i := 0; i < extraCount; i++ {
		worktree, err := r.repo.Worktree()
		if err != nil {
			fmt.Println("Failed to get worktree:", err)
			return
		}

		// Commit the empty changes with the arbitrary commit message
		commitMessage := fmt.Sprintf("Arbitrary commit %d", i+1)
		commitHash, err := worktree.Commit(commitMessage, &git.CommitOptions{
			Author: &object.Signature{
				Name:  r.config.Author,
				Email: r.config.Email,
				When:  commitDate,
			},
			Committer: &object.Signature{
				Name:  r.config.Author,
				Email: r.config.Email,
				When:  commitDate,
			},
			AllowEmptyCommits: true, // Create an empty commit
		})
		if err != nil {
			fmt.Println("Failed to commit changes:", err)
			return
		}

		commitObj, err := r.repo.CommitObject(commitHash)
		if err != nil {
			fmt.Println("Failed to get commit object:", err)
			return
		}

		// Set the commit date explicitly
		commitObj.Author.When = commitDate
		commitObj.Committer.When = commitDate

		commitObjHash, err := r.repo.CommitObject(commitHash)
		if err != nil {
			fmt.Println("Failed to get commit object:", err)
			return
		}

		commitObjHash.Hash = commitHash
		_, err = r.repo.CommitObject(commitObjHash.Hash)
		if err != nil {
			fmt.Println("Failed to recommit changes:", err)
			return
		}

		commitDate = commitDate.Add(time.Hour * 24) // Increment the commit date by one day
	}

	// Redistribute the commits to satisfy the condition
	log, err := r.repo.Log(&git.LogOptions{})
	if err != nil {
		fmt.Println("Failed to retrieve commit history:", err)
		return
	}

	var commits []*object.Commit
	err = log.ForEach(func(commit *object.Commit) error {
		commits = append(commits, commit)
		return nil
	})
	if err != nil {
		fmt.Println("Failed to iterate over commits:", err)
		return
	}

	r.redistributeCommits(commits, len(commits))
}

// Redistribute existing commits to match the target count
func (r *Rewriter) redistributeCommits(commits []*object.Commit, targetCount int) {
	excessCount := len(commits) - targetCount
	lastDay := r.endDate
	remainingCount := excessCount % r.config.BackgroundCommitsPerDay
	distributedCount := excessCount - remainingCount

	if distributedCount > 0 {
		dayCount := distributedCount / r.config.BackgroundCommitsPerDay

		for i, commit := range commits {
			commitDate := lastDay.Add(time.Hour * 24 * time.Duration(i/dayCount))
			commit.Author.When = commitDate
			commit.Committer.When = commitDate
		}
	}

	// Set the remaining commits on the last day
	for i := targetCount; i < targetCount+remainingCount; i++ {
		commitDate := lastDay.Add(time.Hour * 24)
		commit := commits[i]
		commit.Author.When = commitDate
		commit.Committer.When = commitDate
	}
}

// getSunday returns Saturday of the given week.
// weekOffset is the number of weeks in the past year.
func getSunday(weekOffset int) (time.Time, error) {
	now := time.Now()
	latestSunday := latestSunday(&now)

	// Calculate the start date by subtracting 365 days from the current date
	startDate := latestSunday.AddDate(0, 0, -365)

	// Find the first Saturday within the range
	for startDate.Weekday() != time.Sunday {
		startDate = startDate.AddDate(0, 0, 1)
	}

	sunday := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	sunday = sunday.AddDate(0, 0, (weekOffset-1)*7)

	// Ensure that the result does not exceed the current date
	if sunday.After(now) {
		return time.Time{}, fmt.Errorf("sunday is after now")
	}

	return sunday, nil
}

func latestSunday(date *time.Time) time.Time {
	// check if nil
	if date == nil {
		now := time.Now()
		date = &now
	}
	return date.Truncate(24 * time.Hour).Add(-time.Duration(date.Weekday()) * 24 * time.Hour)
}
