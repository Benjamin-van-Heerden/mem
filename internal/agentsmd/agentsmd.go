package agentsmd

import (
	_ "embed"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

//go:embed instructions.md
var instructions string

const (
	blockOpen     = "<mem>"
	blockClose    = "</mem>"
	memoriesOpen  = "<memories>"
	memoriesClose = "</memories>"
	memoriesTitle = "# Project Memories"
)

var (
	memoryName   = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
	stampVersion = regexp.MustCompile(`<!-- Managed by mem (\S+)\.`)
)

type Memory struct {
	Name string
	Body string
}

// Install adds the managed block and an empty memories block to existing AGENTS.md content.
func Install(text, version string) (string, error) {
	if strings.Contains(text, blockOpen) {
		return "", errors.New("AGENTS.md already contains a <mem> block")
	}
	parts := []string{strings.TrimSpace(text), block(version)}
	if !strings.Contains(text, memoriesOpen) {
		parts = append(parts, renderMemories(nil))
	}
	return strings.TrimLeft(strings.Join(parts, "\n\n"), "\n") + "\n", nil
}

var legacyBlock = regexp.MustCompile(`(?s)<(core_instructions|AGENT_CORE|agent_core)>.*?</(core_instructions|AGENT_CORE|agent_core)>`)

// ReplaceLegacy swaps the Python coding harness's managed block for the mem
// block in the same position and adds an empty memories block.
func ReplaceLegacy(text, version string) (string, error) {
	loc := legacyBlock.FindStringIndex(text)
	if loc == nil {
		return "", errors.New("AGENTS.md has no <core_instructions> block from the Python coding harness")
	}
	if strings.Contains(text, blockOpen) {
		return "", errors.New("AGENTS.md already contains a <mem> block")
	}
	updated := text[:loc[0]] + block(version) + text[loc[1]:]
	if !strings.Contains(updated, memoriesOpen) {
		updated = strings.TrimRight(updated, "\n") + "\n\n" + renderMemories(nil) + "\n"
	}
	return updated, nil
}

// Refresh replaces the managed block with the instructions shipped in this
// executable. A block written by a newer mem is left alone and its version returned.
func Refresh(text, version string) (updated string, newer string, err error) {
	start, end, err := span(text, blockOpen, blockClose)
	if err != nil {
		return text, "", err
	}
	if m := stampVersion.FindStringSubmatch(text[start:end]); m != nil && semverLess(version, m[1]) {
		return text, m[1], nil
	}
	return text[:start] + block(version) + text[end:], "", nil
}

func block(version string) string {
	body := strings.TrimPrefix(strings.TrimSpace(instructions), blockOpen)
	return blockOpen + "\n<!-- Managed by mem " + version + ". Edits inside this block are replaced on onboard. -->" + body
}

// semverLess reports whether a < b; versions that are not x.y.z never compare as less.
func semverLess(a, b string) bool {
	pa, oka := semver(a)
	pb, okb := semver(b)
	if !oka || !okb {
		return false
	}
	for i := range pa {
		if pa[i] != pb[i] {
			return pa[i] < pb[i]
		}
	}
	return false
}

func semver(v string) ([3]int, bool) {
	var parts [3]int
	fields := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if len(fields) != 3 {
		return parts, false
	}
	for i, f := range fields {
		n, err := strconv.Atoi(f)
		if err != nil {
			return parts, false
		}
		parts[i] = n
	}
	return parts, true
}

func Memories(text string) ([]Memory, error) {
	start, end, err := span(text, memoriesOpen, memoriesClose)
	if err != nil {
		return nil, err
	}
	inner := strings.TrimSuffix(strings.TrimPrefix(text[start:end], memoriesOpen), memoriesClose)
	var memories []Memory
	for _, line := range strings.Split(inner, "\n") {
		if name, ok := strings.CutPrefix(line, "## "); ok {
			memories = append(memories, Memory{Name: strings.TrimSpace(name)})
			continue
		}
		if len(memories) > 0 {
			memories[len(memories)-1].Body += line + "\n"
		}
	}
	for i := range memories {
		memories[i].Body = strings.TrimSpace(memories[i].Body)
	}
	return memories, nil
}

func SetMemory(text, name, body string) (string, error) {
	if !memoryName.MatchString(name) {
		return text, fmt.Errorf("memory name %q must be lowercase letters, digits, hyphens or underscores", name)
	}
	body = strings.TrimSpace(body)
	if body == "" {
		return text, errors.New("memory content is empty")
	}
	for _, line := range strings.Split(body, "\n") {
		if strings.HasPrefix(line, "## ") || strings.Contains(line, memoriesOpen) || strings.Contains(line, memoriesClose) {
			return text, errors.New("memory content must not contain level-2 headings or memories tags; use ### for sub-headings")
		}
	}
	memories, err := Memories(text)
	if err != nil {
		return text, err
	}
	replaced := false
	for i := range memories {
		if memories[i].Name == name {
			memories[i].Body = body
			replaced = true
		}
	}
	if !replaced {
		memories = append(memories, Memory{Name: name, Body: body})
	}
	return replaceMemories(text, memories)
}

func RemoveMemory(text, name string) (string, error) {
	memories, err := Memories(text)
	if err != nil {
		return text, err
	}
	kept := memories[:0]
	for _, m := range memories {
		if m.Name != name {
			kept = append(kept, m)
		}
	}
	if len(kept) == len(memories) {
		return text, fmt.Errorf("no memory named %q", name)
	}
	return replaceMemories(text, kept)
}

func replaceMemories(text string, memories []Memory) (string, error) {
	start, end, err := span(text, memoriesOpen, memoriesClose)
	if err != nil {
		return text, err
	}
	return text[:start] + renderMemories(memories) + text[end:], nil
}

func renderMemories(memories []Memory) string {
	var b strings.Builder
	b.WriteString(memoriesOpen + "\n" + memoriesTitle + "\n")
	for _, m := range memories {
		b.WriteString("\n## " + m.Name + "\n" + m.Body + "\n")
	}
	b.WriteString(memoriesClose)
	return b.String()
}

func span(text, open, close string) (int, int, error) {
	start := strings.Index(text, open)
	end := strings.Index(text, close)
	if start < 0 || end < start || strings.Count(text, open) != 1 || strings.Count(text, close) != 1 {
		return 0, 0, fmt.Errorf("AGENTS.md must contain exactly one %s ... %s block", open, close)
	}
	return start, end + len(close), nil
}
