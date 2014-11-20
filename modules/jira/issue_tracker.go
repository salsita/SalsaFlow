package jira

import (
	// Stdlib
	"fmt"
	"net/url"

	// Internal
	"github.com/salsita/salsaflow/log"
	"github.com/salsita/salsaflow/modules/common"
	"github.com/salsita/salsaflow/modules/jira/client"
	"github.com/salsita/salsaflow/version"

	// Other
	"github.com/toqueteos/webbrowser"
)

type issueTracker struct {
	config       Config
	versionCache map[string]*client.Version
}

func Factory() (common.IssueTracker, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return &issueTracker{config, nil}, nil
}

func (tracker *issueTracker) CurrentUser() (common.User, error) {
	data, _, err := newClient(tracker.config).Myself.Get()
	if err != nil {
		return nil, err
	}
	return &user{data}, nil
}

func (tracker *issueTracker) StartableStories() (stories []common.Story, err error) {
	query := fmt.Sprintf("project=%v AND (%v) AND (%v)", tracker.config.ProjectKey(),
		formatInRange("type", codingIssueTypeIds...),
		formatInRange("status", startableStateIds...))

	issues, _, err := newClient(tracker.config).Issues.Search(&client.SearchOptions{
		JQL:        query,
		MaxResults: 200,
	})
	if err != nil {
		return nil, err
	}

	return toCommonStories(issues, tracker.config), nil
}

func (tracker *issueTracker) StoriesInDevelopment() (stories []common.Story, err error) {
	query := fmt.Sprintf("project=%v AND (%v) AND (%v)", tracker.config.ProjectKey(),
		formatInRange("type", codingIssueTypeIds...),
		formatInRange("status", inDevelopmentStateIds...))

	issues, _, err := newClient(tracker.config).Issues.Search(&client.SearchOptions{
		JQL:        query,
		MaxResults: 200,
	})
	if err != nil {
		return nil, err
	}

	return toCommonStories(issues, tracker.config), nil
}

func (tracker *issueTracker) NextRelease(
	trunkVersion *version.Version,
	nextTrunkVersion *version.Version,
) (common.NextRelease, error) {

	return newNextRelease(tracker, trunkVersion, nextTrunkVersion)
}

func (tracker *issueTracker) RunningRelease(
	releaseVersion *version.Version,
) (common.RunningRelease, error) {

	return newRunningRelease(tracker, releaseVersion)
}

func (tracker *issueTracker) ReleaseStoriesNotAccepted(ver *version.Version) ([]common.Story, error) {
	// Make sure the relevant JIRA version exists.
	// This is necessary to do since JIRA returns 400 Bad Request when
	// JQL with 'fixVersion = "non-existing version"' is sent.
	res, err := tracker.getVersionResource(ver)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, common.ErrReleaseNotFound
	}

	// Get the issues.
	var (
		key = tracker.config.ProjectKey()
		tag = ver.ReleaseTagString()
	)
	issues, err := issuesByVersion(newClient(tracker.config), key, tag)
	if err != nil {
		return nil, err
	}

	// Drop the accepted issues.
	var notAccepted []*client.Issue
IssueLoop:
	for _, issue := range issues {
		for _, id := range acceptedStateIds {
			if id == issue.Fields.Status.Id {
				continue IssueLoop
			}
		}
		notAccepted = append(notAccepted, issue)
	}

	// Return what is left.
	return toCommonStories(notAccepted, tracker.config), nil
}

func (tracker *issueTracker) SelectActiveStoryIds(ids []string) (activeIds []string, err error) {
	info := log.V(log.Info)

	// Fetch the relevant issues
	task := "Fetch the relevant issues"
	info.Run(task)

	issues, err := listStoriesById(newClient(tracker.config), ids)
	if err != nil {
		return nil, err
	}

	issueMap := make(map[string]*client.Issue)
	for _, issue := range issues {
		issueMap[issue.Id] = issue
	}

	// Filter the issues according to the issue state.
	task = "Filter the issues according to the issue state"
	var active []string
	for _, id := range ids {
		_, ok := issueMap[id]
		if !ok {
			info.Fail(task)
			err = fmt.Errorf("issue with id %v not found", id)
			return nil, err
		}

		// XXX: Implement this!
		panic("Not implemented")
	}

	return active, nil
}

func (tracker *issueTracker) OpenStory(storyId string) error {
	relativeURL, _ := url.Parse("browse/" + storyId)
	return webbrowser.Open(tracker.config.BaseURL().ResolveReference(relativeURL).String())
}

func (tracker *issueTracker) getVersionResource(ver *version.Version) (*client.Version, error) {
	var (
		projectKey  = tracker.config.ProjectKey()
		versionName = ver.ReleaseTagString()
		api         = newClient(tracker.config)
	)

	// In case the resource cache is empty, fill it.
	if tracker.versionCache == nil {
		vs, _, err := api.Projects.ListVersions(projectKey)
		if err != nil {
			return nil, err
		}

		m := make(map[string]*client.Version, len(vs))
		for _, v := range vs {
			m[v.Name] = v
		}
		tracker.versionCache = m
	}

	// Return the resource we are looking for.
	if res, ok := tracker.versionCache[versionName]; ok {
		return res, nil
	}
	return nil, nil
}

func toCommonStories(issues []*client.Issue, config Config) []common.Story {
	api := newClient(config)
	stories := make([]common.Story, len(issues))
	for i := range issues {
		stories[i] = &story{issues[i], api}
	}
	return stories
}
