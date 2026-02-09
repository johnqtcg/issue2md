package github

import (
	"reflect"
	"testing"
)

func TestIssueDataContractFields(t *testing.T) {
	t.Parallel()

	required := []string{
		"Meta",
		"Description",
		"Reactions",
		"Timeline",
		"Reviews",
		"Thread",
	}

	assertStructHasFields(t, reflect.TypeOf(IssueData{}), required)
}

func TestMetadataContractFields(t *testing.T) {
	t.Parallel()

	required := []string{
		"Type",
		"Title",
		"Number",
		"State",
		"Author",
		"CreatedAt",
		"UpdatedAt",
		"URL",
		"Labels",
		"Merged",
		"MergedAt",
		"ReviewCount",
		"Category",
		"IsAnswered",
		"AcceptedAnswerID",
		"AcceptedAnswerAuthor",
	}

	assertStructHasFields(t, reflect.TypeOf(Metadata{}), required)
}

func TestReactionSummaryContractFields(t *testing.T) {
	t.Parallel()

	required := []string{
		"PlusOne",
		"MinusOne",
		"Laugh",
		"Hooray",
		"Confused",
		"Heart",
		"Rocket",
		"Eyes",
		"Total",
	}

	assertStructHasFields(t, reflect.TypeOf(ReactionSummary{}), required)
}

func assertStructHasFields(t *testing.T, typ reflect.Type, fields []string) {
	t.Helper()

	for _, field := range fields {
		if _, ok := typ.FieldByName(field); !ok {
			t.Fatalf("missing required field %q in %s", field, typ.Name())
		}
	}
}
