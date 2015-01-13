package modules

import (
	// Stdlib
	"bytes"
	"fmt"

	// Internal
	"github.com/salsaflow/salsaflow/errs"
	"github.com/salsaflow/salsaflow/modules/common"

	// Internal: modules
	"github.com/salsaflow/salsaflow/modules/jira"
	"github.com/salsaflow/salsaflow/modules/pivotaltracker"
	"github.com/salsaflow/salsaflow/modules/sprintly"
)

// IssueTracker instantiation --------------------------------------------------

type IssueTrackerFactory func() (common.IssueTracker, error)

var issueTrackerFactories = map[string]IssueTrackerFactory{
	jira.Id:           jira.Factory,
	pivotaltracker.Id: pivotaltracker.Factory,
	sprintly.Id:       sprintly.Factory,
}

func AvailableIssueTrackerKeys() []string {
	keys := make([]string, 0, len(issueTrackerFactories))
	for key := range issueTrackerFactories {
		keys = append(keys, key)
	}
	return keys
}

func GetIssueTracker() (common.IssueTracker, error) {
	// Load configuration.
	config, err := common.LoadConfig()
	if err != nil && config == nil {
		return nil, err
	}

	// Choose the issue tracker based on the configuration.
	var task = "Instantiate the selected issue tracker plugin"
	id := config.IssueTrackerId()
	factory, ok := issueTrackerFactories[id]
	if !ok {
		// Collect the available tracker ids.
		ids := make([]string, 0, len(issueTrackerFactories))
		for id := range issueTrackerFactories {
			ids = append(ids, id)
		}

		hint := new(bytes.Buffer)
		fmt.Fprintf(hint, "\nAvailable issue trackers: %v\n\n", ids)
		return nil, errs.NewError(task, fmt.Errorf("unknown issue tracker: '%v'", id), hint)
	}

	// Try to instantiate the issue tracker.
	tracker, err := factory()
	if err != nil {
		return nil, errs.NewError(task, err, nil)
	}

	return tracker, nil
}
