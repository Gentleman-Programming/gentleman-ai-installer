package backup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/internal/components/filemerge"
)

type RestoreService struct{}

func (s RestoreService) Restore(manifest Manifest) error {
	for _, entry := range manifest.Entries {
		if entry.Existed {
			if err := restoreEntry(entry); err != nil {
				return err
			}
			continue
		}

		if err := os.Remove(entry.OriginalPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove path %q: %w", entry.OriginalPath, err)
		}
	}

	return nil
}

func restoreEntry(entry ManifestEntry) error {
	content, err := os.ReadFile(entry.SnapshotPath)
	if err != nil {
		return fmt.Errorf("read snapshot file %q: %w", entry.SnapshotPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(entry.OriginalPath), 0o755); err != nil {
		return fmt.Errorf("create restore directory for %q: %w", entry.OriginalPath, err)
	}

	if _, err := filemerge.WriteFileAtomic(entry.OriginalPath, content, os.FileMode(entry.Mode)); err != nil {
		return fmt.Errorf("restore path %q: %w", entry.OriginalPath, err)
	}

	return nil
}
