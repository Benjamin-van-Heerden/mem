package hooks

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Benjamin-van-Heerden/memr/internal/git"
	"github.com/Benjamin-van-Heerden/memr/internal/project"
	"github.com/Benjamin-van-Heerden/memr/internal/release"
)

func testProject(t *testing.T, protect bool) project.Project {
	root := t.TempDir()
	if _, err := git.Run(context.Background(), root, "init", "--quiet", "-b", "dev"); err != nil {
		t.Fatal(err)
	}
	return project.Project{Root: root, Config: project.Config{Git: project.GitConfig{Remote: "origin", Development: "dev", Staging: "test", Production: "main", Protect: protect}}}
}

func TestPrePushBlocksStageBranchesUnlessPromoting(t *testing.T) {
	p := testProject(t, true)
	refs := func(ref string) *strings.Reader {
		return strings.NewReader("refs/heads/dev abc " + ref + " def\n")
	}
	if err := PrePush(p, "origin", refs("refs/heads/dev")); err != nil {
		t.Fatalf("push to dev blocked: %v", err)
	}
	if err := PrePush(p, "origin", refs("refs/heads/main")); err == nil {
		t.Fatal("push to main allowed")
	}
	t.Setenv(release.PromoteEnv, "1")
	if err := PrePush(p, "origin", refs("refs/heads/main")); err != nil {
		t.Fatalf("promotion push blocked: %v", err)
	}
}

func TestSyncKeepsForeignHooksAndRemovesOwnWhenUnprotected(t *testing.T) {
	ctx := context.Background()
	p := testProject(t, true)
	foreign := filepath.Join(p.Root, ".git", "hooks", "pre-commit")
	if err := os.WriteFile(foreign, []byte("#!/bin/sh\nnpx lint-staged\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	lines, err := Sync(ctx, p)
	if err != nil || len(lines) != 2 {
		t.Fatalf("lines=%v err=%v", lines, err)
	}
	if data, _ := os.ReadFile(foreign); !strings.Contains(string(data), "lint-staged") {
		t.Fatal("foreign hook was overwritten")
	}
	p.Config.Git.Protect = false
	if _, err := Sync(ctx, p); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(p.Root, ".git", "hooks", "pre-push")); !os.IsNotExist(err) {
		t.Fatal("memr pre-push hook not removed")
	}
	if _, err := os.Stat(foreign); err != nil {
		t.Fatal("foreign hook removed")
	}
}

func TestInstalledHookIsInertWithoutAMemrThatSupportsHooks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Git runs hooks through its own sh on Windows")
	}
	p := testProject(t, true)
	if _, err := Sync(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	bin := t.TempDir()
	old := "#!/bin/sh\necho \"unknown command\" >&2\nexit 1\n"
	if err := os.WriteFile(filepath.Join(bin, "memr"), []byte(old), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := exec.Command(filepath.Join(p.Root, ".git", "hooks", "pre-commit")).Run(); err != nil {
		t.Fatalf("hook blocked a commit when memr has no hook command: %v", err)
	}
}
