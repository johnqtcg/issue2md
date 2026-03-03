package github

// ResourceType identifies the GitHub resource kind represented by a URL.
type ResourceType string

const (
	// ResourceIssue is the type for GitHub issue URLs.
	ResourceIssue ResourceType = "issue"
	// ResourcePullRequest is the type for GitHub pull request URLs.
	ResourcePullRequest ResourceType = "pull_request"
	// ResourceDiscussion is the type for GitHub discussion URLs.
	ResourceDiscussion ResourceType = "discussion"
)

// ResourceRef is the normalized resource identity extracted from an input URL.
type ResourceRef struct {
	Owner  string
	Repo   string
	Type   ResourceType
	URL    string
	Number int
}

// ReactionSummary captures aggregate reaction counts on a GitHub entity.
type ReactionSummary struct {
	PlusOne  int
	MinusOne int
	Laugh    int
	Hooray   int
	Confused int
	Heart    int
	Rocket   int
	Eyes     int
	Total    int
}

// Label stores minimal label data needed for output rendering.
type Label struct {
	Name string
}

// Metadata stores top-level fields used in front matter and metadata sections.
type Metadata struct {
	Category             string
	MergedAt             string
	AcceptedAnswerAuthor string
	State                string
	Author               string
	CreatedAt            string
	Title                string
	AcceptedAnswerID     string
	Type                 ResourceType
	URL                  string
	UpdatedAt            string
	Labels               []Label
	ReviewCount          int
	Number               int
	IsAnswered           bool
	Merged               bool
}

// TimelineEvent represents one normalized timeline event.
type TimelineEvent struct {
	EventType string
	Actor     string
	CreatedAt string
	Details   string
}

// CommentNode represents one comment and its nested replies.
type CommentNode struct {
	ID        string
	Author    string
	Body      string
	CreatedAt string
	UpdatedAt string
	URL       string
	Replies   []CommentNode
	Reactions ReactionSummary
}

// ReviewData stores review summary data and review-thread comments.
type ReviewData struct {
	ID        string
	State     string
	Author    string
	Body      string
	CreatedAt string
	Comments  []CommentNode
	Reactions ReactionSummary
}

// IssueData is the normalized transport payload consumed by other layers.
type IssueData struct {
	Description string
	Timeline    []TimelineEvent
	Reviews     []ReviewData
	Thread      []CommentNode
	Meta        Metadata
	Reactions   ReactionSummary
}
