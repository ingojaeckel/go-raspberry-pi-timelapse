package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

// Versions before the JSON settings file kept the settings in a bbolt
// database, as JSON under key "s" in bucket "settings". This file is the only
// user of bbolt: delete it, and the dependency, once every camera has
// started a version that migrates.

// MigrateLegacy moves the settings from the old database at legacyPath to
// the settings file at path, once: only while path does not exist. The old
// database is then renamed to legacyPath.migrated. A damaged old database is
// left in place and reported; the app then starts with the defaults.
func MigrateLegacy(legacyPath, path string) error {
	if exists(path) || !exists(legacyPath) {
		return nil
	}
	v, err := readLegacy(legacyPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", legacyPath, err)
	}
	if err := writeSettings(path, v); err != nil {
		return err
	}
	slog.Info("settings moved to a new file", "from", legacyPath, "to", path)
	return os.Rename(legacyPath, legacyPath+".migrated")
}

func readLegacy(path string) (v Settings, err error) {
	defer func() {
		if r := recover(); r != nil { // bbolt panics on some damaged pages
			err = fmt.Errorf("damaged database: %v", r)
		}
	}()
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: time.Second, ReadOnly: true})
	if err != nil {
		return v, err
	}
	defer db.Close()
	err = db.View(func(tx *bolt.Tx) error {
		var raw []byte
		if bucket := tx.Bucket([]byte("settings")); bucket != nil {
			raw = bucket.Get([]byte("s"))
		}
		if raw == nil {
			return errors.New("no settings saved")
		}
		return json.Unmarshal(raw, &v)
	})
	return v, err
}

// exists reports whether path may exist: anything but "not found" counts.
func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
