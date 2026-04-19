package facts

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SSHCollector gathers SSH host key fingerprints.
type SSHCollector struct{}

func (s *SSHCollector) Collect() (map[string]any, error) {
	result := make(map[string]any)
	dir := "/etc/ssh"

	matches, err := filepath.Glob(filepath.Join(dir, "ssh_host_*_key.pub"))
	if err != nil {
		return result, nil
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		parts := strings.Fields(string(data))
		if len(parts) < 2 {
			continue
		}

		keyType := parts[0]
		keyData, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			continue
		}

		// Derive short name: ssh-rsa -> rsa, ssh-ed25519 -> ed25519, ecdsa-sha2-nistp256 -> ecdsa
		name := keyTypeName(keyType)

		sha256Hash := sha256.Sum256(keyData)
		md5Hash := md5.Sum(keyData)

		result[fmt.Sprintf("ssh.%s.fingerprint.sha256", name)] = "SHA256:" + base64.RawStdEncoding.EncodeToString(sha256Hash[:])
		result[fmt.Sprintf("ssh.%s.fingerprint.md5", name)] = colonHex(md5Hash[:])
		result[fmt.Sprintf("ssh.%s.type", name)] = keyType
	}

	return result, nil
}

func keyTypeName(keyType string) string {
	switch {
	case strings.Contains(keyType, "ed25519"):
		return "ed25519"
	case strings.Contains(keyType, "ecdsa"):
		return "ecdsa"
	case strings.Contains(keyType, "rsa"):
		return "rsa"
	case strings.Contains(keyType, "dsa"):
		return "dsa"
	default:
		return strings.TrimPrefix(keyType, "ssh-")
	}
}

func colonHex(b []byte) string {
	parts := make([]string, len(b))
	for i, v := range b {
		parts[i] = fmt.Sprintf("%02x", v)
	}
	return strings.Join(parts, ":")
}
