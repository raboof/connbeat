// +build !linux

package processes

import (
	"errors"
)

func (ps *Processes) Refresh() error {
	// Only supported under linux
	return nil
}

// Refresh reloads all the data associated with this process.
func (p *UnixProcess) Refresh(exposeCmdline, exposeEnviron bool) error {
	return errors.New("Only supported under linux")
}

func (ps *Processes) FindProcessByInode(inode uint64) *UnixProcess {
	// Only supported under linux
	return nil
}
