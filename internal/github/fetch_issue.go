package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	goGithub "github.com/google/go-github/v72/github"
)

func (f *fetcher) fetchIssue(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error) {
	issue, err := f.rest.getIssue(ctx, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch issue resource: %w", err)
	}

	data := IssueData{
		Meta: Metadata{
			Type:      ResourceIssue,
			Title:     issue.GetTitle(),
			Number:    issue.GetNumber(),
			State:     issue.GetState(),
			Author:    issue.GetUser().GetLogin(),
			CreatedAt: formatTimestamp(issue.CreatedAt),
			UpdatedAt: formatTimestamp(issue.UpdatedAt),
			URL:       issue.GetHTMLURL(),
			Labels:    mapLabels(issue.Labels),
		},
		Description: issue.GetBody(),
		Reactions:   mapReactions(issue.Reactions),
	}

	data.Timeline = append(data.Timeline, TimelineEvent{
		EventType: "opened",
		Actor:     data.Meta.Author,
		CreatedAt: data.Meta.CreatedAt,
	})

	timeline, err := f.fetchIssueTimeline(ctx, ref)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch issue timeline: %w", err)
	}
	data.Timeline = append(data.Timeline, timeline...)
	data.Timeline = dedupeTimelineEvents(data.Timeline)

	if opts.IncludeComments {
		comments, err := f.rest.listIssueComments(ctx, ref.Owner, ref.Repo, ref.Number)
		if err != nil {
			return IssueData{}, fmt.Errorf("fetch issue comments: %w", err)
		}
		data.Thread = mapIssueComments(comments)
	}

	return data, nil
}

type issueTimelinePayload struct {
	Repository struct {
		Issue *struct {
			TimelineItems struct {
				Nodes    []issueTimelineNode `json:"nodes"`
				PageInfo struct {
					HasNextPage bool   `json:"hasNextPage"`
					EndCursor   string `json:"endCursor"`
				} `json:"pageInfo"`
			} `json:"timelineItems"`
		} `json:"issue"`
	} `json:"repository"`
}

type issueTimelineNode struct {
	TypeName string `json:"__typename"`
	Created  string `json:"createdAt"`
	Actor    struct {
		Login string `json:"login"`
	} `json:"actor"`
	Label struct {
		Name string `json:"name"`
	} `json:"label"`
	Assignee struct {
		Login string `json:"login"`
	} `json:"assignee"`
	Milestone struct {
		Title string `json:"title"`
	} `json:"milestone"`
	MilestoneTitle string `json:"milestoneTitle"`
}

func (f *fetcher) fetchIssueTimeline(ctx context.Context, ref ResourceRef) ([]TimelineEvent, error) {
	query := `query IssueTimeline($owner:String!, $repo:String!, $number:Int!, $after:String) {
  repository(owner:$owner, name:$repo) {
    issue(number:$number) {
      timelineItems(first:100, after:$after) {
        nodes {
          __typename
          ... on ClosedEvent {
            createdAt
            actor { login }
          }
          ... on ReopenedEvent {
            createdAt
            actor { login }
          }
          ... on LabeledEvent {
            createdAt
            actor { login }
            label { name }
          }
          ... on AssignedEvent {
            createdAt
            actor { login }
            assignee { __typename }
          }
          ... on MilestonedEvent {
            createdAt
            actor { login }
            milestoneTitle
          }
          ... on LockedEvent {
            createdAt
            actor { login }
          }
        }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`

	var events []TimelineEvent
	err := f.gql.QueryPaginated(ctx, query, map[string]any{
		"owner":  ref.Owner,
		"repo":   ref.Repo,
		"number": ref.Number,
	}, func(page json.RawMessage) (bool, string, error) {
		var payload issueTimelinePayload
		if err := json.Unmarshal(page, &payload); err != nil {
			return false, "", fmt.Errorf("decode issue timeline page payload: %w", err)
		}
		if payload.Repository.Issue == nil {
			return false, "", fmt.Errorf("issue timeline missing issue node: %w", ErrResourceNotFound)
		}

		for _, node := range payload.Repository.Issue.TimelineItems.Nodes {
			eventType, details, ok := mapIssueTimelineNode(node)
			if !ok {
				continue
			}
			events = append(events, TimelineEvent{
				EventType: eventType,
				Actor:     node.Actor.Login,
				CreatedAt: node.Created,
				Details:   details,
			})
		}
		return payload.Repository.Issue.TimelineItems.PageInfo.HasNextPage, payload.Repository.Issue.TimelineItems.PageInfo.EndCursor, nil
	})
	if err != nil {
		return nil, fmt.Errorf("query timeline items: %w", err)
	}

	return events, nil
}

func mapIssueTimelineNode(node issueTimelineNode) (eventType, details string, ok bool) {
	switch node.TypeName {
	case "OpenedEvent":
		return "opened", "", true
	case "ClosedEvent":
		return "closed", "", true
	case "ReopenedEvent":
		return "reopened", "", true
	case "LabeledEvent":
		return "labeled", node.Label.Name, true
	case "AssignedEvent":
		if node.Assignee.Login != "" {
			return "assigned", node.Assignee.Login, true
		}
		return "assigned", node.Actor.Login, true
	case "MilestonedEvent":
		if node.MilestoneTitle != "" {
			return "milestoned", node.MilestoneTitle, true
		}
		return "milestoned", node.Milestone.Title, true
	case "LockedEvent":
		return "locked", "", true
	default:
		return "", "", false
	}
}

func dedupeTimelineEvents(events []TimelineEvent) []TimelineEvent {
	type key TimelineEvent

	out := make([]TimelineEvent, 0, len(events))
	seen := make(map[key]struct{}, len(events))
	for _, event := range events {
		k := key(event)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, event)
	}
	return out
}

func mapIssueComments(comments []*goGithub.IssueComment) []CommentNode {
	nodes := make([]CommentNode, 0, len(comments))
	for _, comment := range comments {
		nodes = append(nodes, CommentNode{
			ID:        strconv.FormatInt(comment.GetID(), 10),
			Author:    comment.GetUser().GetLogin(),
			Body:      comment.GetBody(),
			CreatedAt: formatTimestamp(comment.CreatedAt),
			UpdatedAt: formatTimestamp(comment.UpdatedAt),
			URL:       comment.GetHTMLURL(),
			Reactions: mapReactions(comment.Reactions),
		})
	}
	return nodes
}

func mapLabels(labels []*goGithub.Label) []Label {
	result := make([]Label, 0, len(labels))
	for _, label := range labels {
		result = append(result, Label{Name: label.GetName()})
	}
	return result
}

func mapReactions(reactions *goGithub.Reactions) ReactionSummary {
	if reactions == nil {
		return ReactionSummary{}
	}
	return ReactionSummary{
		PlusOne:  reactions.GetPlusOne(),
		MinusOne: reactions.GetMinusOne(),
		Laugh:    reactions.GetLaugh(),
		Hooray:   reactions.GetHooray(),
		Confused: reactions.GetConfused(),
		Heart:    reactions.GetHeart(),
		Rocket:   reactions.GetRocket(),
		Eyes:     reactions.GetEyes(),
		Total:    reactions.GetTotalCount(),
	}
}

func formatTimestamp(ts *goGithub.Timestamp) string {
	if ts == nil {
		return ""
	}
	t := ts.GetTime()
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z07:00")
}
