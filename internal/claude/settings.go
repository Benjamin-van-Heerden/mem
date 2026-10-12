// Package claude maintains mem's entries in a project's Claude Code settings.
package claude

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SettingsPath is the project's shared Claude Code settings file, relative to the repository root.
const SettingsPath = ".claude/settings.json"

// CompactCommand runs mem's catch-up after compaction; PromptCommand reports the work since the last work log
// when the user sends a message. Both do nothing where mem is not installed.
const (
	CompactCommand = "command -v mem >/dev/null 2>&1 && mem hook compact || true"
	PromptCommand  = "command -v mem >/dev/null 2>&1 && mem hook prompt || true"
)

type hookEntry struct {
	Matcher string        `json:"matcher,omitempty"`
	Hooks   []hookCommand `json:"hooks"`
}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// memHook is one of mem's entries: the event it belongs to and the command that identifies it.
type memHook struct {
	event, marker string
	entry         hookEntry
}

var (
	compactHook = memHook{"SessionStart", "mem hook compact", hookEntry{Matcher: "compact", Hooks: []hookCommand{{Type: "command", Command: CompactCommand, Timeout: 60}}}}
	promptHook  = memHook{"UserPromptSubmit", "mem hook prompt", hookEntry{Hooks: []hookCommand{{Type: "command", Command: PromptCommand, Timeout: 10}}}}
)

// Hooks says which of mem's Claude Code hooks a project wants.
type Hooks struct {
	Compact, Prompt bool
}

// SyncHooks adds, updates or removes mem's hooks in .claude/settings.json: the SessionStart "compact" catch-up and
// the UserPromptSubmit work-log nudge. It keeps every other setting and its order, writes only when mem's entries
// change and reports whether it did.
func SyncHooks(root string, want Hooks) (bool, error) {
	path := filepath.Join(root, SettingsPath)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if !want.Compact && !want.Prompt {
			return false, nil
		}
		data = []byte("{}")
	} else if err != nil {
		return false, err
	}
	var settings object
	if err := json.Unmarshal(data, &settings); err != nil {
		return false, fmt.Errorf("%s is not a JSON object: %w", SettingsPath, err)
	}
	var hooks object
	if raw, ok := settings.get("hooks"); ok {
		if err := json.Unmarshal(raw, &hooks); err != nil {
			return false, fmt.Errorf("%s: hooks is not an object: %w", SettingsPath, err)
		}
	}
	for _, h := range []struct {
		hook    memHook
		enabled bool
	}{{compactHook, want.Compact}, {promptHook, want.Prompt}} {
		if err := syncEntry(&hooks, h.hook, h.enabled); err != nil {
			return false, err
		}
	}
	if len(hooks) > 0 {
		encoded, err := marshal(hooks)
		if err != nil {
			return false, err
		}
		settings.set("hooks", encoded)
	} else {
		settings.remove("hooks")
	}

	updated, err := indent(settings)
	if err != nil {
		return false, err
	}
	if current, err := indent(json.RawMessage(data)); err == nil && bytes.Equal(current, updated) {
		return false, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, err
	}
	return true, os.WriteFile(path, updated, 0o644)
}

// syncEntry puts h's entry in place of mem's existing one in its event's list, or appends it, or removes it.
func syncEntry(hooks *object, h memHook, enabled bool) error {
	var entries []json.RawMessage
	if raw, ok := hooks.get(h.event); ok {
		if err := json.Unmarshal(raw, &entries); err != nil {
			return fmt.Errorf("%s: hooks.%s is not a list: %w", SettingsPath, h.event, err)
		}
	}
	want, err := marshal(h.entry)
	if err != nil {
		return err
	}
	var kept []json.RawMessage
	placed := false
	for _, raw := range entries {
		if !isEntry(raw, h.marker) {
			kept = append(kept, raw)
			continue
		}
		if enabled && !placed {
			kept = append(kept, want)
			placed = true
		}
	}
	if enabled && !placed {
		kept = append(kept, want)
	}
	if len(kept) == 0 {
		hooks.remove(h.event)
		return nil
	}
	list, err := marshal(kept)
	if err != nil {
		return err
	}
	hooks.set(h.event, list)
	return nil
}

func isEntry(raw json.RawMessage, marker string) bool {
	var e hookEntry
	if json.Unmarshal(raw, &e) != nil {
		return false
	}
	for _, h := range e.Hooks {
		if strings.Contains(h.Command, marker) {
			return true
		}
	}
	return false
}

// marshal encodes without HTML escaping, so shell operators such as && stay readable.
func marshal(v any) (json.RawMessage, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSpace(buf.Bytes()), nil
}

func indent(v any) ([]byte, error) {
	raw, err := marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// object is a JSON object that keeps its keys in their original order.
type object []field

type field struct {
	key   string
	value json.RawMessage
}

func (o object) get(key string) (json.RawMessage, bool) {
	for _, f := range o {
		if f.key == key {
			return f.value, true
		}
	}
	return nil, false
}

func (o *object) set(key string, value json.RawMessage) {
	for i := range *o {
		if (*o)[i].key == key {
			(*o)[i].value = value
			return
		}
	}
	*o = append(*o, field{key, value})
}

func (o *object) remove(key string) {
	for i := range *o {
		if (*o)[i].key == key {
			*o = append((*o)[:i], (*o)[i+1:]...)
			return
		}
	}
}

func (o *object) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return errors.New("expected an object")
	}
	*o = nil
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return err
		}
		*o = append(*o, field{tok.(string), value})
	}
	_, err := dec.Token()
	return err
}

func (o object) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, f := range o {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, err := marshal(f.key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(f.value)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
