package github

import (
	"context"
	"fmt"
	"strconv"

	goGithub "github.com/google/go-github/v72/github"
)

func (f *fetcher) fetchPullRequest(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error) {
	pr, err := f.rest.getPullRequest(ctx, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch pull request resource: %w", err)
	}

	issueForPR, err := f.rest.getIssue(ctx, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch pull request issue envelope: %w", err)
	}

	data := IssueData{
		Meta: Metadata{
			Type:        ResourcePullRequest,
			Title:       pr.GetTitle(),
			Number:      pr.GetNumber(),
			State:       pr.GetState(),
			Author:      pr.GetUser().GetLogin(),
			CreatedAt:   formatTimestamp(pr.CreatedAt),
			UpdatedAt:   formatTimestamp(pr.UpdatedAt),
			URL:         pr.GetHTMLURL(),
			Labels:      mapLabels(pr.Labels),
			Merged:      pr.GetMerged(),
			MergedAt:    formatTimestamp(pr.MergedAt),
			ReviewCount: pr.GetReviewComments(),
		},
		Description: pr.GetBody(),
		Reactions:   mapReactions(issueForPR.Reactions),
	}

	if !opts.IncludeComments {
		return data, nil
	}

	reviews, err := f.rest.listPullRequestReviews(ctx, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch pull request reviews: %w", err)
	}

	data.Reviews = mapPRReviews(reviews)
	reviewIDToIndex := make(map[int64]int, len(data.Reviews))
	for idx, review := range reviews {
		reviewIDToIndex[review.GetID()] = idx
	}

	comments, err := f.rest.listPullRequestComments(ctx, ref.Owner, ref.Repo, ref.Number)
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch pull request review comments: %w", err)
	}
	for _, comment := range comments {
		commentNode := mapPRComment(comment)
		if index, ok := reviewIDToIndex[comment.GetPullRequestReviewID()]; ok {
			data.Reviews[index].Comments = append(data.Reviews[index].Comments, commentNode)
			continue
		}
		// Keep unmatched review comments so discussion context is never dropped.
		data.Thread = append(data.Thread, commentNode)
	}

	return data, nil
}

func mapPRReviews(reviews []*goGithub.PullRequestReview) []ReviewData {
	out := make([]ReviewData, 0, len(reviews))
	for _, review := range reviews {
		out = append(out, ReviewData{
			ID:        strconv.FormatInt(review.GetID(), 10),
			State:     review.GetState(),
			Author:    review.GetUser().GetLogin(),
			Body:      review.GetBody(),
			CreatedAt: formatTimestamp(review.SubmittedAt),
			Reactions: ReactionSummary{},
		})
	}
	return out
}

func mapPRComment(comment *goGithub.PullRequestComment) CommentNode {
	return CommentNode{
		ID:        strconv.FormatInt(comment.GetID(), 10),
		Author:    comment.GetUser().GetLogin(),
		Body:      comment.GetBody(),
		CreatedAt: formatTimestamp(comment.CreatedAt),
		UpdatedAt: formatTimestamp(comment.UpdatedAt),
		URL:       comment.GetHTMLURL(),
		Reactions: mapReactions(comment.Reactions),
	}
}
