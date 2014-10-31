package pivotaltracker

import (
	// Internal
	"github.com/salsita/salsaflow/errs"
	"github.com/salsita/salsaflow/log"
	"github.com/salsita/salsaflow/modules/common"
	"github.com/salsita/salsaflow/version"

	// Other
	"gopkg.in/salsita/go-pivotaltracker.v0/v5/pivotal"
)

type runningRelease struct {
	stories []*pivotal.Story
}

func newRunningRelease(ver *version.Version) (*runningRelease, error) {
	stories, err := listStoriesByRelease(ver)
	if err != nil {
		return nil, err
	}
	return &runningRelease{stories}, nil
}

func (release *runningRelease) Stories() ([]common.Story, error) {
	var stories []common.Story
	for _, s := range release.stories {
		stories = append(stories, &story{s})
	}
	return stories, nil
}

func (release *runningRelease) EnsureDeliverable() error {
	// Make sure that all stories are reviewed and QA'd.
	msg := "Make sure that the relevant stories are deliverable"
	log.Run(msg)
	ok, details := releaseDeliverable(release.stories)
	if !ok {
		log.FailWithDetails(msg, details)
		return ErrReleaseNotDeliverable
	}
	return nil
}

func (release *runningRelease) Deliver() (common.Action, error) {
	// Deliver the stories in Pivotal Tracker.
	msg := "Deliver the stories"
	log.Run(msg)
	stories, stderr, err := setStoriesState(release.stories, pivotal.StoryStateDelivered)
	if err != nil {
		errs.NewError(msg, err, stderr).Log(log.V(log.Info))
		return nil, err
	}
	release.stories = stories
	return common.ActionFunc(func() error {
		// On error, set the story state back to Finished.
		stories, stderr, err := setStoriesState(release.stories, pivotal.StoryStateFinished)
		if err != nil {
			errs.NewError(msg, err, stderr).Log(log.V(log.Info))
			return err
		}
		release.stories = stories
		return nil
	}), nil
}
