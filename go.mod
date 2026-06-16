module git.henrydowd.dev/henry/portfolio

// Deliberately dependency-free: the entire service is the Go standard library.
// That keeps the build reproducible without a module proxy, produces a tiny
// static binary, and means there is no go.sum to manage. See README.md.
go 1.23
