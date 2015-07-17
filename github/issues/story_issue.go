package issues

import (
	// Stdlib
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	// Vendor
	"github.com/google/go-github/github"
)

// StoryReviewIssue represents the review issue type associated with a story.
type StoryReviewIssue struct {
	StoryId      string
	StoryURL     string
	StorySummary string
	TrackerName  string
	StoryKey     string

	*ReviewIssueCommonBody
}

func NewStoryReviewIssue(
	storyId string,
	storyURL string,
	storySummary string,
	trackerName string,
	storyKey string,
) *StoryReviewIssue {

	return &StoryReviewIssue{
		StoryId:               storyId,
		StoryURL:              storyURL,
		StorySummary:          storySummary,
		TrackerName:           trackerName,
		StoryKey:              storyKey,
		ReviewIssueCommonBody: newReviewIssueCommonBody(),
	}
}

// Formatting ------------------------------------------------------------------

func (ctx *StoryReviewIssue) FormatTitle() string {
	return fmt.Sprintf("Review story %v: %v", ctx.StoryId, ctx.StorySummary)
}

const (
	TagTrackerName = "SF-Issue-Tracker"
	TagStoryKey    = "SF-Story-Key"
)

var storyReviewIssueBodyTemplate = fmt.Sprintf(`Story being reviewed: [{{.StoryId}}]({{.StoryURL}})

%v: {{.TrackerName}}
%v: {{.StoryKey}}

`, TagTrackerName, TagStoryKey)

func (ctx *StoryReviewIssue) FormatBody() string {
	var buffer bytes.Buffer
	ctx.execTemplate(&buffer)
	ctx.ReviewIssueCommonBody.execTemplate(&buffer)
	return buffer.String()
}

func (ctx *StoryReviewIssue) execTemplate(w io.Writer) {
	execTemplate(w, "story issue body", storyReviewIssueBodyTemplate, ctx)
}

// Parsing ---------------------------------------------------------------------

const (
	stateIntroLine = iota + 1
	stateIntroLineTrailingLine
	stateMetadata
)

var (
	titleRegexp       = regexp.MustCompile(`^Review story ([^:]+): (.+)$`)
	introLineRegexp   = regexp.MustCompile(`^Story being reviewed: \[([^\]]+)\]\(([^ ]+)\)`)
	trackerNameRegexp = regexp.MustCompile(fmt.Sprintf("^%v: (.+)$", TagTrackerName))
	storyKeyRegexp    = regexp.MustCompile(fmt.Sprintf("^%v: (.+)$", TagStoryKey))
)

func parseStoryReviewIssue(issue *github.Issue) (*StoryReviewIssue, error) {
	var (
		title = *issue.Title
		body  = *issue.Body
	)

	// Prepare the context to be filled.
	ctx := NewStoryReviewIssue("", "", "", "", "")

	// Parse the title.
	match := titleRegexp.FindStringSubmatch(title)
	if len(match) == 0 {
		return nil, &ErrInvalidTitle{issue}
	}

	ctx.StoryId, ctx.StorySummary = match[1], match[2]

	// Parse the body.
	var lineNo int
	state := stateIntroLine

	scanner := bufio.NewScanner(strings.NewReader(body))
Scanning:
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNo++

		// In case we encounder a separator, we just start adding the lines to user content.
		if line == separator {
			state = stateUserContent
			continue
		}

		switch state {
		// stateIntroLine - The first line is mentioning the story and contains the story link.
		case stateIntroLine:
			match := introLineRegexp.FindStringSubmatch(line)
			if len(match) != 3 {
				return nil, &ErrInvalidBody{issue, lineNo, line}
			}

			ctx.StoryId = match[1]
			ctx.StoryURL = match[2]

			state = stateIntroLineTrailingLine

		// stateIntroLineTrailingLine - the trailing line following the intro story line.
		case stateIntroLineTrailingLine:
			state = stateMetadata

		// stateMetadata - this section contains the metadata tags for SalsaFlow.
		case stateMetadata:
			match := trackerNameRegexp.FindStringSubmatch(line)
			if len(match) == 2 {
				ctx.TrackerName = match[1]
				continue
			}

			match = storyKeyRegexp.FindStringSubmatch(line)
			if len(match) == 2 {
				ctx.StoryKey = match[1]
				continue
			}

			// In case the line is empty, the metadata section ends.
			if line == "" {
				// Make sure the tags are filled.
				switch {
				case ctx.TrackerName == "":
					return nil, &ErrTagNotFound{issue, TagTrackerName}
				case ctx.StoryKey == "":
					return nil, &ErrTagNotFound{issue, TagStoryKey}
				}

				// Parse what is left.
				commonBody, err := parseRemainingIssueBody(issue, scanner, lineNo)
				if err != nil {
					return nil, err
				}
				ctx.ReviewIssueCommonBody = commonBody
				break Scanning
			}

		default:
			panic("unknown state")
		}
	}

	return ctx, nil
}
