package v1

import (
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"
)

func TestSubjectAndEvents(t *testing.T) {
	if Subject != "cafe.discovery.scanners.presence" {
		t.Fatalf("subject = %q", Subject)
	}
	if EventJoined != "joined" || EventLeft != "left" {
		t.Fatalf("events %q %q", EventJoined, EventLeft)
	}
}

func TestPresenceVectors(t *testing.T) {
	entries, err := fs.ReadDir(Vectors, "testdata")
	if err != nil {
		t.Fatal(err)
	}
	var sawJoined, sawLeft, sawNone, sawUnknownType, sawInvalid, sawBad bool
	for _, entry := range entries {
		name := entry.Name()
		raw, err := fs.ReadFile(Vectors, "testdata/"+name)
		if err != nil {
			t.Fatal(err)
		}
		switch {
		case strings.HasPrefix(name, "ok_"):
			msg, err := Decode(raw)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			out, err := json.Marshal(msg)
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != strings.TrimSpace(string(raw)) {
				t.Fatalf("%s round trip:\n%s\nvs\n%s", name, out, raw)
			}
			switch name {
			case "ok_joined.json":
				sawJoined = true
				if msg.Event != EventJoined || msg.Indexer() != IndexerUnknown || msg.OnchainIndexer != "" {
					t.Fatalf("%+v indexer %s", msg, msg.Indexer())
				}
				if strings.Contains(string(out), "onchain_indexer") {
					t.Fatalf("absent field written: %s", out)
				}
			case "ok_left.json":
				sawLeft = true
				if msg.Event != EventLeft || msg.Type != "tls" {
					t.Fatalf("%+v", msg)
				}
			case "ok_joined_none.json":
				sawNone = true
				if msg.Indexer() != IndexerNone || msg.OnchainIndexer != IndexerNone {
					t.Fatalf("%+v indexer %s", msg, msg.Indexer())
				}
			case "ok_joined_etherscan.json":
				if msg.Indexer() != IndexerEtherscan {
					t.Fatalf("indexer %s", msg.Indexer())
				}
			case "ok_joined_moralis.json":
				if msg.Indexer() != IndexerMoralis {
					t.Fatalf("indexer %s", msg.Indexer())
				}
			case "ok_unknown_type.json":
				sawUnknownType = true
				if msg.Type != "indexer-gateway" || msg.Indexer() != IndexerUnknown {
					t.Fatalf("%+v indexer %s", msg, msg.Indexer())
				}
			case "ok_invalid_indexer.json":
				sawInvalid = true
				if msg.OnchainIndexer != "bogus" || msg.Indexer() != IndexerUnknown || msg.Indexer() == IndexerNone {
					t.Fatalf("field %q read %s", msg.OnchainIndexer, msg.Indexer())
				}
			default:
				t.Fatalf("unclassified ok fixture %s", name)
			}
		case strings.HasPrefix(name, "bad_"):
			sawBad = true
			if _, err := Decode(raw); err == nil || !errors.Is(err, ErrMalformed) {
				t.Fatalf("%s: got %v", name, err)
			}
		default:
			t.Fatalf("unclassified fixture %s", name)
		}
	}
	if !sawJoined || !sawLeft || !sawNone || !sawUnknownType || !sawInvalid || !sawBad {
		t.Fatalf("fixtures missing: joined=%v left=%v none=%v type=%v invalid=%v bad=%v",
			sawJoined, sawLeft, sawNone, sawUnknownType, sawInvalid, sawBad)
	}
}

func TestNullIndexerReadsUnknown(t *testing.T) {
	raw := []byte(`{"event":"joined","scanner_id":"scanner-1","type":"wallet","onchain_indexer":null}`)
	msg, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Indexer() != IndexerUnknown || msg.Indexer() == IndexerNone {
		t.Fatalf("indexer %s", msg.Indexer())
	}
}

func TestNoneIsDistinctFromAbsent(t *testing.T) {
	absent, err := fs.ReadFile(Vectors, "testdata/ok_joined.json")
	if err != nil {
		t.Fatal(err)
	}
	none, err := fs.ReadFile(Vectors, "testdata/ok_joined_none.json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(absent)) == strings.TrimSpace(string(none)) {
		t.Fatal("absent and none encode the same bytes")
	}
	absentMsg, err := Decode(absent)
	if err != nil {
		t.Fatal(err)
	}
	noneMsg, err := Decode(none)
	if err != nil {
		t.Fatal(err)
	}
	if absentMsg.Indexer() != IndexerUnknown || noneMsg.Indexer() != IndexerNone {
		t.Fatalf("absent %s none %s", absentMsg.Indexer(), noneMsg.Indexer())
	}
}

func TestMalformedPresence(t *testing.T) {
	bodies := []string{
		`null`,
		`[]`,
		`{`,
	}
	for _, body := range bodies {
		if _, err := Decode([]byte(body)); err == nil || !errors.Is(err, ErrMalformed) {
			t.Fatalf("%s: %v", body, err)
		}
	}
}
