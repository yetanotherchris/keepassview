package server_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"keepassview/internal/config"
	"keepassview/internal/server"
	"keepassview/internal/vault"
)

const (
	testDB   = "../../keepassview-test.kdbx"
	testPass = "keepassview-Test-database-1"
	// assets are at repo root; os.DirFS("../..") lets the server resolve
	// "assets/templates/…" and "assets/static/…" paths correctly.
	assetsRoot = "../.."
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	v, err := vault.Open(testDB, testPass)
	if err != nil {
		t.Fatalf("open vault: %v", err)
	}
	cfg := &config.Config{
		DBPath:             testDB,
		UseCredentialStore: false,
		PromptCharCount:    3,
		IdleTimeoutSeconds: 300,
	}
	mux, _ := server.NewMux(v, cfg, os.DirFS(assetsRoot))
	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return ts
}

func getJSON(t *testing.T, url string, dest any) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		t.Fatalf("decode %s: %v", url, err)
	}
	resp.Body.Close()
	return resp
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestIndex(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type: %q", ct)
	}
}

func TestAppJS(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Errorf("app.js Content-Type: %q", ct)
	}
}

func TestAppCSS(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/app.css")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "css") {
		t.Errorf("app.css Content-Type: %q", ct)
	}
}

func TestGroups(t *testing.T) {
	ts := newTestServer(t)
	var groups []map[string]any
	getJSON(t, ts.URL+"/api/groups", &groups)
	t.Logf("groups: %s", mustJSON(groups))
	if len(groups) == 0 {
		t.Fatal("expected at least one group")
	}
	// Root group should have the subgroup "notes" as a child
	root := groups[0]
	children, _ := root["children"].([]any)
	if len(children) == 0 {
		t.Error("expected Root to have child groups")
	}
}

func TestSearchAll(t *testing.T) {
	ts := newTestServer(t)
	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search", &entries)
	t.Logf("all entries (%d): %s", len(entries), mustJSON(entries))
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
	// Passwords must NOT appear in summaries
	for _, e := range entries {
		if _, ok := e["password"]; ok {
			t.Errorf("password leaked in summary for %v", e["title"])
		}
	}
}

func TestSearchByTitle(t *testing.T) {
	ts := newTestServer(t)
	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search?q=title", &entries)
	t.Logf("search 'title' → %d entries", len(entries))
	// "My title" and "my title2" should match; "my note" should not
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d: %s", len(entries), mustJSON(entries))
	}
}

func TestSearchByNotes(t *testing.T) {
	ts := newTestServer(t)
	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search?q=notes+live", &entries)
	t.Logf("search 'notes live' → %d entries", len(entries))
	if len(entries) != 1 {
		t.Errorf("expected 1 entry for notes search, got %d", len(entries))
	}
}

func TestSearchNoMatch(t *testing.T) {
	ts := newTestServer(t)
	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search?q=zzznomatch", &entries)
	if len(entries) != 0 {
		t.Errorf("expected 0 entries for no-match query, got %d", len(entries))
	}
}

func TestSearchGroupFilter(t *testing.T) {
	ts := newTestServer(t)

	var groups []map[string]any
	getJSON(t, ts.URL+"/api/groups", &groups)

	// Filter by root group UUID — should include subgroup entries too
	rootUUID := fmt.Sprint(groups[0]["uuid"])
	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search?group="+rootUUID, &entries)
	t.Logf("root group filter → %d entries", len(entries))
	if len(entries) != 3 {
		t.Errorf("expected 3 entries (including subgroup), got %d", len(entries))
	}
}

func TestSearchSubgroupFilter(t *testing.T) {
	ts := newTestServer(t)

	var groups []map[string]any
	getJSON(t, ts.URL+"/api/groups", &groups)

	children, _ := groups[0]["children"].([]any)
	if len(children) == 0 {
		t.Skip("no child groups")
	}
	subUUID := fmt.Sprint((children[0].(map[string]any))["uuid"])
	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search?group="+subUUID, &entries)
	t.Logf("subgroup 'notes' filter → %d entries", len(entries))
	if len(entries) != 1 {
		t.Errorf("expected 1 entry in notes subgroup, got %d", len(entries))
	}
}

func TestEntryDetail(t *testing.T) {
	ts := newTestServer(t)

	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search", &entries)

	for _, e := range entries {
		uuid := fmt.Sprint(e["uuid"])
		var detail map[string]any
		r := getJSON(t, ts.URL+"/api/entries/"+uuid, &detail)
		if r.StatusCode != http.StatusOK {
			t.Errorf("entry %s: status %d", uuid, r.StatusCode)
		}
		t.Logf("detail %q: pass=%q notes=%q", detail["title"], detail["password"], detail["notes"])
		if _, ok := detail["password"]; !ok {
			t.Errorf("password missing from detail for %v", detail["title"])
		}
	}
}

func TestEntryPasswordValues(t *testing.T) {
	ts := newTestServer(t)

	var entries []map[string]any
	getJSON(t, ts.URL+"/api/search", &entries)

	passwords := map[string]string{}
	for _, e := range entries {
		uuid := fmt.Sprint(e["uuid"])
		title := fmt.Sprint(e["title"])
		var detail map[string]any
		getJSON(t, ts.URL+"/api/entries/"+uuid, &detail)
		passwords[title] = fmt.Sprint(detail["password"])
	}

	if passwords["My title"] != "mypassword" {
		t.Errorf("My title: want password %q, got %q", "mypassword", passwords["My title"])
	}
	if passwords["my title2"] != "mypassword2" {
		t.Errorf("my title2: want password %q, got %q", "mypassword2", passwords["my title2"])
	}
}

func TestEntryNotFound(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/api/entries/doesnotexist")
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
