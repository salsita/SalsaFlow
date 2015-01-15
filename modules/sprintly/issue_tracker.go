package sprintly

import (
	// Stdlib
	"fmt"
	"strings"

	// Internal
	"github.com/salsaflow/salsaflow/errs"
	"github.com/salsaflow/salsaflow/modules/common"
	"github.com/salsaflow/salsaflow/version"

	// Other
	"github.com/salsita/go-sprintly/sprintly"
	"github.com/toqueteos/webbrowser"
)

type issueTracker struct {
	config Config
	client *sprintly.Client
}

func Factory() (common.IssueTracker, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}

	client := sprintly.NewClient(config.Username(), config.Token())
	return &issueTracker{config, client}, nil
}

func (tracker *issueTracker) CurrentUser() (common.User, error) {
	task := "Fetch the current user from Sprintly"

	var (
		productId = tracker.config.ProductId()
		username  = tracker.config.Username()
	)

	// Fetch all members of this product.
	users, _, err := tracker.client.People.List(productId)
	if err != nil {
		return nil, errs.NewError(task, err, nil)
	}

	// Find the user with matching username.
	for _, usr := range users {
		if usr.Email == username {
			return &user{&usr}, nil
		}
	}

	// In case there is no such user, they were not invited yet.
	return nil, errs.NewError(
		task, fmt.Errorf("user '%v' not a member of this product", username), nil)
}

func (tracker *issueTracker) StartableStories() (stories []common.Story, err error) {
	task := "Fetch the startable items from Sprintly"

	var (
		productId = tracker.config.ProductId()
		username  = tracker.config.Username()
	)

	// Fetch the items from Sprintly.
	items, err := listItems(tracker.client, productId, &sprintly.ItemListArgs{
		Status: []sprintly.ItemStatus{sprintly.ItemStatusBacklog},
	})
	if err != nil {
		return nil, errs.NewError(task, err, nil)
	}

	// Drop the items that were already assigned.
	// However, keep the items that are assigned to the current user.
	for i, item := range items {
		if item.AssignedTo == nil {
			continue
		}
		if item.AssignedTo.Email == username {
			continue
		}
		items = append(items[:i], items[:i+1]...)
	}

	// Wrap the result as []common.Story
	return toCommonStories(items), nil
}

func (tracker *issueTracker) StoriesInDevelopment() (stories []common.Story, err error) {
	task := "Fetch the items that are in progress"

	// Fetch all items that are in progress.
	productId := tracker.config.ProductId()
	items, err := listItems(tracker.client, productId, &sprintly.ItemListArgs{
		Status: []sprintly.ItemStatus{sprintly.ItemStatusInProgress},
	})
	if err != nil {
		return nil, errs.NewError(task, err, nil)
	}

	// Drop the items that are not assigned to the current user.
	// We could do that in the query already, but we need user ID for that,
	// which would require another remote call. However, we know the username
	// since it is saved in the configuration file, so we can filter locally
	// based on that information.
	//
	// Also drom the items that are tagged as reviewed or no review.
	// That means that the coding phase is finished.
	var (
		inDevelopment = make([]sprintly.Item, 0, len(items))
		username      = tracker.config.Username()
		reviewedTag   = tracker.config.ReviewedTag()
		noReviewTag   = tracker.config.NoReviewTag()
	)
	for _, item := range items {
		if item.AssignedTo.Email != username {
			continue
		}
		if tagged(&item, reviewedTag) || tagged(&item, noReviewTag) {
			continue
		}
		inDevelopment = append(inDevelopment, item)
	}

	// Convert the items into []common.Story
	return toCommonStories(inDevelopment), nil
}

func (tracker *issueTracker) NextRelease(
	trunkVersion *version.Version,
	nextTrunkVersion *version.Version,
) (common.NextRelease, error) {

	return &nextRelease{
		client:           tracker.client,
		config:           tracker.config,
		trunkVersion:     trunkVersion,
		nextTrunkVersion: nextTrunkVersion,
	}, nil
}

func (tracker *issueTracker) RunningRelease(
	releaseVersion *version.Version,
) (common.RunningRelease, error) {

	return &runningRelease{
		client:  tracker.client,
		config:  tracker.config,
		version: releaseVersion,
	}, nil
}

func (tracker *issueTracker) OpenStory(storyId string) error {
	productId := tracker.config.ProductId()
	return webbrowser.Open(fmt.Sprintf("https://sprint.ly/product/%v/item/%v", productId, storyId))
}

func (tracker *issueTracker) StoryTagToReadableStoryId(tag string) (storyId string, err error) {
	task := "Parse Story-Id tag"
	parts := strings.Split(tag, "/")
	if len(parts) != 2 {
		return "", errs.NewError(task, fmt.Errorf("invalid Story-Id tag: %v", tag), nil)
	}
	return parts[1], nil
}
