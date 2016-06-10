// +build darwin

package processes

type Processes struct {
}

func New(exposeCmdline, exposeEnviron bool) *Processes {
	return &Processes{}
}

func (ps *Processes) FindProcessByInode(inode int64) *UnixProcess {
	return nil
}
