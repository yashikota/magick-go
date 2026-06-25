package runtimebundle

import (
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	lastCleanup   time.Time
	lastCleanupMu sync.Mutex
)

func ApplyPolicy(permissive bool) (string, error) {
	cleanupOldPolicies()
	dir, err := os.MkdirTemp("", "mahou-policy-*")
	if err != nil {
		return "", err
	}
	policy := safePolicyXML
	if permissive {
		policy = permissivePolicyXML
	}
	if err := os.WriteFile(filepath.Join(dir, "policy.xml"), []byte(policy), 0o644); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

func cleanupOldPolicies() {
	lastCleanupMu.Lock()
	if time.Since(lastCleanup) < 10*time.Minute {
		lastCleanupMu.Unlock()
		return
	}
	lastCleanup = time.Now()
	lastCleanupMu.Unlock()

	tempDir := os.TempDir()
	pattern := filepath.Join(tempDir, "mahou-policy-*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	now := time.Now()
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > 1*time.Hour {
			_ = os.RemoveAll(match)
		}
	}
}

const safePolicyXML = `<?xml version="1.0" encoding="UTF-8"?>
<policymap>
  <policy domain="coder" rights="none" pattern="PDF" />
  <policy domain="coder" rights="none" pattern="PS" />
  <policy domain="coder" rights="none" pattern="PS2" />
  <policy domain="coder" rights="none" pattern="PS3" />
  <policy domain="coder" rights="none" pattern="EPS" />
  <policy domain="coder" rights="none" pattern="EPS2" />
  <policy domain="coder" rights="none" pattern="EPS3" />
  <policy domain="coder" rights="none" pattern="MVG" />
  <policy domain="coder" rights="none" pattern="MSL" />
  <policy domain="delegate" rights="none" pattern="URL" />
  <policy domain="delegate" rights="none" pattern="HTTP" />
  <policy domain="delegate" rights="none" pattern="HTTPS" />
</policymap>
`

const permissivePolicyXML = `<?xml version="1.0" encoding="UTF-8"?>
<policymap>
  <policy domain="coder" rights="read|write" pattern="*" />
  <policy domain="delegate" rights="execute" pattern="*" />
</policymap>
`
