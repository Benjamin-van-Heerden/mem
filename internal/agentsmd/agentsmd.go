package agentsmd

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// Kind selects the instructions a project gets.
type Kind int

const (
	Staged Kind = iota
	SingleBranch
	Folder
)

//go:embed instructions.md
var instructions string

// folderInstructions are a block of their own: a folder project does document work, so it shares little with the coding
// instructions.
//
//go:embed instructions_folder.md
var folderInstructions string

// instructionsTemplate numbers the Ground Rules with {{rule}} and leaves out with .Single what applies only to
// projects with staging and production branches.
var instructionsTemplate = template.Must(template.New("instructions.md").Funcs(template.FuncMap{"rule": func() int { return 0 }}).Parse(instructions))

//go:embed install.md
var installGuide string

//go:embed readme.md
var readmeGuide string

// Guide is a file mem ships in this executable and rewrites in every project at onboard. Guides live in the project,
// not AGENTS.md, so a teammate without mem has them as soon as they pull, and they cost no context until an agent
// needs them.
type Guide struct {
	Path string
	// About says what the guide is, for mem's output.
	About   string
	content string
}

var (
	// InstallGuide is where the managed block points agents when mem is not installed.
	InstallGuide = Guide{".mem/install.md", "the install guide AGENTS.md points to when mem is missing", installGuide}
	// Readme explains how mem works in a project: its files, configuration, runnables, records and hooks.
	Readme = Guide{".mem/README.md", "the guide to how mem works in this project", readmeGuide}
	Guides = []Guide{InstallGuide, Readme}
)

// Write writes the guide into the project at root, reporting whether the file changed.
func (g Guide) Write(root string) (bool, error) {
	path := filepath.Join(root, filepath.FromSlash(g.Path))
	if current, err := os.ReadFile(path); err == nil && string(current) == g.content {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, []byte(g.content), 0o644)
}

// WriteGuides writes every guide into the project at root.
func WriteGuides(root string) error {
	for _, g := range Guides {
		if _, err := g.Write(root); err != nil {
			return err
		}
	}
	return nil
}

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

// Install puts the managed block for kind and an empty memories block ahead of existing AGENTS.md content.
func Install(text, version string, kind Kind) (string, error) {
	if hasTag(text, blockOpen) {
		return "", errors.New("AGENTS.md already contains a <mem> block")
	}
	parts := []string{block(version, kind)}
	if !hasTag(text, memoriesOpen) {
		parts = append(parts, renderMemories(nil))
	}
	parts = append(parts, strings.TrimSpace(text))
	return strings.TrimRight(strings.Join(parts, "\n\n"), "\n") + "\n", nil
}

var (
	legacyBlock = regexp.MustCompile(`(?s)<(core_instructions|AGENT_CORE|agent_core)>.*?</(core_instructions|AGENT_CORE|agent_core)>`)
	// Later harness versions wrote a new block without removing the old one, so a file can hold several.
	extraLegacyBlock = regexp.MustCompile(legacyBlock.String() + `\n*`)
)

// HasLegacy reports whether text holds a managed block from the Python coding harness.
func HasLegacy(text string) bool { return legacyBlock.MatchString(text) }

// ReplaceLegacy swaps the Python coding harness's first managed block for the mem
// block in the same position, drops any further harness blocks and adds an empty memories block.
func ReplaceLegacy(text, version string, kind Kind) (string, error) {
	loc := legacyBlock.FindStringIndex(text)
	if loc == nil {
		return "", errors.New("AGENTS.md has no <core_instructions> block from the Python coding harness")
	}
	if hasTag(text, blockOpen) {
		return "", errors.New("AGENTS.md already contains a <mem> block")
	}
	updated := text[:loc[0]] + block(version, kind) + extraLegacyBlock.ReplaceAllString(text[loc[1]:], "")
	if !hasTag(updated, memoriesOpen) {
		updated = strings.TrimRight(updated, "\n") + "\n\n" + renderMemories(nil) + "\n"
	}
	return updated, nil
}

// Refresh replaces the managed block with the instructions for kind shipped in this executable. A block written by a
// newer mem is left alone and its version returned.
func Refresh(text, version string, kind Kind) (updated string, newer string, err error) {
	start, end, err := span(text, blockOpen, blockClose)
	if err != nil {
		return text, "", err
	}
	if m := stampVersion.FindStringSubmatch(text[start:end]); m != nil && semverLess(version, m[1]) {
		return text, m[1], nil
	}
	return text[:start] + block(version, kind) + text[end:], "", nil
}

func block(version string, kind Kind) string {
	body := strings.TrimPrefix(strings.TrimSpace(render(kind)), blockOpen)
	return blockOpen + "\n<!-- Managed by mem " + version + ". Edits inside this block are replaced on onboard. -->" + body
}

// render gives the instructions for kind. Executing the template can only fail on a mistake in instructions.md, which
// tests catch.
func render(kind Kind) string {
	if kind == Folder {
		return folderInstructions
	}
	rule := 0
	t := template.Must(instructionsTemplate.Clone()).Funcs(template.FuncMap{"rule": func() int { rule++; return rule }})
	var b strings.Builder
	if err := t.Execute(&b, struct{ Single bool }{kind == SingleBranch}); err != nil {
		panic(err)
	}
	return b.String()
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
	opens, closes := tagLines(text, open), tagLines(text, close)
	if len(opens) != 1 || len(closes) != 1 || closes[0] < opens[0] {
		return 0, 0, fmt.Errorf("AGENTS.md must contain exactly one %s ... %s block, with each tag on a line of its own", open, close)
	}
	return opens[0], closes[0] + len(close), nil
}

// tagLines returns the offsets of the lines that consist of tag alone, so that
// mentions of a tag inside prose or code spans are not mistaken for the block.
func tagLines(text, tag string) []int {
	var offsets []int
	offset := 0
	for _, line := range strings.SplitAfter(text, "\n") {
		if strings.TrimRight(line, "\r\n") == tag {
			offsets = append(offsets, offset)
		}
		offset += len(line)
	}
	return offsets
}

func hasTag(text, tag string) bool { return len(tagLines(text, tag)) > 0 }
