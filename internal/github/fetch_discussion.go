package github

import (
	"context"
	"encoding/json"
	"fmt"
)

func (f *fetcher) fetchDiscussion(ctx context.Context, ref ResourceRef, opts FetchOptions) (IssueData, error) {
	query := discussionPageQuery(opts.IncludeComments)

	var result IssueData
	result.Meta.Type = ResourceDiscussion

	err := f.gql.QueryPaginated(ctx, query, map[string]any{
		"owner":  ref.Owner,
		"repo":   ref.Repo,
		"number": ref.Number,
	}, func(page json.RawMessage) (bool, string, error) {
		var payload discussionPayload
		if err := json.Unmarshal(page, &payload); err != nil {
			return false, "", fmt.Errorf("decode discussion page payload: %w", err)
		}

		discussion := payload.Repository.Discussion
		if discussion == nil {
			return false, "", fmt.Errorf("discussion node missing: %w", ErrResourceNotFound)
		}

		if result.Meta.Title == "" {
			acceptedAnswerID := ""
			acceptedAnswerAuthor := ""
			if discussion.Answer != nil {
				acceptedAnswerID = discussion.Answer.ID
				if discussion.Answer.Author != nil {
					acceptedAnswerAuthor = discussion.Answer.Author.Login
				}
			}

			result.Meta = Metadata{
				Type:                 ResourceDiscussion,
				Title:                discussion.Title,
				Number:               discussion.Number,
				State:                discussionState(discussion.Closed),
				Author:               discussion.Author.Login,
				CreatedAt:            discussion.CreatedAt,
				UpdatedAt:            discussion.UpdatedAt,
				URL:                  discussion.URL,
				Category:             discussion.Category.Name,
				IsAnswered:           discussion.IsAnswered,
				AcceptedAnswerID:     acceptedAnswerID,
				AcceptedAnswerAuthor: acceptedAnswerAuthor,
			}
			result.Description = discussion.Body
			result.Reactions = mapGraphQLReactions(discussion.Reactions)
		}

		if !opts.IncludeComments {
			return false, "", nil
		}

		for _, node := range discussion.Comments.Nodes {
			commentNode, err := f.mapDiscussionComment(ctx, node)
			if err != nil {
				return false, "", fmt.Errorf("map discussion comment %q: %w", node.ID, err)
			}
			result.Thread = append(result.Thread, commentNode)
		}

		return discussion.Comments.PageInfo.HasNextPage, discussion.Comments.PageInfo.EndCursor, nil
	})
	if err != nil {
		return IssueData{}, fmt.Errorf("fetch discussion pages: %w", err)
	}

	return result, nil
}

func discussionPageQuery(includeComments bool) string {
	if includeComments {
		return `query DiscussionPage($owner:String!, $repo:String!, $number:Int!, $after:String) {
  repository(owner:$owner, name:$repo) {
    discussion(number:$number) {
      number
      title
      body
      url
      createdAt
      updatedAt
      closed
      author { login }
      category { name }
      isAnswered
      answer {
        id
        author { login }
      }
      reactions { plusOne heart total }
      comments(first:50, after:$after) {
        nodes {
          id
          body
          createdAt
          updatedAt
          url
          author { login }
          reactions { plusOne heart total }
          replies(first:50) {
            nodes {
              id
              body
              createdAt
              updatedAt
              url
              author { login }
              reactions { plusOne heart total }
            }
            pageInfo { hasNextPage endCursor }
          }
        }
        pageInfo { hasNextPage endCursor }
	      }
	    }
	  }
}`
	}
	return `query DiscussionPage($owner:String!, $repo:String!, $number:Int!) {
  repository(owner:$owner, name:$repo) {
    discussion(number:$number) {
      number
      title
      body
      url
      createdAt
      updatedAt
      closed
      author { login }
      category { name }
      isAnswered
      answer {
        id
        author { login }
      }
      reactions { plusOne heart total }
    }
  }
}`
}

func (f *fetcher) mapDiscussionComment(ctx context.Context, in discussionCommentPayload) (CommentNode, error) {
	out := CommentNode{
		ID:        in.ID,
		Author:    in.Author.Login,
		Body:      in.Body,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
		URL:       in.URL,
		Reactions: mapGraphQLReactions(in.Reactions),
	}

	for _, reply := range in.Replies.Nodes {
		out.Replies = append(out.Replies, mapDiscussionReply(reply))
	}

	if in.Replies.PageInfo.HasNextPage {
		extraReplies, err := f.fetchDiscussionReplies(ctx, in.ID, in.Replies.PageInfo.EndCursor)
		if err != nil {
			return CommentNode{}, fmt.Errorf("fetch additional discussion replies: %w", err)
		}
		out.Replies = append(out.Replies, extraReplies...)
	}

	return out, nil
}

func (f *fetcher) fetchDiscussionReplies(ctx context.Context, commentID, initialCursor string) ([]CommentNode, error) {
	query := `query DiscussionReplies($commentID:ID!, $after:String) {
  node(id:$commentID) {
    ... on DiscussionComment {
      replies(first:50, after:$after) {
        nodes {
          id
          body
          createdAt
          updatedAt
          url
          author { login }
          reactions { plusOne heart total }
        }
        pageInfo { hasNextPage endCursor }
      }
    }
  }
}`

	var out []CommentNode
	variables := map[string]any{
		"commentID": commentID,
		"after":     initialCursor,
	}
	err := f.gql.QueryPaginated(ctx, query, variables, func(page json.RawMessage) (bool, string, error) {
		var payload discussionRepliesPayload
		if err := json.Unmarshal(page, &payload); err != nil {
			return false, "", fmt.Errorf("decode discussion replies payload: %w", err)
		}
		if payload.Node == nil {
			return false, "", fmt.Errorf("discussion reply node missing for comment %q: %w", commentID, ErrResourceNotFound)
		}

		for _, reply := range payload.Node.Replies.Nodes {
			out = append(out, mapDiscussionReply(reply))
		}
		return payload.Node.Replies.PageInfo.HasNextPage, payload.Node.Replies.PageInfo.EndCursor, nil
	})
	if err != nil {
		return nil, fmt.Errorf("query discussion replies: %w", err)
	}
	return out, nil
}

type discussionPayload struct {
	Repository struct {
		Discussion *discussionNode `json:"discussion"`
	} `json:"repository"`
}

type discussionNode struct {
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	URL       string `json:"url"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Closed    bool   `json:"closed"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Category struct {
		Name string `json:"name"`
	} `json:"category"`
	IsAnswered bool `json:"isAnswered"`
	Answer     *struct {
		ID     string `json:"id"`
		Author *struct {
			Login string `json:"login"`
		} `json:"author"`
	} `json:"answer"`
	Reactions graphQLReactionSummary `json:"reactions"`
	Comments  struct {
		Nodes    []discussionCommentPayload `json:"nodes"`
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
	} `json:"comments"`
}

type discussionCommentPayload struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	URL       string `json:"url"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Reactions graphQLReactionSummary `json:"reactions"`
	Replies   struct {
		Nodes    []discussionReplyPayload `json:"nodes"`
		PageInfo struct {
			HasNextPage bool   `json:"hasNextPage"`
			EndCursor   string `json:"endCursor"`
		} `json:"pageInfo"`
	} `json:"replies"`
}

type discussionRepliesPayload struct {
	Node *struct {
		Replies struct {
			Nodes    []discussionReplyPayload `json:"nodes"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
		} `json:"replies"`
	} `json:"node"`
}

type discussionReplyPayload struct {
	ID        string `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	URL       string `json:"url"`
	Author    struct {
		Login string `json:"login"`
	} `json:"author"`
	Reactions graphQLReactionSummary `json:"reactions"`
}

type graphQLReactionSummary struct {
	PlusOne int `json:"plusOne"`
	Heart   int `json:"heart"`
	Total   int `json:"total"`
}

func mapDiscussionReply(in discussionReplyPayload) CommentNode {
	return CommentNode{
		ID:        in.ID,
		Author:    in.Author.Login,
		Body:      in.Body,
		CreatedAt: in.CreatedAt,
		UpdatedAt: in.UpdatedAt,
		URL:       in.URL,
		Reactions: mapGraphQLReactions(in.Reactions),
	}
}

func mapGraphQLReactions(in graphQLReactionSummary) ReactionSummary {
	return ReactionSummary{
		PlusOne: in.PlusOne,
		Heart:   in.Heart,
		Total:   in.Total,
	}
}

func discussionState(closed bool) string {
	if closed {
		return "closed"
	}
	return "open"
}
