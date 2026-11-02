package composeassets

import "embed"

// Files contains the Compose deployment assets embedded in the YACS binary.
//
//go:embed compose.yml csp.yaml
var Files embed.FS
