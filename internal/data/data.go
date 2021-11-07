package data

import _ "embed" // required to enable go:embed directives.

var (
	//go:embed gmailctl.libsonnet
	gmailctlLib string
	//go:embed default-config.jsonnet
	defaultConfig string
)

// GmailctlLib returns the embedded gmailctl.libsonnet file
func GmailctlLib() string {
	return gmailctlLib
}

// DefaultConfig returns the embedded default configuration file
func DefaultConfig() string {
	return defaultConfig
}
