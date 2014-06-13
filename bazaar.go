package vcs

import (
	"fmt"
	"strings"
	"time"
)

const (
	beginBzrCommit     = "\n------------------------------------------------------------\n"
	bzrSeparator       = "------------------------------------------------------------"
	bzrEoc             = "Use --include-merged or -n0 to see merged revisions."
	bzrTimestampLayout = "Mon 2006-01-02 15:04:05 -0700"
)

type bazaar struct {
}

func (b *bazaar) Cmd() string {
	return "bzr"
}

func (b *bazaar) Dir() string {
	return ".bzr"
}

func (b *bazaar) Head() string {
	return "-1"
}

func (b *bazaar) Revision(id string) []string {
	return []string{"log", "-p", "-r", id}
}

func (b *bazaar) History(since string) []string {
	args := []string{"log", "-p"}
	if since != "" {
		args = append(args, "-r")
		args = append(args, since+"..")
	}
	return args
}

func (b *bazaar) Checkout(rev string) []string {
	if rev == "" {
		rev = "-1"
	}
	return []string{"revert", "-r", rev}
}

func (b *bazaar) Clone(src string, dst string) []string {
	return []string{"branch", src, dst, "--use-existing-dir"}
}

func (b *bazaar) Update() []string {
	return []string{"pull", "--overwrite"}
}

func (b *bazaar) ParseRevisions(since string, data []byte) ([]*Revision, error) {
	commits := strings.Split(string(data), beginBzrCommit)
	revs, err := b.parseCommits(commits)
	// Bazaar always includes the since revision
	if err == nil && since != "" {
		revs = revs[:len(revs)-1]
	}
	return revs, err
}

func (b *bazaar) parseCommits(commits []string) ([]*Revision, error) {
	const (
		stateOther = iota
		stateMsg
		stateDiff
	)
	revs := make([]*Revision, 0, len(commits))
	for _, v := range commits {
		if v == "" {
			continue
		}
		var r Revision
		var state int
		var msg []string
		var diff []string
		for _, line := range strings.Split(v, "\n") {
			switch {
			case strings.HasPrefix(line, "revno:"):
				r.Identifier = strings.TrimSuffix(line[7:], " [merge]")
				r.ShortIdentifier = r.Identifier
			case strings.HasPrefix(line, "committer:"):
				r.Author, r.Email = splitAuthor(line[11:])
			case strings.HasPrefix(line, "timestamp:"):
				var err error
				r.Timestamp, err = parseBzrTimestamp(line[11:])
				if err != nil {
					return nil, err
				}
			case strings.HasPrefix(line, "message:"):
				state = stateMsg
			case strings.HasPrefix(line, "diff:"):
				state = stateDiff
			case strings.HasPrefix(line, "branch nick:"):
			case strings.HasPrefix(line, "tags:"):
			case strings.HasPrefix(line, "author:"):
			case strings.HasPrefix(line, "fixes bug:"):
				// Do nothing for now. Eventually we'll parse
				// branches, tags and authors.
			default:
				switch state {
				case stateMsg:
					msg = append(msg, line)
				case stateDiff:
					diff = append(diff, line)
				default:
					if line != "" && line != bzrSeparator && line != bzrEoc {
						return nil, fmt.Errorf("unknown line %q in commit %q", line, v)
					}
				}
			}
		}
		if r.Identifier != "" {
			r.Subject, r.Message = splitMessage(msg)
			r.Diff = strings.TrimSpace(strings.Join(diff, "\n"))
			revs = append(revs, &r)
		}
	}
	return revs, nil
}

// Stubs

func (b *bazaar) Branches() []string { return nil }

func (b *bazaar) Tags() []string {
	return []string{"tags"}
}

func (b *bazaar) ParseBranches(data []byte) ([]*Branch, error) {
	return nil, fmt.Errorf("listing branches not supported for bzr")
}

func (b *bazaar) ParseTags(data []byte) ([]*Tag, error) {
	lines, err := splitTwoValuesLines(data)
	if err != nil {
		return nil, err
	}
	tags := make([]*Tag, len(lines))
	for ii, v := range lines {
		tags[ii] = &Tag{
			Name:     v.first,
			Revision: v.second,
		}
	}
	return tags, nil
}

func parseBzrTimestamp(s string) (time.Time, error) {
	t, err := time.Parse(bzrTimestampLayout, s)
	if err == nil {
		t = t.UTC()
	}
	return t, err
}

func init() {
	Register(&bazaar{})
}
