package types

import "golang.org/x/sys/unix"

type Statfs struct {
	Bsize  int64
	Blocks uint64
	Bavail uint64
}

func (s *Statfs) Syscall(path string) error {
	var buf unix.Statfs_t
	if err := unix.Statfs(path, &buf); err != nil {
		return err
	}
	s.Bsize = int64(buf.Bsize)
	s.Blocks = buf.Blocks
	s.Bavail = buf.Bavail
	return nil
}

func (s *Statfs) Available() int64 {
	return int64(s.Bavail) * s.Bsize
}
