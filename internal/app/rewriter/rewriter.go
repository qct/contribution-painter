package rewriter

import (
	"fmt"
	"rewriting-history/configs"
	"rewriting-history/internal/pkg/graphql"
	"rewriting-history/internal/pkg/helper"
	"rewriting-history/internal/pkg/repo"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

const (
	minWidth = 3
)

type Rewriter struct {
	rewriterCfg configs.Rewriter
	gitCfg      configs.GitInfo

	repo      *git.Repository
	startDate time.Time
	endDate   time.Time

	ghGraphql *graphql.GhGraphql
}

func NewRewriter(cfg configs.Configuration) *Rewriter {
	ghGraphql := graphql.NewGhGraphql(cfg.GitInfo)

	startDate, err := getStartSunday(time.Now(), cfg.Rewriter.WeekOffset)
	if err != nil {
		logrus.Fatalf("Get first Saturday failed: %v", err)
	}
	endDate := getLatestSunday(time.Now()).Add(-24 * time.Hour)
	logrus.Infof("painting date range: %s --- %s", startDate.Format(helper.DateFormat), endDate.Format(helper.DateFormat))

	return &Rewriter{
		rewriterCfg: cfg.Rewriter,
		gitCfg:      cfg.GitInfo,
		startDate:   startDate,
		endDate:     endDate,
		ghGraphql:   ghGraphql,
	}
}

func (r *Rewriter) Run() error {
	err := r.sanityCheck()
	if err != nil {
		return fmt.Errorf("sanity check failed: %w", err)
	}

	err = r.printCommitStat()
	if err != nil {
		return fmt.Errorf("print commit stat failed: %w", err)
	}

	err = r.prepare()
	if err != nil {
		return fmt.Errorf("prepare repo failed: %w", err)
	}

	err = r.drawBackground()
	if err != nil {
		return fmt.Errorf("draw backgroud failed: %w", err)
	}

	err = r.drawForeground()
	if err != nil {
		return fmt.Errorf("draw foreground failed: %w", err)
	}

	if !r.rewriterCfg.DryRun {
		err = repo.ForcePush(r.repo, r.gitCfg.GhToken)
		if err != nil {
			return fmt.Errorf("force push failed: %w", err)
		}
		logrus.Info("force push success")
	}

	return nil
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
	r.repo, err = repo.CloneRepo(r.gitCfg.RepoUrl, r.gitCfg.GhToken)
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
	commitsToCreate := r.rewriterCfg.BackgroundCommitsPerDay - cs.Commits
	if commitsToCreate <= 0 {
		return nil
	}

	var dailyCommits []dailyCommit
	for i := 0; i < commitsToCreate; i++ {
		*globalCount++
		dailyCommits = append(dailyCommits, r.createCommit(cs.Date, fmt.Sprintf("Arbitrary commit #%d", *globalCount)))
	}

	return dailyCommits
}

func (r *Rewriter) createCommit(date time.Time, commitMsg string) dailyCommit {
	return dailyCommit{
		date:    date,
		message: commitMsg,
		commitOptions: &git.CommitOptions{
			Author: &object.Signature{
				Name:  r.gitCfg.Author,
				Email: r.gitCfg.Email,
				When:  date,
			},
			Committer: &object.Signature{
				Name:  r.gitCfg.Author,
				Email: r.gitCfg.Email,
				When:  date,
			},
			AllowEmptyCommits: true, // Create an empty commit
		},
	}
}

// getStartSunday returns Sunday of the week (52 - offset) weeks ago.
// weekOffset is the number of weeks to offset from the target date.
func getStartSunday(now time.Time, weekOffset int) (time.Time, error) {
	latestSunday := getLatestSunday(now)

	// Calculate the start date 52 weeks ago
	targetSunday := latestSunday.AddDate(0, 0, (-52+weekOffset)*7)
	targetSunday = time.Date(targetSunday.Year(), targetSunday.Month(), targetSunday.Day(), 0, 0, 0, 0, time.UTC)

	// Ensure that the result does not exceed the current date
	if targetSunday.After(now) {
		return time.Time{}, fmt.Errorf("sunday is after now")
	}

	return targetSunday, nil
}

func getLatestSunday(d time.Time) time.Time {
	return d.Truncate(24 * time.Hour).Add(-time.Duration(d.Weekday()) * 24 * time.Hour)
}
