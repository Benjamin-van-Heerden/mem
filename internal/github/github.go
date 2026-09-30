// Package github is the small part of GitHub's REST API that promotion pull requests need. It is only used by
// projects that release production through a pull request, so nothing else depends on GitHub access.
package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

const requestTimeout = 20 * time.Second

var remotePattern = regexp.MustCompile(`^(?:https://(?:[^@/]+@)?github\.com/|git@github\.com:|ssh://git@github\.com/)([^/]+)/([^/]+?)(?:\.git)?/?$`)

type Client struct {
	api, token, owner, repo string
	http                    *http.Client
}

type PullRequest struct {
	Number  int    `json:"number"`
	URL     string `json:"html_url"`
	State   string `json:"state"`
	Merged  bool   `json:"merged"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	HeadRef string `json:"-"`
	HeadSHA string `json:"-"`
}

// Review sums up a pull request's reviews: each reviewer's latest verdict counts.
type Review struct {
	Approvals        int
	ChangesRequested bool
}

// New connects to the GitHub repository behind remoteURL, with a token from GITHUB_TOKEN, GH_TOKEN or `gh auth token`.
func New(ctx context.Context, remoteURL string) (*Client, error) {
	m := remotePattern.FindStringSubmatch(strings.TrimSpace(remoteURL))
	if m == nil {
		return nil, fmt.Errorf("%s is not a GitHub repository URL; promotion pull requests need a GitHub remote", remoteURL)
	}
	token, err := findToken(ctx)
	if err != nil {
		return nil, err
	}
	api := "https://api.github.com"
	if override := os.Getenv("MEM_GITHUB_API"); override != "" {
		api = strings.TrimSuffix(override, "/")
	}
	return &Client{api: api, token: token, owner: m[1], repo: m[2], http: &http.Client{Timeout: requestTimeout}}, nil
}

func findToken(ctx context.Context) (string, error) {
	for _, name := range []string{"GITHUB_TOKEN", "GH_TOKEN"} {
		if token := strings.TrimSpace(os.Getenv(name)); token != "" {
			return token, nil
		}
	}
	if _, err := exec.LookPath("gh"); err == nil {
		out, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()
		if token := strings.TrimSpace(string(out)); err == nil && token != "" {
			return token, nil
		}
	}
	return "", errors.New("this repository releases production through a pull request, which needs GitHub access: set GITHUB_TOKEN (or GH_TOKEN) to a token with pull request and contents access, or log in with `gh auth login`")
}

// Repo is "owner/name".
func (c *Client) Repo() string { return c.owner + "/" + c.repo }

func (c *Client) CreatePR(ctx context.Context, title, body, head, base string) (PullRequest, error) {
	var pr pullJSON
	err := c.do(ctx, http.MethodPost, c.path("pulls"), map[string]string{"title": title, "body": body, "head": head, "base": base}, &pr)
	return pr.value(), err
}

// OpenPRs lists the open pull requests into base whose head branch starts with headPrefix.
func (c *Client) OpenPRs(ctx context.Context, base, headPrefix string) ([]PullRequest, error) {
	var list []pullJSON
	if err := c.do(ctx, http.MethodGet, c.path("pulls")+"?state=open&per_page=100&base="+url.QueryEscape(base), nil, &list); err != nil {
		return nil, err
	}
	var prs []PullRequest
	for _, pr := range list {
		if strings.HasPrefix(pr.Head.Ref, headPrefix) {
			prs = append(prs, pr.value())
		}
	}
	return prs, nil
}

func (c *Client) PR(ctx context.Context, number int) (PullRequest, error) {
	var pr pullJSON
	err := c.do(ctx, http.MethodGet, c.path(fmt.Sprintf("pulls/%d", number)), nil, &pr)
	return pr.value(), err
}

func (c *Client) Reviews(ctx context.Context, number int) (Review, error) {
	var list []struct {
		User  struct{ Login string } `json:"user"`
		State string                 `json:"state"`
	}
	if err := c.do(ctx, http.MethodGet, c.path(fmt.Sprintf("pulls/%d/reviews?per_page=100", number)), nil, &list); err != nil {
		return Review{}, err
	}
	latest := map[string]string{}
	for _, r := range list {
		if r.State == "APPROVED" || r.State == "CHANGES_REQUESTED" || r.State == "DISMISSED" {
			latest[r.User.Login] = r.State
		}
	}
	var review Review
	for _, state := range latest {
		switch state {
		case "APPROVED":
			review.Approvals++
		case "CHANGES_REQUESTED":
			review.ChangesRequested = true
		}
	}
	return review, nil
}

func (c *Client) Comment(ctx context.Context, number int, body string) error {
	return c.do(ctx, http.MethodPost, c.path(fmt.Sprintf("issues/%d/comments", number)), map[string]string{"body": body}, nil)
}

func (c *Client) Close(ctx context.Context, number int) error {
	return c.do(ctx, http.MethodPatch, c.path(fmt.Sprintf("pulls/%d", number)), map[string]string{"state": "closed"}, nil)
}

func (c *Client) path(rest string) string { return fmt.Sprintf("%s/repos/%s/%s/%s", c.api, c.owner, c.repo, rest) }

func (c *Client) do(ctx context.Context, method, target string, body, result any) error {
	var payload io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub request failed: %w", err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		var apiErr struct {
			Message string `json:"message"`
			Errors  []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		json.Unmarshal(data, &apiErr)
		msg := apiErr.Message
		for _, e := range apiErr.Errors {
			if e.Message != "" {
				msg += ": " + e.Message
			}
		}
		return fmt.Errorf("GitHub %s %s: %d %s", method, strings.TrimPrefix(target, c.api), resp.StatusCode, msg)
	}
	if result == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, result)
}

type pullJSON struct {
	PullRequest
	Head struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
}

func (p pullJSON) value() PullRequest {
	pr := p.PullRequest
	pr.HeadRef, pr.HeadSHA = p.Head.Ref, p.Head.SHA
	return pr
}
