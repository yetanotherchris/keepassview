package vault

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tobischo/gokeepasslib/v3"
)

// Group represents a KeePass group node in the tree.
type Group struct {
	UUID     string  `json:"uuid"`
	Name     string  `json:"name"`
	Count    int     `json:"count"`
	Children []Group `json:"children,omitempty"`
}

// EntrySummary is a lightweight entry representation for list views.
type EntrySummary struct {
	UUID     string `json:"uuid"`
	Title    string `json:"title"`
	Username string `json:"username"`
	URL      string `json:"url"`
	Group    string `json:"group"`
	Icon     int    `json:"icon"`
}

// CustomField is a non-standard entry field.
type CustomField struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// AttachmentInfo describes a binary attachment (content download is deferred).
type AttachmentInfo struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// EntryDetail is a full entry including password and metadata.
type EntryDetail struct {
	EntrySummary
	Password     string           `json:"password"`
	Notes        string           `json:"notes"`
	Created      string           `json:"created,omitempty"`
	Modified     string           `json:"modified,omitempty"`
	CustomFields []CustomField    `json:"customFields,omitempty"`
	Attachments  []AttachmentInfo `json:"attachments,omitempty"`
}

type entryRecord struct {
	entry     gokeepasslib.Entry
	groupName string
}

// Vault holds the decrypted database in memory (immutable after Open).
type Vault struct {
	groups        []Group
	summaries     []EntrySummary
	summaryByUUID map[string]EntrySummary
	entries       map[string]entryRecord
	groupEntries  map[string][]string // group uuid -> all entry UUIDs in subtree
	db            *gokeepasslib.Database
}

// Open decrypts the KDBX file at path using password and builds the search index.
func Open(path, password string) (*Vault, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("database not found at %q; run keepassview init or pass --db", path)
		}
		return nil, fmt.Errorf("open database: %w", err)
	}
	defer f.Close()

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	if err := gokeepasslib.NewDecoder(f).Decode(db); err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password or corrupt file): %w", err)
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, fmt.Errorf("unlock entries: %w", err)
	}

	v := &Vault{
		summaryByUUID: make(map[string]EntrySummary),
		entries:       make(map[string]entryRecord),
		groupEntries:  make(map[string][]string),
		db:            db,
	}

	if db.Content != nil && db.Content.Root != nil {
		for _, g := range db.Content.Root.Groups {
			group, _ := v.processGroup(g)
			v.groups = append(v.groups, group)
		}
	}

	return v, nil
}

func uuidToString(u gokeepasslib.UUID) string {
	return base64.RawURLEncoding.EncodeToString(u[:])
}

func (v *Vault) processGroup(g gokeepasslib.Group) (Group, []string) {
	groupUUID := uuidToString(g.UUID)
	var allUUIDs []string

	for _, entry := range g.Entries {
		uuid := uuidToString(entry.UUID)
		sum := EntrySummary{
			UUID:     uuid,
			Title:    entry.GetTitle(),
			Username: entry.GetContent("UserName"),
			URL:      entry.GetContent("URL"),
			Group:    g.Name,
			Icon:     int(entry.IconID),
		}
		v.summaries = append(v.summaries, sum)
		v.summaryByUUID[uuid] = sum
		v.entries[uuid] = entryRecord{entry: entry, groupName: g.Name}
		allUUIDs = append(allUUIDs, uuid)
	}

	var children []Group
	for _, child := range g.Groups {
		childGroup, childUUIDs := v.processGroup(child)
		children = append(children, childGroup)
		allUUIDs = append(allUUIDs, childUUIDs...)
	}

	v.groupEntries[groupUUID] = allUUIDs

	return Group{
		UUID:     groupUUID,
		Name:     g.Name,
		Count:    len(allUUIDs),
		Children: children,
	}, allUUIDs
}

// Groups returns the root-level group tree.
func (v *Vault) Groups() []Group { return v.groups }

// Search returns entry summaries matching query q, optionally filtered to a group subtree.
// groupUUID="" means all entries. q="" means no text filter.
func (v *Vault) Search(q, groupUUID string) []EntrySummary {
	q = strings.ToLower(strings.TrimSpace(q))

	var candidates []EntrySummary
	if groupUUID == "" {
		candidates = v.summaries
	} else {
		uuids := v.groupEntries[groupUUID]
		for _, uid := range uuids {
			if s, ok := v.summaryByUUID[uid]; ok {
				candidates = append(candidates, s)
			}
		}
	}

	if q == "" {
		return candidates
	}

	var results []EntrySummary
	for _, s := range candidates {
		rec := v.entries[s.UUID]
		notes := strings.ToLower(rec.entry.GetContent("Notes"))
		if strings.Contains(strings.ToLower(s.Title), q) ||
			strings.Contains(strings.ToLower(s.Username), q) ||
			strings.Contains(strings.ToLower(s.URL), q) ||
			strings.Contains(notes, q) {
			results = append(results, s)
		}
	}
	return results
}

// GetEntry returns the full detail for an entry identified by base64url UUID.
func (v *Vault) GetEntry(uuid string) (*EntryDetail, error) {
	rec, ok := v.entries[uuid]
	if !ok {
		return nil, fmt.Errorf("entry %q not found", uuid)
	}

	entry := rec.entry
	detail := &EntryDetail{
		EntrySummary: EntrySummary{
			UUID:     uuid,
			Title:    entry.GetTitle(),
			Username: entry.GetContent("UserName"),
			URL:      entry.GetContent("URL"),
			Group:    rec.groupName,
			Icon:     int(entry.IconID),
		},
		Password: entry.GetPassword(),
		Notes:    entry.GetContent("Notes"),
	}

	if t := entry.Times.CreationTime; t != nil {
		detail.Created = t.Time.Format(time.RFC3339)
	}
	if t := entry.Times.LastModificationTime; t != nil {
		detail.Modified = t.Time.Format(time.RFC3339)
	}

	known := map[string]bool{
		"Title": true, "UserName": true, "Password": true,
		"URL": true, "Notes": true,
	}
	for _, vd := range entry.Values {
		if !known[vd.Key] {
			detail.CustomFields = append(detail.CustomFields, CustomField{
				Key:   vd.Key,
				Value: vd.Value.Content,
			})
		}
	}

	for _, binRef := range entry.Binaries {
		size := v.binarySize(binRef.Value.ID)
		detail.Attachments = append(detail.Attachments, AttachmentInfo{
			Name: binRef.Name,
			Size: size,
		})
	}

	return detail, nil
}

func (v *Vault) binarySize(id int) int {
	b := v.db.FindBinary(id)
	if b == nil {
		return 0
	}
	data, err := b.GetContentBytes()
	if err != nil {
		return 0
	}
	return len(data)
}
