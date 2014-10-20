package jira

import (
	"errors"
	"net/http"
)

type jiraWorkflowEntity struct {
	Id   string
	Name string
}

var startableStates = []jiraWorkflowEntity{
	{
		"10107",
		"Scheduled",
	},
}

var statusIcebox = jiraWorkflowEntity{
	"10000",
	"IceBox",
}

var statusScheduled = jiraWorkflowEntity{
	"10107",
	"Scheduled",
}

var statusInProgress = jiraWorkflowEntity{
	"3",
	"In Progress",
}

var statusVerification = jiraWorkflowEntity{
	"10104",
	"Verification",
}

var statusStaged = jiraWorkflowEntity{
	"10105",
	"Staged",
}

var statusReleased = jiraWorkflowEntity{
	"10106",
	"Released",
}

var statusClosed = jiraWorkflowEntity{
	"6",
	"Closed",
}

var transitionStart = jiraWorkflowEntity{
	"4",
	"Start Progress",
}

var transitionStage = jiraWorkflowEntity{
	"731",
	"stage",
}

var transitionRelease = jiraWorkflowEntity{
	"741",
	"release",
}

var transitions = []jiraWorkflowEntity{
	{
		"2",
		"Close Issue",
	},
	{
		"721",
		"Implemented",
	},
	{
		"751",
		"schedule",
	},
	{
		"761",
		"iceicebaby",
	},
	{
		"4",
		"Start Progress",
	},
	{
		"731",
		"stage",
	},
	{
		"741",
		"release",
	},
}

// srcState dstState -> transition ID
var transitionMachine = map[string]map[string]string{
	"6": {}, // Haven't found any transitions from Closed.
	"10000": {
		"6":     "2",   // Icebox -> Closed
		"10107": "751", // Icebox -> Scheduled
	},
	"10107": {
		"3":     "4",   // Scheduled -> In Progress
		"6":     "2",   // Scheduled -> Closed
		"10000": "761", // Scheduled -> Icebox
	},
	"3": {
		"6":     "2",   // In Progress -> Closed
		"10104": "721", // In Progress -> Verification
	},
	"10104": {
		"3":     "4",   // Verification -> In Progress
		"6":     "2",   // Verification -> Closed
		"10105": "731", // Verification -> Staged
	},
	"10105": {
		"3":     "4",   // Staged -> In Progress
		"10106": "741", // Staged -> Released
	},
	"10106": {}, // No transitions from Released.
}

func transitionTo(issueIdOrKey, srcState, dstState string) (*http.Response, error) {
	transitionId := transitionMachine[srcState][dstState]
	if transitionId == "" {
		return nil, errors.New("invalid transition")
	}

	return newClient().Issues.PerformTransition(issueIdOrKey, transitionId)
}
