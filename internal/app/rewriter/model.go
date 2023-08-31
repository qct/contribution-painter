package rewriter

import (
	"time"

	"github.com/go-git/go-git/v5"
)

type dailyCommit struct {
	date          time.Time
	message       string
	commitOptions *git.CommitOptions
}
