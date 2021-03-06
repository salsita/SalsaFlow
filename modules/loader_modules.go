package modules

import (
	// Internal
	"github.com/salsaflow/salsaflow/config/loader"
	githubCodeReview "github.com/salsaflow/salsaflow/modules/code_review/github"
	noopReview "github.com/salsaflow/salsaflow/modules/code_review/noop"
	githubIssueTracking "github.com/salsaflow/salsaflow/modules/issue_tracking/github"
	"github.com/salsaflow/salsaflow/modules/issue_tracking/pivotaltracker"
	githubReleaseNotes "github.com/salsaflow/salsaflow/modules/release_notes/github"
)

var registeredModules = []loader.Module{
	githubCodeReview.NewModule(),
	githubIssueTracking.NewModule(),
	githubReleaseNotes.NewModule(),
	noopReview.NewModule(),
	pivotaltracker.NewModule(),
}
