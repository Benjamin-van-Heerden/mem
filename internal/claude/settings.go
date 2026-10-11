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

// CompactCommand runs mem's catch-up after compaction; it does nothing where mem is not installed.
const CompactCommand = "command -v mem >/dev/null 2>&1 && mem hook compact || true"

const compactMarker = "mem hook compact"

type hookEntry struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookCommand `json:"hooks"`
}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

var compactEntry = hookEntry{Matcher: "compact", Hooks: []hookCommand{{Type: "command", Command: CompactCommand, Timeout: 60}}}

// SyncCompactHook adds, updates or removes mem's SessionStart "compact" hook in
// .claude/settings.json, keeping every other setting and its order. It writes
// only when mem's entry changes and reports whether it did.
func SyncCompactHook(root string, enabled bool) (bool, error) {
	path := filepath.Join(root, SettingsPath)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if !enabled {
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
	var entries []json.RawMessage
	if raw, ok := hooks.get("SessionStart"); ok {
		if err := json.Unmarshal(raw, &entries); err != nil {
			return false, fmt.Errorf("%s: hooks.SessionStart is not a list: %w", SettingsPath, err)
		}
	}

	want, err := marshal(compactEntry)
	if err != nil {
		return false, err
	}
	var kept []json.RawMessage
	placed := false
	for _, raw := range entries {
		if !isMemEntry(raw) {
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

	if len(kept) > 0 {
		list, err := marshal(kept)
		if err != nil {
			return false, err
		}
		hooks.set("SessionStart", list)
	} else {
		hooks.remove("SessionStart")
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

func isMemEntry(raw json.RawMessage) bool {
	var e hookEntry
	if json.Unmarshal(raw, &e) != nil {
		return false
	}
	for _, h := range e.Hooks {
		if strings.Contains(h.Command, compactMarker) {
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
