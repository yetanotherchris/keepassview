package vault_test

import (
	"testing"

	"keepassview/internal/vault"
)

const (
	testDB   = "../../keepassview-test.kdbx"
	testPass = "keepassview-Test-database-1"
)

func openTestVault(t *testing.T) *vault.Vault {
	t.Helper()
	v, err := vault.Open(testDB, testPass)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	return v
}

func TestOpen(t *testing.T) {
	v := openTestVault(t)
	groups := v.Groups()
	if len(groups) == 0 {
		t.Fatal("expected at least one group")
	}
	t.Logf("root groups: %d", len(groups))
	for _, g := range groups {
		t.Logf("  group %q  count=%d  children=%d", g.Name, g.Count, len(g.Children))
		for _, c := range g.Children {
			t.Logf("    child %q  count=%d", c.Name, c.Count)
		}
	}
}

func TestSearchAll(t *testing.T) {
	v := openTestVault(t)
	results := v.Search("", "")
	t.Logf("total entries: %d", len(results))
	for _, e := range results {
		t.Logf("  [%s] %q  user=%q  group=%q", e.UUID, e.Title, e.Username, e.Group)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one entry")
	}
}

func TestSearchQuery(t *testing.T) {
	v := openTestVault(t)
	all := v.Search("", "")
	if len(all) == 0 {
		t.Skip("no entries")
	}
	// Search for part of the first entry's title
	title := all[0].Title
	if len(title) < 3 {
		t.Skip("title too short to search")
	}
	q := title[:3]
	results := v.Search(q, "")
	t.Logf("search %q → %d results", q, len(results))
	found := false
	for _, r := range results {
		if r.UUID == all[0].UUID {
			found = true
		}
	}
	if !found {
		t.Errorf("first entry not found in search results for %q", q)
	}
}

func TestGroupFilter(t *testing.T) {
	v := openTestVault(t)
	groups := v.Groups()
	if len(groups) == 0 {
		t.Skip("no groups")
	}

	// Search within the first root group by its UUID
	g := groups[0]
	results := v.Search("", g.UUID)
	t.Logf("group %q (uuid=%s) → %d entries", g.Name, g.UUID, len(results))
}

func TestGetEntry(t *testing.T) {
	v := openTestVault(t)
	all := v.Search("", "")
	if len(all) == 0 {
		t.Skip("no entries")
	}

	for _, sum := range all {
		detail, err := v.GetEntry(sum.UUID)
		if err != nil {
			t.Errorf("GetEntry %s: %v", sum.UUID, err)
			continue
		}
		t.Logf("entry %q: user=%q  pass=%q  url=%q  notes=%q  customFields=%d",
			detail.Title, detail.Username, detail.Password,
			detail.URL, detail.Notes, len(detail.CustomFields))
		if detail.Title != sum.Title {
			t.Errorf("title mismatch: summary=%q detail=%q", sum.Title, detail.Title)
		}
	}
}

func TestWrongPassword(t *testing.T) {
	_, err := vault.Open(testDB, "wrong-password")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	t.Logf("correctly rejected wrong password: %v", err)
}
