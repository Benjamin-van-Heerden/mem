package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRemoteURLsAndTokens(t *testing.T) {
	t.Setenv("MEM_GITHUB_API", "http://unused")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_TOKEN", "t0ken")
	for _, remote := range []string{"https://github.com/acme/app.git", "https://x-access-token:abc@github.com/acme/app", "git@github.com:acme/app.git", "ssh://git@github.com/acme/app.git"} {
		c, err := New(context.Background(), remote)
		if err != nil || c.Repo() != "acme/app" {
			t.Fatalf("New(%q) = %v, %v", remote, c, err)
		}
	}
	if _, err := New(context.Background(), "https://gitlab.com/acme/app.git"); err == nil {
		t.Fatal("a non-GitHub remote was accepted")
	}
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("PATH", t.TempDir())
	if _, err := New(context.Background(), "git@github.com:acme/app.git"); err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") {
		t.Fatalf("missing token error = %v", err)
	}
}

func TestPullRequestsAndReviews(t *testing.T) {
	var created map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer t0ken" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/repos/acme/app/pulls":
			json.NewDecoder(r.Body).Decode(&created)
			if created["head"] == "taken" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				w.Write([]byte(`{"message":"Validation Failed","errors":[{"message":"A pull request already exists"}]}`))
				return
			}
			w.Write([]byte(`{"number":7,"html_url":"https://github.com/acme/app/pull/7","state":"open","head":{"ref":"promotion/main/1","sha":"abc"}}`))
		case r.URL.Path == "/repos/acme/app/pulls" && r.URL.Query().Get("base") == "main":
			w.Write([]byte(`[{"number":7,"state":"open","head":{"ref":"promotion/main/1","sha":"abc"}},{"number":8,"state":"open","head":{"ref":"feature","sha":"def"}}]`))
		case r.URL.Path == "/repos/acme/app/pulls/7/reviews":
			w.Write([]byte(`[{"user":{"login":"ann"},"state":"CHANGES_REQUESTED"},{"user":{"login":"ann"},"state":"APPROVED"},{"user":{"login":"bob"},"state":"COMMENTED"},{"user":{"login":"cy"},"state":"APPROVED"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	t.Setenv("MEM_GITHUB_API", srv.URL)
	t.Setenv("GITHUB_TOKEN", "t0ken")
	ctx := context.Background()
	c, err := New(ctx, "git@github.com:acme/app.git")
	if err != nil {
		t.Fatal(err)
	}

	pr, err := c.CreatePR(ctx, "Release v1", "Notes", "promotion/main/1", "main")
	if err != nil || pr.Number != 7 || pr.HeadSHA != "abc" || created["body"] != "Notes" || created["base"] != "main" {
		t.Fatalf("CreatePR = %+v, %v (sent %v)", pr, err, created)
	}
	if _, err := c.CreatePR(ctx, "x", "y", "taken", "main"); err == nil || !strings.Contains(err.Error(), "A pull request already exists") {
		t.Fatalf("API error = %v", err)
	}
	open, err := c.OpenPRs(ctx, "main", "promotion/main/")
	if err != nil || len(open) != 1 || open[0].Number != 7 || open[0].HeadRef != "promotion/main/1" {
		t.Fatalf("OpenPRs = %+v, %v", open, err)
	}
	review, err := c.Reviews(ctx, 7)
	if err != nil || review.Approvals != 2 || review.ChangesRequested {
		t.Fatalf("Reviews = %+v, %v (ann's latest review approves)", review, err)
	}
}
