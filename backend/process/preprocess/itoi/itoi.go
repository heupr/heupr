package itoi

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"

	"heupr/backend/process/preprocess"
)

// P is the exported struct needed to implement preprocess.Preprocessor.
type P struct{}

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

// Preprocess provides links between github.Issue structs and github.Issue and
// github.PullRequest structs based on references within titles/bodies.
func (p *P) Preprocess(input []*preprocess.Container) ([]*preprocess.Container, error) {
	if len(input) == 0 {
		return nil, errors.New("empty input slice of preprocess.Container")
	}

	closers := make(map[int][]*preprocess.Container)
	_ = closers // TEMPORARY
	for i := range input {
		switch input[i].Event {
		case "issues":
			evt := &github.IssueEvent{}
			if err := json.Unmarshal(input[i].Payload, evt); err != nil {
				return nil, errors.Wrap(err, "unmarshalling issue event")
			}
			input[i].Issue = evt.Issue

			// [ ] check to make sure there is an issue body
			// - [ ] if not continue w/o analysis
			// - [ ] if yes pull out title + body texts
			// [ ] see if linking values included
			// - [ ] if no do nothing
			// - [ ] if yes

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
			_ = nums

			// [ ] check text for issue references
			// - [ ] if yes add to temp closers
			// - [ ] if no continue w/o adding to closers
			// [ ]

		}
	}

	// NOTE: Actually pulling out the text objects (title/body) and appending
	// them together is only done within the individual model implementations
	// of Learn/Act.

	return input, nil
}
