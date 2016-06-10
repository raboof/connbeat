package processes

// UnixProcess is a Process that contains Unix-specific
// fields and information (but is shared across darwin and linux).
type UnixProcess struct {
	pid    int
	ppid   int
	state  rune
	pgrp   int
	sid    int
	inodes []int64

	Binary  string
	Cmdline string
	Environ string
}
