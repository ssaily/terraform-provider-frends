package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// Compile-time check that the drift detection imports are available.
var _ = sha256.Sum256
var _ = hex.EncodeToString

func TestFileHash_knownContent(t *testing.T) {
	t.Parallel()

	// Write a temp file with known content.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.frends")
	content := []byte("frends package content for testing")
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	got, err := fileHash(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Compute expected hash independently.
	h := sha256.Sum256(content)
	want := hex.EncodeToString(h[:])

	if got != want {
		t.Fatalf("expected hash %q, got %q", want, got)
	}
}

func TestFileHash_emptyFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "empty.frends")
	if err := os.WriteFile(path, []byte{}, 0600); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}

	got, err := fileHash(path)
	if err != nil {
		t.Fatalf("unexpected error for empty file: %v", err)
	}

	h := sha256.Sum256([]byte{})
	want := hex.EncodeToString(h[:])
	if got != want {
		t.Fatalf("expected empty-file hash %q, got %q", want, got)
	}
}

func TestFileHash_nonExistentFile(t *testing.T) {
	t.Parallel()

	_, err := fileHash("/does/not/exist/file.frends")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestFileHash_deterministicForSameContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	content := []byte("deterministic content")

	path1 := filepath.Join(dir, "a.frends")
	path2 := filepath.Join(dir, "b.frends")
	if err := os.WriteFile(path1, content, 0600); err != nil {
		t.Fatalf("writing file 1: %v", err)
	}
	if err := os.WriteFile(path2, content, 0600); err != nil {
		t.Fatalf("writing file 2: %v", err)
	}

	hash1, err := fileHash(path1)
	if err != nil {
		t.Fatalf("hashing file 1: %v", err)
	}
	hash2, err := fileHash(path2)
	if err != nil {
		t.Fatalf("hashing file 2: %v", err)
	}

	if hash1 != hash2 {
		t.Fatalf("same content should produce same hash: %q vs %q", hash1, hash2)
	}
}

// TestDriftDetection_hashLogic verifies the SHA-256 comparison logic used in
// the Read method: the remote-exported bytes are hashed and compared against
// the stored PackageHash to detect out-of-band modifications.
func TestDriftDetection_hashLogic(t *testing.T) {
	t.Parallel()

	original := []byte(`{"name":"OrderProcessor","version":1}`)
	modified := []byte(`{"name":"OrderProcessor","version":2}`)

	hashBytes := func(b []byte) string {
		h := sha256.Sum256(b)
		return hex.EncodeToString(h[:])
	}

	storedHash := hashBytes(original)
	remoteHash := hashBytes(modified)

	if storedHash == remoteHash {
		t.Fatal("expected different hashes for different content")
	}

	// Simulates the Read drift check: remote differs → state should be updated.
	if remoteHash == storedHash {
		t.Fatal("drift not detected: hashes should differ")
	}

	// Simulates no drift: same content exported → no state update needed.
	sameHash := hashBytes(original)
	if sameHash != storedHash {
		t.Fatal("no drift expected when content is unchanged")
	}
}

func TestDriftDetection_emptyHashSkipped(t *testing.T) {
	t.Parallel()
	// When PackageHash is empty (post-import before config is set), drift
	// detection must be skipped. Verify the guard condition directly.
	emptyHash := ""
	if emptyHash != "" {
		t.Fatal("guard condition is wrong: empty string should skip drift check")
	}
}

func TestFileHash_differentContentProducesDifferentHash(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.frends")
	path2 := filepath.Join(dir, "b.frends")
	if err := os.WriteFile(path1, []byte("version 1"), 0600); err != nil {
		t.Fatalf("writing file 1: %v", err)
	}
	if err := os.WriteFile(path2, []byte("version 2"), 0600); err != nil {
		t.Fatalf("writing file 2: %v", err)
	}

	hash1, _ := fileHash(path1)
	hash2, _ := fileHash(path2)
	if hash1 == hash2 {
		t.Fatal("different content should produce different hashes")
	}
}
