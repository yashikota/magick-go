package runtimebundle

import (
	"os"
	"path/filepath"
)

func ApplyPolicy(permissive bool) (string, error) {
	dir, err := os.MkdirTemp("", "magickgo-policy-*")
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
</policymap>
`
