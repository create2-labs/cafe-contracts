package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestPersistenceReady(t *testing.T) {
	if SubjectPersistenceReady != "persistence.ready" {
		t.Fatal(SubjectPersistenceReady)
	}
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawEmpty, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case name == "ok_empty.json":
			sawEmpty = true
			msg, err := Decode(raw)
			if err != nil {
				t.Fatal(err)
			}
			out, err := json.Marshal(msg)
			if err != nil || string(out) != strings.TrimSpace(string(raw)) || string(out) != "{}" {
				t.Fatalf("%s %v", out, err)
			}
		case strings.HasPrefix(name, "bad_"):
			sawBad = true
			reject(t, name, raw)
		default:
			t.Fatalf("unclassified fixture %s", name)
		}
	}
	if !sawEmpty || !sawBad {
		t.Fatal(sawEmpty, sawBad)
	}

	out, err := json.Marshal(Ready{})
	if err != nil || string(out) != "{}" {
		t.Fatalf("%s %v", out, err)
	}
	msg, err := Decode([]byte(" \n{ \"ignored\": 1 }\n "))
	if err != nil {
		t.Fatal(err)
	}
	out, err = json.Marshal(msg)
	if err != nil || string(out) != "{}" {
		t.Fatalf("%s %v", out, err)
	}
	const token = "body-must-not-leak"
	err = decodeErr([]byte(token))
	if err == nil || !errors.Is(err, ErrMalformed) || strings.Contains(err.Error(), token) {
		t.Fatal(err)
	}
	for _, body := range []string{"", " ", "true", "0", `"ready"`, "null", "[]", "{", "{}{}", "{}}"} {
		reject(t, "shape", []byte(body))
	}
}

func reject(t *testing.T, name string, raw []byte) {
	t.Helper()
	if err := decodeErr(raw); err == nil || !errors.Is(err, ErrMalformed) {
		t.Fatalf("%s: %v", name, err)
	}
}

func decodeErr(raw []byte) error {
	_, err := Decode(raw)
	return err
}
