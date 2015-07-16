package issues

import (
	// Stdlib
	"bufio"
	"fmt"
	"regexp"
	"strings"

	// Vendor
	"github.com/google/go-github/github"
)

// CommitReviewIssue represents a review issue associated with a commit.
type CommitReviewIssue struct {
	// Title
	CommitSHA   string
	CommitTitle string

	// Body
	*ReviewIssueCommonBody
}

func NewCommitReviewIssue(commitSHA, commitTitle string) *CommitReviewIssue {
	ctx := &CommitReviewIssue{
		CommitSHA:             commitSHA,
		CommitTitle:           commitTitle,
		ReviewIssueCommonBody: newReviewIssueCommonBody(),
	}
	ctx.AddCommit(commitSHA, commitTitle, false)
	return ctx
}

// Formatting ------------------------------------------------------------------

func (ctx *CommitReviewIssue) FormatTitle() string {
	return fmt.Sprintf("Review commit %v: %v", ctx.CommitSHA, ctx.CommitTitle)
}

var commitIssueBodyTemplate = fmt.Sprintf(`Commits being reviewed:{{range .CommitList.CommitItems}}
- {{if .Done}}[x]{{else}}[ ]{{end}} {{.CommitSHA}}: {{.CommitTitle}}{{end}}
{{with .ReviewBlockerList.ReviewBlockerItems}}
The following review blockers were opened by the reviewer:{{range .}}
- {{if .Fixed}}[x]{{else}}[ ]{{end}} [blocker {{.BlockerNumber}}]({{.CommentURL}}) (commit {{.CommitSHA}}): {{.BlockerSummary}}{{end}}
{{end}}
%v
{{if .UserContent}}{{.UserContent}}{{else}}The content above was generated by SalsaFlow.
You can insert some content here, but not above the line.
{{end}}`, separator)

func (ctx *CommitReviewIssue) FormatBody() string {
	return execTemplate("commit issue body", commitIssueBodyTemplate, ctx)
}

// Parsing ---------------------------------------------------------------------

func parseCommitReviewIssue(issue *github.Issue) (*CommitReviewIssue, error) {
	var (
		title = *issue.Title
		body  = *issue.Body
	)

	// Parse the title.
	titleRegexp := regexp.MustCompile(`^Review commit ([^:]+): (.+)$`)
	match := titleRegexp.FindStringSubmatch(title)
	if len(match) == 0 {
		return nil, &ErrInvalidTitle{issue}
	}
	commitSHA, commitTitle := match[1], match[2]

	// Parse the body.
	// There is actually nothing else than the common part,
	// so we can simply call parseRemainingIssueBody.
	scanner := bufio.NewScanner(strings.NewReader(body))
	bodyCtx, err := parseRemainingIssueBody(issue, scanner, 0)
	if err != nil {
		return nil, err
	}

	// Return the context.
	return &CommitReviewIssue{commitSHA, commitTitle, bodyCtx}, nil
}
