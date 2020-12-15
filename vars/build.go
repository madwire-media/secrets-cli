package vars

var (
	// BuildVersion is the version this application was built with. By default
	// it is set to "dev", but it gets overridden by goreleaser
	BuildVersion = "dev"

	// BuildCommit is the commit this application was built on. By default it is
	// set to "unknown", but it gets overridden by goreleaser
	BuildCommit = "unknown"
)
