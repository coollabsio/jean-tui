package install

import (
	"crypto/sha256"
	"fmt"
)

// CalculateWrapperChecksum generates a SHA256 checksum for the wrapper template
// This is used to detect if the installed wrapper needs to be updated
func CalculateWrapperChecksum(shell Shell) string {
	var template string

	switch shell {
	case Bash, Zsh:
		template = BashZshWrapper
	case Fish:
		template = FishWrapper
	default:
		return ""
	}

	// Calculate SHA256 hash of the template
	hash := sha256.Sum256([]byte(template))
	return fmt.Sprintf("%x", hash)
}
