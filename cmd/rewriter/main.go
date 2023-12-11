package main

import (
	"rewriting-history/configs"
	"rewriting-history/internal/app/rewriter"
	"rewriting-history/internal/pkg/helper"
	"rewriting-history/internal/pkg/logger"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/_examples"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/sirupsen/logrus"
)

var config configs.Config

func main() {
	logger.InitLogger()
	err := configs.LoadConfig("", &config)
	if err != nil {
		logrus.Fatalf("Load config failed: %v", err)
	}

	re := rewriter.NewRewriter(&config)
	err = re.Run()
	if err != nil {
		logrus.Fatalf("Rewriter failed to run: %v", err)
	}
}

func generateCommitPlan(since time.Time, until time.Time) []CommitPlan {
	var plans []CommitPlan
	for d := since; !d.After(until); d = d.AddDate(0, 0, 1) {
		plans = append(plans, CommitPlan{Time: d, Count: 1})
	}
	return plans
}

func commitsNeeded(r *git.Repository, totalNeeded int, since time.Time, until time.Time) (int, []plumbing.Hash) {
	ref, err := r.Head()
	examples.CheckIfError(err)
	examples.Info("Reference: %s", ref.Name().String())

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash(), Since: &since, Until: &until})
	examples.CheckIfError(err)

	// Gets the HEAD history from HEAD, just like this command:
	examples.Info("git logger")
	count := 0
	var existed []plumbing.Hash
	err = cIter.ForEach(func(c *object.Commit) error {
		count++
		//fmt.Println(c)
		existed = append(existed, c.Hash)
		return nil
	})
	examples.CheckIfError(err)
	examples.Info("Total %d commits between %s to %s", count, since.Format(helper.DateTimeFormat), until.Format(helper.DateTimeFormat))
	return totalNeeded - count, existed
}

type CommitPlan struct {
	Time  time.Time
	Count int8
}

func push(r *git.Repository, auth *ssh.PublicKeys) {
	//Push the code to the remote
	err := r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
	examples.CheckIfError(err)
	examples.Info("Pushed to remote.")
}
