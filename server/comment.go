package main

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	xssh "golang.org/x/crypto/ssh"
	"golang.org/x/xerrors"
)

type CreateCommentRequest struct {
	PostID      string `json:"postId"`
	AuthorName  string `json:"-"`
	AuthorEmail string `json:"-"`
	Content     string `json:"content"`
}

type Comment struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Content     string    `json:"content"`
}

type Env struct {
	PrivateKeyPem string
}

func postCommentInner(ctx context.Context, r CreateCommentRequest, env Env) (Comment, error) {
	// create comment
	now := time.Now()
	comment := Comment{
		ID:          strconv.FormatInt(now.Unix(), 10) + "-" + RandStringBytes(8),
		Date:        now,
		AuthorName:  r.AuthorName,
		AuthorEmail: r.AuthorEmail,
		Content:     createCommentFileContent(r, now),
	}

	// create temporary directory
	dir, err := ioutil.TempDir("", "wisdom-content-repo-*")
	if err != nil {
		return Comment{}, xerrors.Errorf("create tempory directory: %v", err)
	}
	defer os.RemoveAll(dir) // clean up

	// create auth method for git
	auth, err := ssh.NewPublicKeys("git", []byte(env.PrivateKeyPem), "")
	auth.HostKeyCallback = xssh.InsecureIgnoreHostKey()
	if err != nil {
		return Comment{}, xerrors.Errorf("read auth to push: %v", err)
	}

	// git clone
	repo, err := git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
		URL:           os.Getenv("CONTENT_GIT_REPOSITORY_URL"),
		Auth:          auth,
		ReferenceName: "refs/heads/master",
		SingleBranch:  true,
		Depth:         1,
	})
	if err != nil {
		return Comment{}, xerrors.Errorf("git clone: %v", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return Comment{}, xerrors.Errorf("get git worktree from repo: %v", err)
	}

	// create comment file
	file, err := createCommentFile(ctx, comment, filepath.Join(dir, "comments"))
	if err != nil {
		return Comment{}, xerrors.Errorf("create comment file: %v", err)
	}

	// stage comment file
	_, err = worktree.Add(file[(len(dir) + 1):])
	if err != nil {
		return Comment{}, xerrors.Errorf("add comment file into git worktree: %v", err)
	}

	// commit
	message := "[api] " + r.AuthorName + " comment on " + r.PostID
	_, err = worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  r.AuthorName,
			Email: r.AuthorEmail,
			When:  now,
		},
		Committer: &object.Signature{
			Name:  "Wisdom HTTP API",
			Email: "wisdom-http-api@koharakazuya.net",
			When:  now,
		},
	})
	if err != nil {
		return Comment{}, xerrors.Errorf("git commit: %v", err)
	}

	// push
	err = repo.PushContext(ctx, &git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/heads/master:refs/heads/master"},
		Auth:       auth,
	})
	if err != nil {
		return Comment{}, xerrors.Errorf("git push: %v", err)
	}

	return comment, nil
}

func createCommentFileContent(r CreateCommentRequest, now time.Time) string {
	var c string

	// front matter
	c += "---\n"
	c += "date: " + now.Format(time.RFC3339) + "\n"
	c += "author: " + r.AuthorName + " <" + r.AuthorEmail + ">" + "\n"
	c += "---\n"

	// body
	c += "\n" + r.Content

	// eof
	if !strings.HasSuffix(c, "\n") {
		c += "\n"
	}

	return c
}

func createCommentFile(ctx context.Context, comment Comment, dir string) (string, error) {
	path := filepath.Join(dir, comment.ID)

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", xerrors.Errorf("create comment file: %v", err)
	}

	_, err = file.WriteString(comment.Content)
	if err != nil {
		return "", xerrors.Errorf("write content into comment file: %v", err)
	}

	err = file.Close()
	if err != nil {
		return "", xerrors.Errorf("close comment file: %v", err)
	}

	return path, nil
}
