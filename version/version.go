package version

// Version is stamped at build time via:
//   -ldflags "-X github.com/truedem0n/playbridge-stream-resolver/version.Version=x.y.z"
// Defaults to "dev" for local builds.
var Version = "dev"
