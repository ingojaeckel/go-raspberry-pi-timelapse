package conf

import (
	"os"
	"path/filepath"
	"testing"

	bolt "go.etcd.io/bbolt"
)

// writeLegacy writes value the way versions before the JSON file saved it.
func writeLegacy(t *testing.T, path, value string) {
	t.Helper()
	db, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("settings"))
		if err != nil {
			return err
		}
		return bucket.Put([]byte("s"), []byte(value))
	})
	if err != nil {
		t.Fatal(err)
	}
}

func legacyPaths(t *testing.T) (legacy, path string) {
	dir := t.TempDir()
	return filepath.Join(dir, "settings.db"), filepath.Join(dir, "settings.json")
}

func TestMigrateLegacyMovesTheSettingsToTheJSONFile(t *testing.T) {
	legacy, path := legacyPaths(t)
	writeLegacy(t, legacy, `{"SecondsBetweenCaptures":600,"OffsetWithinHour":0,"PhotoResolutionWidth":1920,"PhotoResolutionHeight":1080,"PreviewResolutionWidth":640,"PreviewResolutionHeight":480,"Quality":90,"SkipNightPhotos":true,"Latitude":53.41,"Longitude":-8.24}`)
	if err := MigrateLegacy(legacy, path); err != nil {
		t.Fatal(err)
	}
	got, err := ReadSettings(path)
	if err != nil || got.SecondsBetweenCaptures != 600 || got.Quality != 90 || got.PhotoResolutionWidth != 1920 || !got.SkipNightPhotos || got.Latitude != 53.41 {
		t.Fatalf("migrated %s, %v", got, err)
	}
	if exists(legacy) || !exists(legacy+".migrated") {
		t.Fatal("the old database should be renamed to .migrated, so it is not migrated again")
	}
}

func TestMigrateLegacyWithoutAnOldDatabaseDoesNothing(t *testing.T) {
	legacy, path := legacyPaths(t)
	if err := MigrateLegacy(legacy, path); err != nil {
		t.Fatal(err)
	}
	if exists(path) {
		t.Fatal("created the settings file")
	}
}

func TestMigrateLegacyKeepsExistingJSONSettings(t *testing.T) {
	legacy, path := legacyPaths(t)
	writeLegacy(t, legacy, `{"Quality":90}`)
	os.WriteFile(path, []byte(`{"Quality":42}`), 0o600)
	if err := MigrateLegacy(legacy, path); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != `{"Quality":42}` || !exists(legacy) {
		t.Fatalf("settings file %s; the old database must be left alone", b)
	}
}

func TestMigrateLegacyReportsADamagedOldDatabase(t *testing.T) {
	legacy, path := legacyPaths(t)
	os.WriteFile(legacy, []byte("garbage"), 0o600)
	if err := MigrateLegacy(legacy, path); err == nil {
		t.Fatal("expected an error")
	}
	if exists(path) || !exists(legacy) {
		t.Fatal("a damaged old database must be left for inspection, and nothing written")
	}
}
