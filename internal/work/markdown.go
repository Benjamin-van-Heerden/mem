package work

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.yaml.in/yaml/v3"
)

func ReadMarkdown(path string, meta any) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	rest, ok := strings.CutPrefix(text, "---\n")
	if !ok {
		return "", fmt.Errorf("%s has no YAML frontmatter", path)
	}
	front, body, ok := strings.Cut(rest, "\n---\n")
	if !ok {
		return "", fmt.Errorf("%s has unterminated YAML frontmatter", path)
	}
	if err := yaml.Unmarshal([]byte(front), meta); err != nil {
		return "", fmt.Errorf("%s has invalid frontmatter: %w", path, err)
	}
	return strings.TrimLeft(body, "\n"), nil
}

func WriteMarkdown(path string, meta any, body string) error {
	var front bytes.Buffer
	enc := yaml.NewEncoder(&front)
	enc.SetIndent(2)
	if err := enc.Encode(meta); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := "---\n" + front.String() + "---\n\n" + strings.TrimSpace(body) + "\n"
	return os.WriteFile(path, []byte(content), 0o644)
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

type named interface {
	slug() string
	title() string
}

// resolve finds an item by exact slug, then case-insensitive title, then unique slug prefix.
func resolve[T named](kind, ref string, items []T) (T, error) {
	var zero T
	ref = strings.TrimSpace(ref)
	for _, item := range items {
		if item.slug() == ref {
			return item, nil
		}
	}
	var matches []T
	for _, item := range items {
		if strings.EqualFold(item.title(), ref) {
			matches = append(matches, item)
		}
	}
	if len(matches) == 0 {
		for _, item := range items {
			if strings.HasPrefix(item.slug(), ref) {
				matches = append(matches, item)
			}
		}
	}
	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return zero, fmt.Errorf("no %s matches %q; run `memr %s list` to see what exists", kind, ref, kind)
	}
	slugs := make([]string, len(matches))
	for i, m := range matches {
		slugs[i] = m.slug()
	}
	return zero, fmt.Errorf("%q matches several %ss: %s; use the full slug", ref, kind, strings.Join(slugs, ", "))
}

func uniqueSlug(base string, exists func(string) bool) (string, error) {
	if base == "" {
		return "", errors.New("title must contain letters or digits")
	}
	slug := base
	for i := 2; exists(slug); i++ {
		slug = fmt.Sprintf("%s_%d", base, i)
	}
	return slug, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
