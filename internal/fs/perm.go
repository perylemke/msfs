package fs

func (s *Session) checkPerm(n *inode, want permMode) error {
	if s.fs.userCan(s.user, n, want) {
		return nil
	}
	return ErrPermission
}

func (f *FS) userCan(user string, n *inode, want permMode) bool {
	if user == n.owner {
		return n.mode&want == want
	}
	u, ok := f.users[user]
	if !ok {
		return false
	}
	for _, g := range u.Groups {
		if g == n.group {
			return n.mode&want == want
		}
	}
	return n.mode&want == want
}
