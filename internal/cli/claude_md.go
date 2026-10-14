package cli

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Benjamin-van-Heerden/mem/internal/agentsmd"
	"github.com/Benjamin-van-Heerden/mem/internal/git"
)

// removeClaudeLink deletes a CLAUDE.md that is only a link to AGENTS.md. Claude Code reads AGENTS.md itself, so the
// link is one more file to keep. A CLAUDE.md with its own content is left alone.
func removeClaudeLink(root string) (bool, error) {
	path := filepath.Join(root, "CLAUDE.md")
	target, err := os.Readlink(path)
	if err != nil {
		return false, nil
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, target)
	}
	if filepath.Clean(target) != filepath.Join(root, "AGENTS.md") {
		return false, nil
	}
	return true, os.Remove(path)
}

const claudeLinkRemoved = "Removed CLAUDE.md, a link to AGENTS.md: Claude Code reads AGENTS.md itself."

// settleClaudeFile removes a CLAUDE.md that only stands in for AGENTS.md: a link to it, or the old harness's copy
// of its instructions (made on Windows, often untracked), which contradicts mem because Claude Code loads both
// files. A CLAUDE.md with its own content is flagged instead. tracked reports a removal Git still has to record.
func settleClaudeFile(ctx context.Context, root string) (line string, tracked bool, err error) {
	path := filepath.Join(root, "CLAUDE.md")
	removed, err := removeClaudeLink(root)
	switch {
	case err != nil:
		return "", false, err
	case removed:
		line = claudeLinkRemoved
	default:
		data, err := os.ReadFile(path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return "", false, nil
		case err != nil:
			return "", false, err
		case !agentsmd.HasLegacy(string(data)):
			return "⚠️ CLAUDE.md has its own instructions, and Claude Code loads them alongside AGENTS.md. Tell the user; move what is still needed into AGENTS.md and delete CLAUDE.md.", false, nil
		}
		if err := os.Remove(path); err != nil {
			return "", false, err
		}
		line = "Removed CLAUDE.md, a copy of the old harness instructions: Claude Code reads AGENTS.md itself. If this session loaded CLAUDE.md at its start, disregard those instructions and follow AGENTS.md."
	}
	inIndex, _ := git.Run(ctx, root, "ls-files", "--", "CLAUDE.md")
	return line, inIndex != "", nil
}
