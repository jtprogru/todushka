// Package bbolt implements storage.Repository on top of go.etcd.io/bbolt.
package bbolt

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jtprogru/todushka/internal/domain/area"
	"github.com/jtprogru/todushka/internal/domain/id"
	"github.com/jtprogru/todushka/internal/domain/project"
	"github.com/jtprogru/todushka/internal/domain/tag"
	"github.com/jtprogru/todushka/internal/domain/task"
	"github.com/jtprogru/todushka/internal/storage"
	bolt "go.etcd.io/bbolt"
	bolterrors "go.etcd.io/bbolt/errors"
)

const (
	openTimeout = 200 * time.Millisecond

	bktMeta          = "meta"
	bktTasks         = "tasks"
	bktProjects      = "projects"
	bktHeadings      = "headings"
	bktAreas         = "areas"
	bktTags          = "tags"
	bktIdxTagByNorm  = "idx_tag_by_normalized"
	bktIdxAreaByNorm = "idx_area_by_normalized"

	keySchemaVersion = "schema_version"
)

var allBuckets = []string{
	bktMeta, bktTasks, bktProjects, bktHeadings, bktAreas, bktTags,
	bktIdxTagByNorm, bktIdxAreaByNorm,
}

type Repo struct {
	db   *bolt.DB
	path string
}

// Open opens (or creates) a bbolt database file at path and runs all pending
// migrations up to storage.CurrentSchemaVersion.
//
// If another process holds the file lock, returns storage.ErrDatabaseLocked
// after openTimeout.
func Open(path string) (*Repo, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("bbolt: mkdir: %w", err)
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: openTimeout})
	if err != nil {
		if errors.Is(err, bolterrors.ErrTimeout) {
			return nil, storage.ErrDatabaseLocked
		}
		return nil, fmt.Errorf("bbolt: open: %w", err)
	}
	r := &Repo{db: db, path: path}
	if err := r.ensureBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	current, err := r.SchemaVersion(context.Background())
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	if current > storage.CurrentSchemaVersion {
		_ = db.Close()
		return nil, storage.ErrSchemaTooNew
	}
	if err := r.Migrate(context.Background(), storage.CurrentSchemaVersion); err != nil {
		_ = db.Close()
		return nil, err
	}
	return r, nil
}

func (r *Repo) Close() error { return r.db.Close() }

func (r *Repo) ensureBuckets() error {
	return r.db.Update(func(tx *bolt.Tx) error {
		for _, name := range allBuckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(name)); err != nil {
				return fmt.Errorf("bbolt: create bucket %s: %w", name, err)
			}
		}
		return nil
	})
}

// Schema / migration

func (r *Repo) SchemaVersion(_ context.Context) (int, error) {
	var v uint32
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktMeta))
		if bkt == nil {
			return nil
		}
		raw := bkt.Get([]byte(keySchemaVersion))
		if raw == nil {
			return nil
		}
		if len(raw) != 4 {
			return fmt.Errorf("bbolt: malformed schema_version (len=%d)", len(raw))
		}
		v = binary.BigEndian.Uint32(raw)
		return nil
	})
	return int(v), err
}

func (r *Repo) Migrate(_ context.Context, target int) error {
	if target > storage.CurrentSchemaVersion {
		return storage.ErrSchemaTooNew
	}
	current, err := r.SchemaVersion(context.Background())
	if err != nil {
		return err
	}
	for v := current + 1; v <= target; v++ {
		mig, ok := migrationByVersion(v)
		if !ok {
			return fmt.Errorf("bbolt: no migration for version %d", v)
		}
		if err := r.db.Update(func(tx *bolt.Tx) error {
			if err := mig.Apply(tx); err != nil {
				return err
			}
			return writeSchemaVersion(tx, v)
		}); err != nil {
			return err
		}
	}
	return nil
}

func writeSchemaVersion(tx *bolt.Tx, v int) error {
	bkt := tx.Bucket([]byte(bktMeta))
	buf := make([]byte, 4)
	if v < 0 {
		return fmt.Errorf("bbolt: invalid schema version %d", v)
	}
	binary.BigEndian.PutUint32(buf, uint32(v)) //nolint:gosec // bounded by check above
	return bkt.Put([]byte(keySchemaVersion), buf)
}

// JSON helpers

func putJSON[T any](bkt *bolt.Bucket, key []byte, v T) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("bbolt: marshal: %w", err)
	}
	return bkt.Put(key, data)
}

func getJSON[T any](bkt *bolt.Bucket, key []byte) (T, bool, error) {
	var zero T
	raw := bkt.Get(key)
	if raw == nil {
		return zero, false, nil
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, false, fmt.Errorf("bbolt: unmarshal: %w", err)
	}
	return out, true, nil
}

// Tasks

func (r *Repo) TaskCreate(_ context.Context, t task.Task) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTasks))
		if bkt.Get([]byte(t.ID)) != nil {
			return storage.ErrAlreadyExists
		}
		return putJSON(bkt, []byte(t.ID), t)
	})
}

func (r *Repo) TaskGet(_ context.Context, taskID id.ID) (task.Task, error) {
	var out task.Task
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTasks))
		v, found, err := getJSON[task.Task](bkt, []byte(taskID))
		if err != nil {
			return err
		}
		if !found {
			return storage.ErrNotFound
		}
		out = v
		return nil
	})
	return out, err
}

func (r *Repo) TaskUpdate(_ context.Context, t task.Task) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTasks))
		if bkt.Get([]byte(t.ID)) == nil {
			return storage.ErrNotFound
		}
		return putJSON(bkt, []byte(t.ID), t)
	})
}

func (r *Repo) TaskDelete(_ context.Context, taskID id.ID, soft bool) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTasks))
		raw := bkt.Get([]byte(taskID))
		if raw == nil {
			return storage.ErrNotFound
		}
		if !soft {
			return bkt.Delete([]byte(taskID))
		}
		var t task.Task
		if err := json.Unmarshal(raw, &t); err != nil {
			return err
		}
		now := time.Now()
		t.DeletedAt = &now
		return putJSON(bkt, []byte(taskID), t)
	})
}

func (r *Repo) TaskList(_ context.Context, f storage.TaskFilter) ([]task.Task, error) {
	out := make([]task.Task, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTasks))
		return bkt.ForEach(func(_, v []byte) error {
			var t task.Task
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			if !f.IncludeDeleted && t.DeletedAt != nil {
				return nil
			}
			if taskMatches(t, f) {
				out = append(out, t)
			}
			return nil
		})
	})
	return out, err
}

func taskMatches(t task.Task, f storage.TaskFilter) bool {
	if len(f.Statuses) > 0 {
		ok := false
		for _, s := range f.Statuses {
			if t.Status == s {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	if f.AreaID != nil && (t.AreaID == nil || *t.AreaID != *f.AreaID) {
		return false
	}
	if f.ProjectID != nil && (t.ProjectID == nil || *t.ProjectID != *f.ProjectID) {
		return false
	}
	if f.HeadingID != nil && (t.HeadingID == nil || *t.HeadingID != *f.HeadingID) {
		return false
	}
	if f.Someday != nil && t.Someday != *f.Someday {
		return false
	}
	if f.HasStartDate != nil {
		has := t.StartDate != nil
		if has != *f.HasStartDate {
			return false
		}
	}
	if len(f.TagsAll) > 0 {
		for _, want := range f.TagsAll {
			found := false
			for _, have := range t.Tags {
				if have == want {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (r *Repo) TaskMatchShort(_ context.Context, prefix string) ([]task.Task, error) {
	pfx := strings.ToUpper(prefix)
	out := make([]task.Task, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTasks))
		return bkt.ForEach(func(k, v []byte) error {
			if !idMatchesShort(string(k), pfx) {
				return nil
			}
			var t task.Task
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			out = append(out, t)
			return nil
		})
	})
	return out, err
}

func idMatchesShort(full, pfx string) bool {
	upper := strings.ToUpper(full)
	// Exact full-ID match
	if upper == pfx {
		return true
	}
	// Random-portion prefix match: chars 10..len(pfx)+10
	const offset = 10
	if len(upper) < offset+len(pfx) {
		return false
	}
	return upper[offset:offset+len(pfx)] == pfx
}

// Projects

func (r *Repo) ProjectCreate(_ context.Context, p project.Project) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktProjects))
		if bkt.Get([]byte(p.ID)) != nil {
			return storage.ErrAlreadyExists
		}
		return putJSON(bkt, []byte(p.ID), p)
	})
}

func (r *Repo) ProjectGet(_ context.Context, pid id.ID) (project.Project, error) {
	var out project.Project
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktProjects))
		v, found, err := getJSON[project.Project](bkt, []byte(pid))
		if err != nil {
			return err
		}
		if !found {
			return storage.ErrNotFound
		}
		out = v
		return nil
	})
	return out, err
}

func (r *Repo) ProjectUpdate(_ context.Context, p project.Project) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktProjects))
		if bkt.Get([]byte(p.ID)) == nil {
			return storage.ErrNotFound
		}
		return putJSON(bkt, []byte(p.ID), p)
	})
}

func (r *Repo) ProjectDelete(_ context.Context, pid id.ID, soft bool) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktProjects))
		raw := bkt.Get([]byte(pid))
		if raw == nil {
			return storage.ErrNotFound
		}
		if !soft {
			return bkt.Delete([]byte(pid))
		}
		var p project.Project
		if err := json.Unmarshal(raw, &p); err != nil {
			return err
		}
		now := time.Now()
		p.DeletedAt = &now
		return putJSON(bkt, []byte(pid), p)
	})
}

func (r *Repo) ProjectList(_ context.Context, f storage.ProjectFilter) ([]project.Project, error) {
	out := make([]project.Project, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktProjects))
		return bkt.ForEach(func(_, v []byte) error {
			var p project.Project
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			if !f.IncludeDeleted && p.DeletedAt != nil {
				return nil
			}
			if f.AreaID != nil && (p.AreaID == nil || *p.AreaID != *f.AreaID) {
				return nil
			}
			if len(f.Statuses) > 0 {
				match := false
				for _, s := range f.Statuses {
					if p.Status == s {
						match = true
						break
					}
				}
				if !match {
					return nil
				}
			}
			out = append(out, p)
			return nil
		})
	})
	return out, err
}

func (r *Repo) ProjectFindByName(_ context.Context, name string) ([]project.Project, error) {
	target := strings.ToLower(strings.TrimSpace(name))
	out := make([]project.Project, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktProjects))
		return bkt.ForEach(func(_, v []byte) error {
			var p project.Project
			if err := json.Unmarshal(v, &p); err != nil {
				return err
			}
			if p.DeletedAt != nil {
				return nil
			}
			if strings.ToLower(strings.TrimSpace(p.Name)) == target {
				out = append(out, p)
			}
			return nil
		})
	})
	return out, err
}

func (r *Repo) HeadingCreate(_ context.Context, h project.Heading) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktHeadings))
		if bkt.Get([]byte(h.ID)) != nil {
			return storage.ErrAlreadyExists
		}
		return putJSON(bkt, []byte(h.ID), h)
	})
}

func (r *Repo) HeadingUpdate(_ context.Context, h project.Heading) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktHeadings))
		if bkt.Get([]byte(h.ID)) == nil {
			return storage.ErrNotFound
		}
		return putJSON(bkt, []byte(h.ID), h)
	})
}

func (r *Repo) HeadingDelete(_ context.Context, hid id.ID) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktHeadings))
		if bkt.Get([]byte(hid)) == nil {
			return storage.ErrNotFound
		}
		return bkt.Delete([]byte(hid))
	})
}

func (r *Repo) HeadingList(_ context.Context, projectID id.ID) ([]project.Heading, error) {
	out := make([]project.Heading, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktHeadings))
		return bkt.ForEach(func(_, v []byte) error {
			var h project.Heading
			if err := json.Unmarshal(v, &h); err != nil {
				return err
			}
			if h.ProjectID == projectID {
				out = append(out, h)
			}
			return nil
		})
	})
	return out, err
}

// Areas

func (r *Repo) AreaCreate(_ context.Context, a area.Area) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktAreas))
		idx := tx.Bucket([]byte(bktIdxAreaByNorm))
		norm := area.Normalize(a.Name)
		if existing := idx.Get([]byte(norm)); existing != nil {
			return storage.ErrAlreadyExists
		}
		if err := putJSON(bkt, []byte(a.ID), a); err != nil {
			return err
		}
		return idx.Put([]byte(norm), []byte(a.ID))
	})
}

func (r *Repo) AreaGet(_ context.Context, aid id.ID) (area.Area, error) {
	var out area.Area
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktAreas))
		v, found, err := getJSON[area.Area](bkt, []byte(aid))
		if err != nil {
			return err
		}
		if !found {
			return storage.ErrNotFound
		}
		out = v
		return nil
	})
	return out, err
}

func (r *Repo) AreaList(_ context.Context) ([]area.Area, error) {
	out := make([]area.Area, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktAreas))
		return bkt.ForEach(func(_, v []byte) error {
			var a area.Area
			if err := json.Unmarshal(v, &a); err != nil {
				return err
			}
			if a.DeletedAt != nil {
				return nil
			}
			out = append(out, a)
			return nil
		})
	})
	return out, err
}

func (r *Repo) AreaUpdate(_ context.Context, a area.Area) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktAreas))
		idx := tx.Bucket([]byte(bktIdxAreaByNorm))
		oldRaw := bkt.Get([]byte(a.ID))
		if oldRaw == nil {
			return storage.ErrNotFound
		}
		var old area.Area
		if err := json.Unmarshal(oldRaw, &old); err != nil {
			return err
		}
		newNorm := area.Normalize(a.Name)
		oldNorm := area.Normalize(old.Name)
		if newNorm != oldNorm {
			if existing := idx.Get([]byte(newNorm)); existing != nil && string(existing) != string(a.ID) {
				return storage.ErrAlreadyExists
			}
			_ = idx.Delete([]byte(oldNorm))
			if err := idx.Put([]byte(newNorm), []byte(a.ID)); err != nil {
				return err
			}
		}
		return putJSON(bkt, []byte(a.ID), a)
	})
}

func (r *Repo) AreaDelete(_ context.Context, aid id.ID, soft bool) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktAreas))
		idx := tx.Bucket([]byte(bktIdxAreaByNorm))
		raw := bkt.Get([]byte(aid))
		if raw == nil {
			return storage.ErrNotFound
		}
		var a area.Area
		if err := json.Unmarshal(raw, &a); err != nil {
			return err
		}
		if soft {
			now := time.Now()
			a.DeletedAt = &now
			_ = idx.Delete([]byte(area.Normalize(a.Name)))
			return putJSON(bkt, []byte(aid), a)
		}
		_ = idx.Delete([]byte(area.Normalize(a.Name)))
		return bkt.Delete([]byte(aid))
	})
}

func (r *Repo) AreaFindByNormalized(_ context.Context, normalized string) (area.Area, error) {
	var out area.Area
	err := r.db.View(func(tx *bolt.Tx) error {
		idx := tx.Bucket([]byte(bktIdxAreaByNorm))
		aid := idx.Get([]byte(normalized))
		if aid == nil {
			return storage.ErrNotFound
		}
		bkt := tx.Bucket([]byte(bktAreas))
		v, found, err := getJSON[area.Area](bkt, aid)
		if err != nil {
			return err
		}
		if !found {
			return storage.ErrNotFound
		}
		out = v
		return nil
	})
	return out, err
}

// Tags

func (r *Repo) TagCreate(_ context.Context, t tag.Tag) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTags))
		idx := tx.Bucket([]byte(bktIdxTagByNorm))
		if bkt.Get([]byte(t.ID)) != nil {
			return storage.ErrAlreadyExists
		}
		if existing := idx.Get([]byte(t.Normalized)); existing != nil {
			return storage.ErrAlreadyExists
		}
		if err := putJSON(bkt, []byte(t.ID), t); err != nil {
			return err
		}
		return idx.Put([]byte(t.Normalized), []byte(t.ID))
	})
}

func (r *Repo) TagUpsert(_ context.Context, name string) (tag.Tag, error) {
	norm := tag.Normalize(name)
	if norm == "" {
		return tag.Tag{}, tag.ErrEmptyName
	}
	var out tag.Tag
	err := r.db.Update(func(tx *bolt.Tx) error {
		idx := tx.Bucket([]byte(bktIdxTagByNorm))
		bkt := tx.Bucket([]byte(bktTags))
		if existingID := idx.Get([]byte(norm)); existingID != nil {
			v, _, err := getJSON[tag.Tag](bkt, existingID)
			if err != nil {
				return err
			}
			out = v
			return nil
		}
		t := tag.Tag{ID: id.New(), Name: strings.TrimSpace(name), Normalized: norm, CreatedAt: time.Now()}
		if err := putJSON(bkt, []byte(t.ID), t); err != nil {
			return err
		}
		if err := idx.Put([]byte(norm), []byte(t.ID)); err != nil {
			return err
		}
		out = t
		return nil
	})
	return out, err
}

func (r *Repo) TagGet(_ context.Context, tid id.ID) (tag.Tag, error) {
	var out tag.Tag
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTags))
		v, found, err := getJSON[tag.Tag](bkt, []byte(tid))
		if err != nil {
			return err
		}
		if !found {
			return storage.ErrNotFound
		}
		out = v
		return nil
	})
	return out, err
}

func (r *Repo) TagList(_ context.Context) ([]tag.Tag, error) {
	out := make([]tag.Tag, 0)
	err := r.db.View(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTags))
		return bkt.ForEach(func(_, v []byte) error {
			var t tag.Tag
			if err := json.Unmarshal(v, &t); err != nil {
				return err
			}
			out = append(out, t)
			return nil
		})
	})
	return out, err
}

func (r *Repo) TagRename(_ context.Context, tid id.ID, newName string) error {
	newNorm := tag.Normalize(newName)
	if newNorm == "" {
		return tag.ErrEmptyName
	}
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTags))
		idx := tx.Bucket([]byte(bktIdxTagByNorm))
		raw := bkt.Get([]byte(tid))
		if raw == nil {
			return storage.ErrNotFound
		}
		var t tag.Tag
		if err := json.Unmarshal(raw, &t); err != nil {
			return err
		}
		oldNorm := t.Normalized
		if newNorm != oldNorm {
			if existing := idx.Get([]byte(newNorm)); existing != nil && string(existing) != string(tid) {
				return storage.ErrAlreadyExists
			}
			_ = idx.Delete([]byte(oldNorm))
			if err := idx.Put([]byte(newNorm), []byte(tid)); err != nil {
				return err
			}
		}
		t.Name = strings.TrimSpace(newName)
		t.Normalized = newNorm
		return putJSON(bkt, []byte(tid), t)
	})
}

func (r *Repo) TagDelete(_ context.Context, tid id.ID) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(bktTags))
		idx := tx.Bucket([]byte(bktIdxTagByNorm))
		raw := bkt.Get([]byte(tid))
		if raw == nil {
			return storage.ErrNotFound
		}
		var t tag.Tag
		if err := json.Unmarshal(raw, &t); err != nil {
			return err
		}
		_ = idx.Delete([]byte(t.Normalized))
		if err := bkt.Delete([]byte(tid)); err != nil {
			return err
		}
		// Strip tag references from all tasks
		tasksBkt := tx.Bucket([]byte(bktTasks))
		return tasksBkt.ForEach(func(k, v []byte) error {
			var tk task.Task
			if err := json.Unmarshal(v, &tk); err != nil {
				return err
			}
			filtered := tk.Tags[:0]
			changed := false
			for _, x := range tk.Tags {
				if x == tid {
					changed = true
					continue
				}
				filtered = append(filtered, x)
			}
			if changed {
				tk.Tags = filtered
				return putJSON(tasksBkt, k, tk)
			}
			return nil
		})
	})
}
