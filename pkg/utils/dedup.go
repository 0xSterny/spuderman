package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"
)

// Deduplicator tracks seen hashes to avoid duplicates
type Deduplicator struct {
	mu     sync.Mutex
	hashes map[string]bool
}

func NewDeduplicator() *Deduplicator {
	return &Deduplicator{
		hashes: make(map[string]bool),
	}
}

// IsDuplicate checks if hash exists. If not, adds it and returns false.
// If it exists, returns true.
func (d *Deduplicator) IsDuplicate(hash string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.hashes[hash] {
		return true
	}
	d.hashes[hash] = true
	return false
}

// CheckAndAdd calculates hash of reader. If duplicate, returns true.
// If unique, adds to set and returns false.
// IMPORTANT: This consumes the reader. You cannot read from it again unless you seek or it was a buffer.
func (d *Deduplicator) CheckAndAdd(r io.Reader) (bool, string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, r); err != nil {
		return false, "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))

	return d.IsDuplicate(hash), hash, nil
}
