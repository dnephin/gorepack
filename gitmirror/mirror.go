package gitmirror

import (
	"github.com/pkg/errors"
	gitobj "gopkg.in/src-d/go-git.v4/plumbing/object"
)

// MirrorOpts used by MirrorCommit
type MirrorOpts struct {
	Modifiers []ModifyFunc
}

// ModifyFunc takes a list of file paths of the files that changed in the commit
type ModifyFunc func(paths []string) error

// MirrorCommits from one git repo to another. For each commit in source copy all
// the files modified by the commit to the target repo, run modifiers, then
// commit the changes in target using the commit message from source. The commit
// message is amended to include a reference to the original commit. The commit
// is then tagged with with a git tag which contains the original commit sha.
func MirrorCommits(source string, target string, opts MirrorOpts) error {
	newCommits, err := getNewCommitList(source, target)
	if err != nil {
		return errors.Wrap(err, "failed to get git commit list")
	}

	for _, commit := range newCommits {
		checkoutParent(target, commit)
		copyFiles(source, target, commit)
		applyModifiers(target, commit)
		targetSha := commitTarget(target, commit)
		tagTarget(target, commit, targetSha)
	}
	return nil
}

func getNewCommitList(source, target string) ([]gitobj.Commit, error) {
	return nil, nil
}
