package postCmd

import (
	// Stdlib
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"text/tabwriter"

	// Internal
	"github.com/salsaflow/salsaflow/action"
	"github.com/salsaflow/salsaflow/app"
	"github.com/salsaflow/salsaflow/app/appflags"
	"github.com/salsaflow/salsaflow/asciiart"
	"github.com/salsaflow/salsaflow/commands/review/post/constants"
	"github.com/salsaflow/salsaflow/errs"
	"github.com/salsaflow/salsaflow/git"
	"github.com/salsaflow/salsaflow/git/gitutil"
	"github.com/salsaflow/salsaflow/log"
	"github.com/salsaflow/salsaflow/modules"
	"github.com/salsaflow/salsaflow/modules/common"
	"github.com/salsaflow/salsaflow/prompt"
	"github.com/salsaflow/salsaflow/prompt/storyprompt"

	// Other
	"gopkg.in/tchap/gocli.v2"
)

var Command = &gocli.Command{
	UsageLine: `
  post [-fixes=RRID] [-reviewer=REVIEWER] [-open] [REVISION]

  post [-fixes=RRID] [-no_fetch] [-no_rebase] [-ask_once]
       [-pick] [-reviewer=REVIEWER] [-open] [-no_merge] -parent=BRANCH`,
	Short: "post code review requests",
	Long: `
  Post a code review request for each commit specified.

  In case REVISION is specified, the selected revision is posted for review.
  Make sure the Story-Id tag is in the commit message, salsaflow will not try
  to rewrite the commit message for you in case it is not there.

  In case the parent branch BRANCH is specified, all the commits between
  BRANCH and HEAD are selected to be posted for code review. Using git revision
  ranges, these are the commits matching BRANCH..HEAD, or BRANCH.. for short.
  The selected commits are rebased onto the parent branch before posting.
  To prevent rebasing, use -no_rebase. To be asked to pick up the missing
  story ID only once and use it for all commits, set -ask_once.

  Specifying the parent branch implicitly means that the current branch
  is going to be merged into the parent branch. Use -no_merge to change this.

  When no parent branch nor the revision is specified, the last commit
  on the current branch is selected and posted alone into the code review tool.
  `,
	Action: run,
}

var (
	flagAskOnce  bool
	flagFixes    uint
	flagNoFetch  bool
	flagNoMerge  bool
	flagNoRebase bool
	flagOpen     bool
	flagParent   string
	flagPick     bool
	flagReviewer string
)

func init() {
	// Register flags.
	Command.Flags.BoolVar(&flagAskOnce, "ask_once", flagAskOnce,
		"ask once and reuse the story ID for all commits")
	Command.Flags.UintVar(&flagFixes, "fixes", flagFixes,
		"mark the commits as fixing issues in the given review request")
	Command.Flags.BoolVar(&flagNoFetch, "no_fetch", flagNoFetch,
		"do not fetch the upstream repository")
	Command.Flags.BoolVar(&flagNoMerge, "no_merge", flagNoMerge,
		"do not merge the current branch into the parent branch")
	Command.Flags.BoolVar(&flagNoRebase, "no_rebase", flagNoRebase,
		"do not rebase onto the parent branch")
	Command.Flags.BoolVar(&flagOpen, "open", flagOpen,
		"open the review requests in the browser")
	Command.Flags.StringVar(&flagParent, "parent", flagParent,
		"branch to be used in computing the revision range")
	Command.Flags.BoolVar(&flagPick, "pick", flagPick,
		"pick only some of the selected commits for review")
	Command.Flags.StringVar(&flagReviewer, "reviewer", flagReviewer,
		"reviewer to assign to the newly created review requests")

	// Register global flags.
	appflags.RegisterGlobalFlags(&Command.Flags)
}

func run(cmd *gocli.Command, args []string) {
	if len(args) > 1 {
		cmd.Usage()
		os.Exit(2)
	}

	app.InitOrDie()

	defer prompt.RecoverCancel()

	var err error
	switch {
	case len(args) == 1:
		err = postRevision(args[0])
	case flagParent != "":
		err = postBranch(flagParent)
	default:
		err = postTip()
	}
	if err != nil {
		errs.Fatal(err)
	}
}
