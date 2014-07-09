package vcs

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	nCommits = 3
)

func testVcs(t *testing.T, dir string, lastRev string) {
	var root string
	if filepath.IsAbs(dir) {
		root = dir
	} else {
		root = filepath.Join("testdata", dir)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(filepath.Join(root, "foo"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Dir != abs {
		t.Fatalf("expecting dir %s, got %s", abs, s.Dir)
	}
	last, err := s.Last()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("LAST %+v", last)
	if last.Identifier != lastRev {
		t.Errorf("unexpected last revision %s, want %s", last.Identifier, lastRev)
	}
	history, err := s.History("")
	if err != nil {
		t.Fatal(err)
	}
	if n := len(history); n != nCommits {
		t.Errorf("invalid history length. want %d, got %d", nCommits, n)
	}
	t.Logf("previous to last commit was %s", history[1].Identifier)
	history2, err := s.History(history[1].Identifier)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(history2); n != 1 {
		t.Errorf("invalid history since previous to last commit length. want %d, got %d", 1, n)
		for _, v := range history2 {
			t.Errorf("%+v", v)
		}
	} else {
		if history2[0].Identifier != history[0].Identifier {
			t.Errorf("bad identifier when fetching history since HEAD - 1. want %v, got %v", history[0].Identifier, history2[0].Identifier)
		}
	}
}

func TestGit(t *testing.T) {
	// Make a link for the test, then remove it. Otherwise
	// git won't let us add the testdata because it complains
	// that it's a submodule.
	newname := filepath.Join("testdata", "git", ".git")
	defer os.Remove(newname)
	if err := os.Symlink("git", newname); err != nil {
		t.Fatal(err)
	}
	testVcs(t, "git", "fe645e3acddd21db9633c6abeffe2671342d1b08")
}

func TestGitBare(t *testing.T) {
	testVcs(t, "git.bare", "fe645e3acddd21db9633c6abeffe2671342d1b08")
}

func TestMercurial(t *testing.T) {
	testVcs(t, "hg", "55a0ce7ec898b893e63823c188340d676254eb00")
}

func TestBazaar(t *testing.T) {
	testVcs(t, "bzr", "3")
}
