package reloader
// chain to check files whether should be ignored.
// based on .gitignore file syntax

type FilterChain struct {
	rules []string
}
