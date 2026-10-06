package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeGitHub serves the pull request endpoints promotion uses, backed by the test's bare repository.
type fakeGitHub struct {
	mu          sync.Mutex
	bare        string
	prs         []map[string]any
	reviews     string
	autoMerge   bool
	comments    []string
	closedCount int
}

func (f *fakeGitHub) ref(name string) string {
	out, _ := exec.Command("git", "-C", f.bare, "rev-parse", "--verify", "--quiet", name).Output()
	return strings.TrimSpace(string(out))
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path := strings.TrimPrefix(r.URL.Path, "/repos/acme/app/")
	switch {
	case r.Method == http.MethodPost && path == "pulls":
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		pr := map[string]any{"number": len(f.prs) + 1, "html_url": fmt.Sprintf("https://github.com/acme/app/pull/%d", len(f.prs)+1), "state": "open", "body": body["body"], "base": body["base"], "head": map[string]any{"ref": body["head"], "sha": f.ref(body["head"])}}
		f.prs = append(f.prs, pr)
		json.NewEncoder(w).Encode(pr)
	case r.Method == http.MethodGet && path == "pulls":
		var open []map[string]any
		for _, pr := range f.prs {
			if pr["state"] == "open" {
				open = append(open, pr)
			}
		}
		json.NewEncoder(w).Encode(open)
	case strings.HasSuffix(path, "/reviews"):
		w.Write([]byte(f.reviews))
	case strings.HasPrefix(path, "pulls/"):
		var n int
		fmt.Sscanf(path, "pulls/%d", &n)
		pr := f.prs[n-1]
		if r.Method == http.MethodPatch {
			pr["state"] = "closed"
			f.closedCount++
		}
		head := pr["head"].(map[string]any)["sha"].(string)
		// Like GitHub: a pull request whose head is now on the base branch counts as merged.
		if f.autoMerge && pr["state"] == "open" && f.ref(pr["base"].(string)) == head {
			pr["state"], pr["merged"] = "closed", true
		}
		json.NewEncoder(w).Encode(pr)
	case r.Method == http.MethodPost && strings.HasPrefix(path, "issues/"):
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		f.comments = append(f.comments, body["body"])
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func TestProductionReleasesThroughAPullRequestCompletedByFastForward(t *testing.T) {
	mine, _ := sharedProject(t)
	bare := run(t, mine, "config", "--get", "remote.origin.url")
	gh := &fakeGitHub{bare: bare, reviews: "[]", autoMerge: true}
	srv := httptest.NewServer(gh)
	defer srv.Close()
	t.Setenv("MEM_GITHUB_API", srv.URL)
	t.Setenv("GITHUB_TOKEN", "t0ken")
	mergeWait = 0
	run(t, mine, "config", "remote.origin.url", "https://github.com/acme/app.git")
	run(t, mine, "config", "url."+bare+".insteadOf", "https://github.com/acme/app.git")
	releaseConfig(t, mine, "production_pr = true\nnotes = true")
	mem(t, mine, "promote", "staging")

	if out := mem(t, mine, "promote", "production"); !strings.Contains(out, "opens one with these notes for review") {
		t.Fatalf("draft instruction does not mention the pull request:\n%s", out)
	}
	mainBefore := gh.ref("main")
	out := mem(t, mine, "promote", "production", "--confirm")
	if !strings.Contains(out, "https://github.com/acme/app/pull/1") || gh.ref("main") != mainBefore || len(gh.prs) != 1 {
		t.Fatalf("opening the pull request:\n%s", out)
	}
	head := gh.prs[0]["head"].(map[string]any)
	if head["sha"] != gh.ref("test") || !strings.HasPrefix(head["ref"].(string), "promotion/main/") || !strings.Contains(gh.prs[0]["body"].(string), "# Release v") {
		t.Fatalf("pull request = %v", gh.prs[0])
	}
	if out := mem(t, mine, "promote", "production"); !strings.Contains(out, "A release pull request is open") {
		t.Fatalf("status with an open pull request:\n%s", out)
	}

	gh.reviews = `[{"user":{"login":"ann"},"state":"CHANGES_REQUESTED"}]`
	if err := memErr(t, mine, "promote", "production", "--confirm"); err == nil || !strings.Contains(err.Error(), "changes were requested") {
		t.Fatalf("completing with changes requested: %v", err)
	}
	gh.reviews = `[{"user":{"login":"ann"},"state":"APPROVED"}]`
	mem(t, mine, "promote", "production", "--confirm")
	if gh.ref("main") != head["sha"] || gh.prs[0]["merged"] != true || gh.closedCount != 0 || gh.ref(head["ref"].(string)) != "" {
		t.Fatalf("completion: main %s, head %s, pr %v, closed %d, snapshot %q", gh.ref("main"), head["sha"], gh.prs[0], gh.closedCount, gh.ref(head["ref"].(string)))
	}
	tag := run(t, mine, "describe", "--tags", "--exact-match", "origin/main")
	if notes := run(t, mine, "for-each-ref", "--format=%(contents)", "refs/tags/"+tag); !strings.Contains(notes, "# Release "+tag) {
		t.Fatalf("tag notes = %q", notes)
	}

	// When GitHub does not mark the pull request merged, mem closes it with a comment naming the release.
	gh.autoMerge, gh.reviews = false, "[]"
	commit(t, mine, "feature.txt")
	run(t, mine, "push", "--quiet")
	mem(t, mine, "promote", "staging")
	mem(t, mine, "promote", "production")
	time.Sleep(1100 * time.Millisecond) // snapshot branch names have second precision
	mem(t, mine, "promote", "production", "--confirm")
	mem(t, mine, "promote", "production", "--confirm")
	if gh.closedCount != 1 || len(gh.comments) != 1 || !strings.Contains(gh.comments[0], "Released as v") {
		t.Fatalf("fallback: closed %d, comments %v", gh.closedCount, gh.comments)
	}
}
