package postCmd

func pushCurrentBranch() error {
	task := "Check whether the current branch is to be pushed"

	// Get the current branch.
	currentBranch, err := gitutil.CurrentBranch()
	if err != nil {
		return errs.NewError(task, err)
	}

	// Get the associated upstream branch.
	upstreamBranch, err := git.UpstreamBranch(currentBranch)
	if err != nil {
		return nil
	}

	// In case the current branch is not a tracking branch, we are done.
	if upstreamBranch == nil {
		return nil
	}

	// Make sure the project remote is affected.
	gitConfig, err := git.LoadConfig()
	if err != nil {
		return errs.NewError(task, err)
	}
	remoteName := gitConfig.RemoteName

	if upstreamBranch.Remote != remoteName {
		return nil
	}

	// Check whether the branch is up to date or not.
	upToDate, err := upstreamBranch.IsUpToDate()
	if err != nil {
		return errs.NewError(task, err)
	}
	if upToDate {
		return nil
	}

	// Push the branch
	// Use the Force in case we are not on a core branch.
	args := make([]string, 0, 3)
	msg := fmt.Sprintf("Pushing branch '%v' to synchronize", currentBranch)
	isCore, err := git.IsCoreBranch(currentBranch)
	if err != nil {
		return errs.NewError(task, err)
	}
	if !isCore {
		args = append(args, "-f")
		msg += " (using force)"
	}

	args = append(args, remoteName, currentBranch)

	log.Log(msg)
	task = "Push the current branch"
	if _, err = git.RunCommand("push", args...); err != nil {
		return errs.NewError(task, err)
	}
}