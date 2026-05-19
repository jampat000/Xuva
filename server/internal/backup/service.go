package backup

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const manifestVersion = 1

// Manifest is written into every archive so the importer knows what it got.
type Manifest struct {
	Version    int    `json:"version"`
	CreatedAt  string `json:"createdAt"`
	DataDir    string `json:"dataDir"`
	MediaPaths struct {
		Movies string `json:"movies"`
		TV     string `json:"tv"`
	} `json:"mediaPaths"`
}

// Service wraps a database connection and the data directory for backup operations.
type Service struct {
	DataDir string
	DB      *sql.DB
}

// New creates a Service.
func New(dataDir string, db *sql.DB) *Service {
	return &Service{DataDir: dataDir, DB: db}
}

// Export writes a gzip-compressed tar archive to w containing:
//   - manifest.json — metadata about the backup
//   - xuva.db       — a consistent snapshot via VACUUM INTO
//   - settings.json — server settings (if present)
func (s *Service) Export(w io.Writer, moviesPath, tvPath string) error {
	tmp, err := os.MkdirTemp("", "xuva-backup-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	// Consistent SQLite snapshot without locking writes for long
	snapPath := filepath.Join(tmp, "xuva.db")
	if _, err := s.DB.Exec("VACUUM INTO ?", snapPath); err != nil {
		return fmt.Errorf("snapshot db: %w", err)
	}

	// Manifest
	m := Manifest{
		Version:   manifestVersion,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		DataDir:   s.DataDir,
	}
	m.MediaPaths.Movies = moviesPath
	m.MediaPaths.TV = tvPath
	manifestData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "manifest.json"), manifestData, 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	// settings.json (best-effort)
	_ = copyFileIfExists(
		filepath.Join(s.DataDir, "settings.json"),
		filepath.Join(tmp, "settings.json"),
	)

	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)
	for _, name := range []string{"manifest.json", "xuva.db", "settings.json"} {
		p := filepath.Join(tmp, name)
		if err := addFileToTar(tw, p, name); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("archive %s: %w", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// Stage validates an uploaded archive and copies the restored DB into the data
// directory as xuva-restore.db, then writes restore-pending.json.  On the next
// server restart, ApplyIfPending will swap the files into place.
func (s *Service) Stage(r io.Reader) (*Manifest, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close()
	tr := tar.NewReader(gr)

	tmp, err := os.MkdirTemp("", "xuva-import-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tmp)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}
		name := filepath.Base(hdr.Name)
		if name == "" || name == "." {
			continue
		}
		// Only extract known safe files
		switch name {
		case "manifest.json", "xuva.db", "settings.json":
		default:
			continue
		}
		dst := filepath.Join(tmp, name)
		f, err := os.Create(dst)
		if err != nil {
			return nil, err
		}
		if _, err := io.CopyN(f, tr, hdr.Size); err != nil && err != io.EOF {
			f.Close()
			return nil, fmt.Errorf("extract %s: %w", name, err)
		}
		f.Close()
	}

	// Validate manifest
	manifestData, err := os.ReadFile(filepath.Join(tmp, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("manifest.json missing from archive")
	}
	var m Manifest
	if err := json.Unmarshal(manifestData, &m); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}
	if m.Version != manifestVersion {
		return nil, fmt.Errorf("unsupported backup version %d (expected %d)", m.Version, manifestVersion)
	}
	if _, err := os.Stat(filepath.Join(tmp, "xuva.db")); err != nil {
		return nil, fmt.Errorf("xuva.db missing from archive")
	}

	// Move files to staging area
	if err := os.MkdirAll(s.DataDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.Rename(filepath.Join(tmp, "xuva.db"), filepath.Join(s.DataDir, "xuva-restore.db")); err != nil {
		return nil, fmt.Errorf("stage db: %w", err)
	}
	if err := os.WriteFile(filepath.Join(s.DataDir, "restore-pending.json"), manifestData, 0o644); err != nil {
		return nil, fmt.Errorf("write pending marker: %w", err)
	}
	// Optional settings restore
	settingsSrc := filepath.Join(tmp, "settings.json")
	if fi, err := os.Stat(settingsSrc); err == nil && fi.Size() > 0 {
		_ = os.Rename(settingsSrc, filepath.Join(s.DataDir, "settings-restore.json"))
	}

	return &m, nil
}

// ApplyIfPending checks whether a staged restore is waiting and, if so, swaps
// the database files before the main database is opened.  Returns true when a
// restore was applied and the caller should proceed with the fresh DB.
func ApplyIfPending(dataDir string) (bool, error) {
	pendingPath := filepath.Join(dataDir, "restore-pending.json")
	if _, err := os.Stat(pendingPath); os.IsNotExist(err) {
		return false, nil
	}

	restoreDB := filepath.Join(dataDir, "xuva-restore.db")
	if _, err := os.Stat(restoreDB); err != nil {
		_ = os.Remove(pendingPath)
		return false, nil
	}

	liveDB := filepath.Join(dataDir, "xuva.db")
	backupDB := filepath.Join(dataDir, "xuva.db.bak")

	if _, err := os.Stat(liveDB); err == nil {
		if err := os.Rename(liveDB, backupDB); err != nil {
			return false, fmt.Errorf("backup current db: %w", err)
		}
	}
	if err := os.Rename(restoreDB, liveDB); err != nil {
		// Attempt rollback
		_ = os.Rename(backupDB, liveDB)
		return false, fmt.Errorf("apply restore db: %w", err)
	}

	restoreSettings := filepath.Join(dataDir, "settings-restore.json")
	if _, err := os.Stat(restoreSettings); err == nil {
		_ = os.Rename(restoreSettings, filepath.Join(dataDir, "settings.json"))
	}

	_ = os.Remove(pendingPath)
	return true, nil
}

func copyFileIfExists(src, dst string) error {
	f, err := os.Open(src)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, f)
	return err
}

func addFileToTar(tw *tar.Writer, path, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	hdr := &tar.Header{
		Name:    name,
		Mode:    0o644,
		Size:    fi.Size(),
		ModTime: fi.ModTime(),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}
	_, err = io.Copy(tw, f)
	return err
}
