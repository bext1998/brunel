package filetools

import (
	"crypto/sha256"
	"encoding/hex"
)

// hashBytes returns the lowercase hex-encoded SHA-256 digest of data. Every
// hash this package returns or checks covers the complete, unmodified file
// content, never just a requested line range.
func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
