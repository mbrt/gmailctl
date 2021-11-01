//go:generate go run github.com/go-bindata/go-bindata/go-bindata -nometadata -pkg cmd ../../../gmailctl.libsonnet ../../../default-config.jsonnet

package cmd

// GmailctlLib returns the embedded gmailctl.libsonnet file
func GmailctlLib() string {
	return string(MustAsset("../../../gmailctl.libsonnet"))
}

// DefaultConfig returns the embedded default configuration file
func DefaultConfig() string {
	return string(MustAsset("../../../default-config.jsonnet"))
}
