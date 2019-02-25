package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	gh "github.com/google/go-github/v24/github"
	"golang.org/x/oauth2"
)

type ActivityType int

const (
	NonInteresting ActivityType = iota
	OpenedIssue
	ClosedIssue
	CommentedIssue
	OpenedPullRequest
	ReopenedPullRequest
	EditedPullRequest
	ClosedPullRequest
	MergedPullRequest
	CommentedPullRequest
)

func (a ActivityType) String() string {
	return []string{
		"non interesting",
		"opened issue",
		"closed issue",
		"commented issue",
		"opened pull request",
		"reopened pull request",
		"edited pull request",
		"closed pull request",
		"merged pull request",
		"commented pull request",
	}[a]
}

type Client struct {
	User     string
	ghClient *gh.Client
}

func NewClient(user string) Client {
	return newClient(user, nil)
}

func NewAuthClient(user string, accessToken string) Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)

	return newClient(user, tc)
}

func newClient(user string, httpClient *http.Client) Client {
	client := gh.NewClient(httpClient)
	return Client{User: user, ghClient: client}
}

type Activity struct {
	ByRepo           map[string]map[int][]*ActivityEvent
	subjectIDNameMap map[int]string
}

func (a Activity) String() string {
	var str strings.Builder
	for repo, activityBySubj := range a.ByRepo {
		str.WriteString(fmt.Sprintf("- %s:\n", repo))
		for subID, activityEvents := range activityBySubj {
			str.WriteString(fmt.Sprintf("  * (#%d) %s:\n", subID, a.subjectIDNameMap[subID]))
			for i := len(activityEvents) - 1; i >= 0; i-- {
				ae := activityEvents[i]
				str.WriteString(fmt.Sprintf("    -> %s [%v]\n", ae.Type, ae.CreatedAt))
			}
		}
	}

	return str.String()
}

type ActivityEvent struct {
	Type      ActivityType
	RepoName  string
	SubjectID int
	Subject   string
	CreatedAt time.Time
}

func (ae ActivityEvent) String() string {
	return fmt.Sprintf("%s [%s] (#%d) %s", ae.Type, ae.RepoName, ae.SubjectID, ae.Subject)
}

func (c *Client) GetActivity(ctx context.Context, publicOnly bool, since, to *time.Time) *Activity {
	// TODO:
	// - honour `since` and `to` params
	nextPage := -1
	byRepo := make(map[string]map[int][]*ActivityEvent)
	subjectIDNameMap := make(map[int]string)

	for nextPage != 0 {
		events, resp, err := c.getActivity(ctx, publicOnly, nextPage)
		if err != nil {
			panic(err)
		}

		nextPage = resp.NextPage
		for _, ev := range events {
			ae := c.rawEventToActivity(ev)
			if ae == nil {
				continue
			}

			subjectIDNameMap[ae.SubjectID] = ae.Subject

			if _, ok := byRepo[ae.RepoName]; !ok {
				byRepo[ae.RepoName] = make(map[int][]*ActivityEvent)
			}

			if _, ok := byRepo[ae.RepoName][ae.SubjectID]; !ok {
				byRepo[ae.RepoName][ae.SubjectID] = make([]*ActivityEvent, 0)
			}

			byRepo[ae.RepoName][ae.SubjectID] = append(byRepo[ae.RepoName][ae.SubjectID], ae)
		}
	}

	return &Activity{
		ByRepo:           byRepo,
		subjectIDNameMap: subjectIDNameMap,
	}
}

func (c *Client) getActivity(ctx context.Context, publicOnly bool, page int) ([]*gh.Event, *gh.Response, error) {
	var listOpts *gh.ListOptions
	if page > 0 {
		listOpts = &gh.ListOptions{Page: page}
	}

	events, resp, err := c.ghClient.Activity.ListEventsPerformedByUser(
		ctx, c.User, publicOnly, listOpts)

	return events, resp, err
}

func (c *Client) rawEventToActivity(ev *gh.Event) *ActivityEvent {
	payload, err := ev.ParsePayload()
	if err != nil {
		panic(err)
	}

	var activityType ActivityType
	var subject string
	var subjectID int

	if parsedEv, ok := payload.(*gh.IssuesEvent); ok {
		subject = *parsedEv.Issue.Title
		subjectID = *parsedEv.Issue.Number
		switch *parsedEv.Action {
		case "opened":
			activityType = OpenedIssue
		case "closed":
			activityType = ClosedIssue
		default:
			activityType = NonInteresting
		}
	}

	if parsedEv, ok := payload.(*gh.PullRequestEvent); ok {
		subject = *parsedEv.PullRequest.Title
		subjectID = *parsedEv.PullRequest.Number
		switch *parsedEv.Action {
		case "opened":
			activityType = OpenedPullRequest
		case "reopened":
			activityType = ReopenedPullRequest
		case "edited":
			activityType = EditedPullRequest
		case "closed":
			if *parsedEv.PullRequest.Merged {
				activityType = MergedPullRequest
			} else {
				activityType = ClosedPullRequest
			}
		default:
			activityType = NonInteresting
		}
	}

	if parsedEv, ok := payload.(*gh.IssueCommentEvent); ok {
		subject = *parsedEv.Issue.Title
		subjectID = *parsedEv.Issue.Number
		activityType = CommentedIssue
	}

	if parsedEv, ok := payload.(*gh.PullRequestReviewCommentEvent); ok {
		subject = *parsedEv.PullRequest.Title
		subjectID = *parsedEv.PullRequest.Number
		activityType = CommentedPullRequest
	}

	if activityType == NonInteresting {
		return nil
	}

	return &ActivityEvent{
		Type:      activityType,
		RepoName:  *ev.Repo.Name,
		SubjectID: subjectID,
		Subject:   subject,
		CreatedAt: *ev.CreatedAt,
	}
}
