package postCmd

import (
	// Stdlib
	"errors"
	"fmt"

	// Internal
	"github.com/salsaflow/salsaflow/action"
	"github.com/salsaflow/salsaflow/asciiart"
	"github.com/salsaflow/salsaflow/errs"
	"github.com/salsaflow/salsaflow/git"
	"github.com/salsaflow/salsaflow/git/gitutil"
	"github.com/salsaflow/salsaflow/log"
)

func postBranch(parentBranch string) (err error) {
	// Load the git-related config.
	gitConfig, err := git.LoadConfig()
	if err != nil {
		return err
	}
	var (
		remoteName = gitConfig.RemoteName
	)

	// Get the current branch name.
	currentBranch, err := gitutil.CurrentBranch()
	if err != nil {
		return err
	}

	if !flagNoFetch {
		// Fetch the remote repository.
		task := "Fetch the remote repository"
		log.Run(task)

		if err := git.UpdateRemotes(remoteName); err != nil {
			return errs.NewError(task, err)
		}
	}

	// Make sure the parent branch is up to date.
	task := fmt.Sprintf("Make sure reference '%v' is up to date", parentBranch)
	log.Run(task)
	if err := git.EnsureBranchSynchronized(parentBranch, remoteName); err != nil {
		return errs.NewError(task, err)
	}

	// Make sure the current branch is up to date.
	task = fmt.Sprintf("Make sure branch '%v' is up to date", currentBranch)
	log.Run(task)
	if err = git.EnsureBranchSynchronized(currentBranch, remoteName); err != nil {
		return errs.NewError(task, err)
	}

	// Get the commits to be posted
	task = "Get the commits to be posted for code review"
	commits, err := git.ShowCommitRange(parentBranch + "..")
	if err != nil {
		return errs.NewError(task, err)
	}

	// Make sure there are no merge commits.
	if err := ensureNoMergeCommits(commits); err != nil {
		return err
	}

	// Prompt the user to confirm.
	if err := promptUserToConfirmCommits(commits); err != nil {
		return err
	}

	// Rebase the current branch on top the parent branch.
	if !flagNoRebase {
		commits, err = rebase(currentBranch, parentBranch)
		if err != nil {
			return err
		}
	}

	// Ensure the Story-Id tag is there.
	commits, _, err = ensureStoryId(commits)
	if err != nil {
		return err
	}

	// Get data on the current branch.
	task = fmt.Sprintf("Get data on branch '%v'", currentBranch)
	remoteCurrentExists, err := git.RemoteBranchExists(currentBranch, remoteName)
	if err != nil {
		return errs.NewError(task, err)
	}
	currentUpToDate, err := git.IsBranchSynchronized(currentBranch, remoteName)
	if err != nil {
		return errs.NewError(task, err)
	}

	// Merge the current branch into the parent branch unless -no_merge.
	pushTask := "Push the current branch"
	if flagNoMerge {
		// In case the user doesn't want to merge,
		// we need to push the current branch.
		if !remoteCurrentExists || !currentUpToDate {
			if err := push(remoteName, currentBranch); err != nil {
				return errs.NewError(pushTask, err)
			}
		}
	} else {
		// Still push the current branch if necessary.
		if remoteCurrentExists && !currentUpToDate {
			if err := push(remoteName, currentBranch); err != nil {
				return errs.NewError(pushTask, err)
			}
		}

		// Merge the branch into the parent branch
		mergeTask := fmt.Sprintf("Merge branch '%v' into branch '%v'", currentBranch, parentBranch)
		log.Run(mergeTask)
		act, err := merge(mergeTask, currentBranch, parentBranch)
		if err != nil {
			return err
		}

		// Push the parent branch.
		if err := push(remoteName, parentBranch); err != nil {
			// In case the push fails, we revert the merge as well.
			if err := act.Rollback(); err != nil {
				errs.Log(err)
			}
			return errs.NewError(mergeTask, err)
		}

		// Register a rollback function that just says that
		// a pushed merge cannot be reverted.
		defer action.RollbackOnError(&err, action.ActionFunc(func() error {
			log.Rollback(mergeTask)
			hint := "\nCannot revert merge that has already been pushed.\n"
			return errs.NewErrorWithHint(
				"Revert the merge", errors.New("merge commit already pushed"), hint)
		}))
	}

	// Post the review requests.
	if err := postCommitsForReview(commits); err != nil {
		return err
	}

	// In case there is no error, tell the user they can do next.
	return printFollowup()
}

func rebase(currentBranch, parentBranch string) ([]*git.Commit, error) {
	// Tell the user what is happening.
	task := fmt.Sprintf("Rebase branch '%v' onto '%v'", currentBranch, parentBranch)
	log.Run(task)

	// Do the rebase.
	if err := git.Rebase(parentBranch); err != nil {
		ex := errs.Log(errs.NewError(task, err))
		asciiart.PrintGrimReaper("GIT REBASE FAILED")
		fmt.Printf(`Git failed to rebase your branch onto '%v'.

The repository might have been left in the middle of the rebase process.
In case you do not know how to handle this, just execute

  $ git rebase --abort

to make your repository clean again.

In any case, you have to rebase your current branch onto '%v'
if you want to continue and post a review request. In the edge cases
you can as well use -no_rebase to skip this step, but try not to do it.
`, parentBranch, parentBranch)
		return nil, ex
	}

	// Reload the commits.
	task = "Get the commits to be posted for code review, again"
	commits, err := git.ShowCommitRange(parentBranch + "..")
	if err != nil {
		return nil, errs.NewError(task, err)
	}

	// Return new commits.
	return commits, nil
}

func merge(mergeTask, current, parent string) (act action.Action, err error) {
	// Remember the current branch hash.
	currentSHA, err := git.BranchHexsha(current)
	if err != nil {
		return nil, err
	}

	// Checkout the parent branch so that we can perform the merge.
	if err := git.Checkout(parent); err != nil {
		return nil, err
	}
	// Checkout the current branch on return to be consistent.
	defer func() {
		if ex := git.Checkout(current); ex != nil {
			if err == nil {
				err = ex
			} else {
				errs.Log(ex)
			}
		}
	}()

	// Perform the merge.
	// Use --no-ff in case -merge_no_ff is set.
	if flagMergeNoFF {
		err = git.Merge(current, "--no-ff")
	} else {
		err = git.Merge(current)
	}
	if err != nil {
		return nil, err
	}

	// Return a rollback action.
	return action.ActionFunc(func() (err error) {
		log.Rollback(mergeTask)
		task := fmt.Sprintf("Reset branch '%v' to the original position", current)

		// Get the branch is the current branch now.
		currentNow, err := gitutil.CurrentBranch()
		if err != nil {
			return errs.NewError(task, err)
		}

		// Checkout current in case it is not the same as the current branch now.
		if currentNow != current {
			if err := git.Checkout(current); err != nil {
				return errs.NewError(task, err)
			}
			defer func() {
				if ex := git.Checkout(currentNow); ex != nil {
					if err == nil {
						err = ex
					} else {
						errs.Log(ex)
					}
				}
			}()
		}

		// Reset the branch to the original position.
		if err := git.Reset("--keep", currentSHA); err != nil {
			return errs.NewError(task, err)
		}
		return nil
	}), nil
}
