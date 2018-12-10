package backend

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
)

type preprocessor interface {
	preprocess(input []*container) ([]*container, error)
}

type iToI struct{}

var (
	digitRegexp = regexp.MustCompile("[0-9]+")
	keywords    = []string{
		"Close #",
		"Closes #",
		"Closed #",
		"Fix #",
		"Fixes #",
		"Fixed #",
		"Resolve #",
		"Resolves #",
		"Resolved #",
	}
)

func findIssueNumbers(s string, sub []string) []int {
	nums := []int{}
	for _, kw := range sub {
		idx := 0
		for idx != -1 {
			idx = strings.Index(s, kw)
			if idx != -1 {
				s = s[idx:]
				digit := digitRegexp.Find([]byte(s))
				id, _ := strconv.ParseInt(string(digit), 10, 20)
				nums = append(nums, int(id))
				s = s[len(kw):]
			}
		}
	}
	return nums
}

// preprocess provides links between issues and issues/pull requests.
func (itoi *iToI) preprocess(input []*container) ([]*container, error) {
	if len(input) == 0 {
		return nil, errors.New("empty input slice of preprocess.container")
	}

	closers := make(map[int][]*container)
	for i := range input {
		input[i].Linked = make(map[string][]*container)
		switch input[i].Event {
		case "issues":
			evt := &github.IssueEvent{}
			if err := json.Unmarshal(input[i].Payload, evt); err != nil {
				return nil, errors.Wrap(err, "unmarshalling issue event")
			}
			input[i].Issue = evt.Issue
		case "pull_request":
			evt := &github.PullRequestEvent{}
			if err := json.Unmarshal(input[i].Payload, evt); err != nil {
				return nil, errors.Wrap(err, "unmarshalling pull request event")
			}
			input[i].PullRequest = evt.PullRequest

			if evt.PullRequest.Body == nil {
				continue
			}

			text := *evt.PullRequest.Title + " " + *evt.PullRequest.Body
			nums := findIssueNumbers(text, keywords)
			for _, n := range nums {
				closers[n] = append(closers[n], input[i])
			}
		}
	}

	for i := range input {
		if input[i].Issue != nil {
			if linked, ok := closers[*input[i].Issue.Number]; ok {
				input[i].Linked["pull_request"] = linked
			}
		}
	}

	return input, nil
}
