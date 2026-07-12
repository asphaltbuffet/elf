// Package secret encrypts and decrypts an exercise's Solution Set — its
// info.json and language subdirectories — to and from per-file age ciphertext,
// using SSH keys. The plaintext is never removed: the .age files are derived
// artifacts (see docs/adr/0019 and docs/adr/0020).
package secret
