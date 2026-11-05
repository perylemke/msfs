package fs

// User represents a filesystem user.
type User struct {
	Name   string
	Groups []string
}

// Group represents a named group of users.
type Group struct {
	Name    string
	Members []string
}
