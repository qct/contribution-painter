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

	startDate, err := getSunday(time.Now(), cfg.WeekOffset)
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

func (r *Rewriter) prepare() (err error) {
	logrus.Info("preparing...")
	r.repo, err = repo.CloneRepo(r.config.RepoUrl, r.config.GhToken)
	if err != nil {
		logrus.Fatalf("Clone repo failed: %v", err)
	}

	return nil
}

func (r *Rewriter) drawBackground() error {
	logrus.Info("drawing background...")
	//commits, err := repo.GetCommits(r.repo, nil)
	//if err != nil {
	//	return fmt.Errorf("get commits failed: %w", err)
	//}

	// Compute the number of commits by days between the start date and today
	dailyCommits, err := r.createDailyCommits()
	if err != nil {
		return fmt.Errorf("create daily commits failed: %w", err)
	}

	// Commit to working tree
	err = r.commitToWorkTree(dailyCommits)
	if err != nil {
		return fmt.Errorf("commit to work tree failed: %w", err)
	}

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

// commitToWorkTree commits dailyCommits to work tree
func (r *Rewriter) commitToWorkTree(dailyCommits []dailyCommit) error {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("get work tree failed: %w", err)
	}

	for _, dc := range dailyCommits {
		commitHash, err := worktree.Commit(dc.message, dc.commitOptions)
		if err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}
		logrus.Infof("commit %s success at %s", commitHash.String()[:8], dc.date)
	}

	return nil
}

func (r *Rewriter) createDailyCommits() ([]dailyCommit, error) {
	commitStats, err := r.ghGraphql.CommitsByDay()
	if err != nil {
		return nil, fmt.Errorf("get commits by day failed: %w", err)
	}

	msgCount := 0

	var dailyCommits []dailyCommit
	for _, cs := range commitStats {
		dc := r.createCommitByDay(cs, &msgCount)
		dailyCommits = append(dailyCommits, dc...)
	}

	logrus.Infof("create %d daily commits", msgCount)
	return dailyCommits, nil
}

func (r *Rewriter) createCommitByDay(cs graphql.CommitStats, globalCount *int) []dailyCommit {
	commitsToCreate := r.config.BackgroundCommitsPerDay - cs.Commits
	if commitsToCreate <= 0 {
		return nil
	}

	var dailyCommits []dailyCommit
	for i := 0; i < commitsToCreate; i++ {
		*globalCount++
		dailyCommits = append(dailyCommits, dailyCommit{
			date:    cs.Date,
			message: fmt.Sprintf("Arbitrary commit #%d", *globalCount),
			commitOptions: &git.CommitOptions{
				Author: &object.Signature{
					Name:  r.config.Author,
					Email: r.config.Email,
					When:  cs.Date,
				},
				Committer: &object.Signature{
					Name:  r.config.Author,
					Email: r.config.Email,
					When:  cs.Date,
				},
				AllowEmptyCommits: true, // Create an empty commit
			},
		})
	}

	return dailyCommits
}

// getSunday returns Sunday of the given week.
// weekOffset is the number of weeks in the past year.
func getSunday(now time.Time, weekOffset int) (time.Time, error) {
	latestSunday := latestSunday(&now)

	// Calculate the start date 52 weeks ago
	targetSunday := latestSunday.AddDate(0, 0, (-52+weekOffset)*7)
	targetSunday = time.Date(targetSunday.Year(), targetSunday.Month(), targetSunday.Day(), 0, 0, 0, 0, time.UTC)

	// Ensure that the result does not exceed the current date
	if targetSunday.After(now) {
		return time.Time{}, fmt.Errorf("sunday is after now")
	}

	return targetSunday, nil
}

func latestSunday(date *time.Time) time.Time {
	// check if nil
	if date == nil {
		now := time.Now()
		date = &now
	}
	return date.Truncate(24 * time.Hour).Add(-time.Duration(date.Weekday()) * 24 * time.Hour)
}
