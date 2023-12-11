package rewriter

import (
	"contribution-painter/configs"
	"contribution-painter/internal/domain"
	"contribution-painter/internal/pkg/dict"
	"contribution-painter/internal/pkg/graphql"
	"contribution-painter/internal/pkg/helper"
	"contribution-painter/internal/pkg/repo"
	"contribution-painter/internal/pkg/stat"
	"fmt"
	"sort"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

type Rewriter struct {
	rewriterCfg configs.Rewriter
	gitCfg      configs.GitInfo

	repo         *git.Repository
	startDate    time.Time
	endDate      time.Time
	currentState []stat.CommitStat

	stats *stat.ContributionStats
	dict  domain.Dictionary
}

func NewRewriter(cfg configs.Configuration) *Rewriter {
	ghGraphql := graphql.NewGhGraphql(cfg.GitInfo)

	startDate, err := getStartSunday(time.Now(), cfg.Rewriter.LeadingColumns)
	if err != nil {
		logrus.Fatalf("Get first Saturday failed: %v", err)
	}

	return &Rewriter{
		rewriterCfg: cfg.Rewriter,
		gitCfg:      cfg.GitInfo,
		startDate:   startDate,
		stats:       stat.NewContributionStats(ghGraphql),
		dict:        dict.NewDictionary(domain.Font(cfg.Rewriter.Font)),
	}
}

func (r *Rewriter) getEndDate() time.Time {
	c := r.rewriterCfg
	letters, err := r.dict.GetLetters(c.TargetLetters, c.LetterSpacing, c.LeadingColumns, c.TrailingColumns)
	if err != nil {
		logrus.Fatalf("Get letters failed: %v", err)
	}

	width := 0
	for _, letter := range letters {
		width += len(letter[0])
	}

	endDate := r.startDate.Add(time.Duration(width) * 24 * 7 * time.Hour)
	logrus.Infof("letter width: %d, end date: %s", width, endDate.Format(helper.DateFormat))

	return endDate
}

func (r *Rewriter) Run() error {
	err := r.prepare()
	if err != nil {
		return fmt.Errorf("prepare repo failed: %w", err)
	}

	err = r.printCommitStat()
	if err != nil {
		return fmt.Errorf("print commit stat failed: %w", err)
	}

	_, err = r.drawBackground()
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

func (r *Rewriter) prepare() (err error) {
	logrus.Info("preparing...")
	r.repo, err = repo.CloneRepo(r.gitCfg.RepoUrl, r.gitCfg.GhToken)
	if err != nil {
		logrus.Fatalf("Clone repo failed: %v", err)
	}

	currentState, err := r.stats.CommitsByDay()
	if err != nil {
		return fmt.Errorf("get contribution collection failed: %w", err)
	}
	r.currentState = currentState

	r.endDate = r.getEndDate()
	logrus.Infof("painting date range: %s --- %s", r.startDate.Format(helper.DateFormat), r.endDate.Format(helper.DateFormat))

	//weeks between start and end date
	now := time.Now()
	latestSunday := getLatestSunday(now)
	if now.Weekday() != time.Saturday {
		latestSunday = latestSunday.AddDate(0, 0, -7)
	}
	if r.endDate.After(latestSunday) {
		return fmt.Errorf("end date is after now, end date: %s, latestSunday: %s",
			r.endDate.Format(helper.DateFormat), latestSunday.Format(helper.DateFormat))
	}

	return nil
}

func (r *Rewriter) drawBackground() ([]stat.CommitStat, error) {
	logrus.Info("drawing background...")

	// Compute the number of commits by days between the start date and today
	commitStats, err := r.stats.CommitsByDay()
	if err != nil {
		return nil, fmt.Errorf("get commits by day failed: %w", err)
	}

	var commitStatsInDateRange []stat.CommitStat
	for _, cs := range commitStats {
		if cs.Date.Before(r.startDate) || cs.Date.After(r.endDate) {
			continue
		}

		cs.Commits = r.rewriterCfg.BackgroundCommitsPerDay - cs.Commits
		if cs.Commits < 0 {
			logrus.Warnf("commits is less than 0, %s: %d", cs.Date.Format(helper.DateFormat), cs.Commits)
		}

		commitStatsInDateRange = append(commitStatsInDateRange, cs)
	}

	dailyCommits, err := r.createDailyCommits(commitStatsInDateRange)
	if err != nil {
		return nil, fmt.Errorf("create daily commits failed: %w", err)
	}

	// Commit to working tree
	err = r.commitToWorkTree(dailyCommits)
	if err != nil {
		return nil, fmt.Errorf("commit to work tree failed: %w", err)
	}

	return commitStatsInDateRange, nil
}

func (r *Rewriter) drawForeground() error {
	logrus.Info("drawing foreground...")
	c := r.rewriterCfg
	letters, err := r.dict.GetLetters(c.TargetLetters, c.LetterSpacing, c.LeadingColumns, c.TrailingColumns)
	if err != nil {
		logrus.Fatalf("Get letters failed: %v", err)
	}

	// group commits by date
	commits, err := repo.GetCommits(r.repo, &git.LogOptions{})
	if err != nil {
		return fmt.Errorf("get commits failed: %w", err)
	}
	commitMap := make(map[time.Time]int)
	for _, commit := range commits {
		commitMap[commit.Author.When.Truncate(24*time.Hour)]++
	}
	existedCommitStats, err := r.stats.CommitsByDay()
	if err != nil {
		return fmt.Errorf("get existed commits failed: %w", err)
	}
	for _, cs := range existedCommitStats {
		if _, ok := commitMap[cs.Date]; !ok {
			commitMap[cs.Date] = cs.Commits
		} else {
			commitMap[cs.Date] += cs.Commits
		}
	}

	var stats []stat.CommitStat
	lettersWithoutLeadingAndTrailingColumns := letters[r.rewriterCfg.LeadingColumns : len(letters)-r.rewriterCfg.TrailingColumns]
	dataCursor := r.startDate
	for _, letter := range lettersWithoutLeadingAndTrailingColumns { // every letter
		if len(letter[0]) != r.dict.FontWidth() {
			dataCursor = dataCursor.AddDate(0, 0, 7)
			continue
		}

		for i := 0; i < r.dict.FontWidth(); i++ { // every column, 0 - 5
			weekStart, weekEnd := 0, 7
			if len(letter) < 7 {
				dataCursor = dataCursor.Add(24 * time.Hour)
				weekStart = 1
				weekEnd = 6
			}

			for j := weekStart; j < weekEnd; j++ { // every row, 0 - 6
				if letter[j][i] == 1 {
					stats = append(stats, stat.CommitStat{
						Date:    dataCursor,
						Commits: r.rewriterCfg.ForegroundCommitsPerDay - commitMap[dataCursor],
					})
				}
				dataCursor = dataCursor.Add(24 * time.Hour)
			}

			if len(letter) < 7 {
				dataCursor = dataCursor.Add(24 * time.Hour)
			}
		}
	}

	dailyCommits, err := r.createDailyCommits(stats)
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

// commitToWorkTree commits dailyCommits to work tree
func (r *Rewriter) commitToWorkTree(dailyCommits []dailyCommit) error {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return fmt.Errorf("get work tree failed: %w", err)
	}

	infoToPrint := make(map[time.Time]int)
	for _, dc := range dailyCommits {
		_, err = worktree.Commit(dc.message, dc.commitOptions)
		if err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}

		infoToPrint[dc.date.Truncate(24*time.Hour)]++
	}

	// Print commit info order by date asc
	var dates []time.Time
	for k := range infoToPrint {
		dates = append(dates, k)
	}
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	for _, date := range dates {
		logrus.Infof("commit %d commits at %s", infoToPrint[date], date.Format(helper.DateFormat))
	}
	return nil
}

func (r *Rewriter) createDailyCommits(commitStatsInDateRange []stat.CommitStat) ([]dailyCommit, error) {
	msgCount := 0

	var dailyCommits []dailyCommit
	for _, cs := range commitStatsInDateRange {
		if cs.Commits <= 0 {
			continue
		}

		dc := r.createCommitByDay(cs, &msgCount)
		dailyCommits = append(dailyCommits, dc...)
	}

	logrus.Infof("create %d daily commits", msgCount)
	return dailyCommits, nil
}

func (r *Rewriter) createCommitByDay(cs stat.CommitStat, globalCount *int) []dailyCommit {
	var dailyCommits []dailyCommit
	for i := 0; i < cs.Commits; i++ {
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
