package fs

// Session is one client view (cwd and user) over a shared FS.
type Session struct {
	fs  *FS
	cwd string
	user string
}

// NewSession creates a session at / as user "user".
func NewSession(f *FS) *Session {
	return &Session{fs: f, cwd: "/", user: "user"}
}

// Pwd returns the absolute path of the working directory.
func (s *Session) Pwd() (string, error) {
	return s.cwd, nil
}
