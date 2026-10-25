package agenthooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// memEntry is one of mem's entries in an agent's hook file: the event list it belongs to, the entry, how to recognise
// it among other entries and whether the project wants it.
type memEntry struct {
	event   string
	entry   any
	mine    func(json.RawMessage) bool
	enabled bool
}

// syncFile puts mem's entries in place in the lists under the "hooks" object of the JSON file at rel, removing the
// disabled ones. It keeps every other key, event and entry in order and writes only on change. A missing file starts
// from fresh, and is not created when no entry is enabled.
func syncFile(root, rel, fresh string, entries []memEntry) (bool, error) {
	path := filepath.Join(root, rel)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		wanted := false
		for _, e := range entries {
			wanted = wanted || e.enabled
		}
		if !wanted {
			return false, nil
		}
		data = []byte(fresh)
	} else if err != nil {
		return false, err
	}
	var file object
	if err := json.Unmarshal(data, &file); err != nil {
		return false, fmt.Errorf("%s is not a JSON object: %w", rel, err)
	}
	var hooks object
	if raw, ok := file.get("hooks"); ok {
		if err := json.Unmarshal(raw, &hooks); err != nil {
			return false, fmt.Errorf("%s: hooks is not an object: %w", rel, err)
		}
	}
	for _, e := range entries {
		if err := syncEntry(&hooks, rel, e); err != nil {
			return false, err
		}
	}
	if len(hooks) > 0 {
		encoded, err := marshal(hooks)
		if err != nil {
			return false, err
		}
		file.set("hooks", encoded)
	} else {
		file.remove("hooks")
	}

	updated, err := indent(file)
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

// syncEntry puts e's entry in place of mem's existing one in its event's list, or appends it, or removes it.
func syncEntry(hooks *object, rel string, e memEntry) error {
	var entries []json.RawMessage
	if raw, ok := hooks.get(e.event); ok {
		if err := json.Unmarshal(raw, &entries); err != nil {
			return fmt.Errorf("%s: hooks.%s is not a list: %w", rel, e.event, err)
		}
	}
	want, err := marshal(e.entry)
	if err != nil {
		return err
	}
	var kept []json.RawMessage
	placed := false
	for _, raw := range entries {
		if !e.mine(raw) {
			kept = append(kept, raw)
			continue
		}
		if e.enabled && !placed {
			kept = append(kept, want)
			placed = true
		}
	}
	if e.enabled && !placed {
		kept = append(kept, want)
	}
	if len(kept) == 0 {
		hooks.remove(e.event)
		return nil
	}
	list, err := marshal(kept)
	if err != nil {
		return err
	}
	hooks.set(e.event, list)
	return nil
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
