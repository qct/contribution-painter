package rewriter

import (
	"fmt"
)

func (r *Rewriter) printCommitStat() error {
	if err := r.stats.PrintCommitStat(); err != nil {
		return fmt.Errorf("print commit stat failed: %w", err)
	}

	return nil
}
