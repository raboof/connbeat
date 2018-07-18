# netlink - a simple netlink library for go

[![Build Status](https://travis-ci.org/eleme/netlink.png?branch=master)](https://travis-ci.org/eleme/netlink)
[![GoDoc](https://godoc.org/github.com/eleme/netlink?status.svg)](https://godoc.org/github.com/eleme/netlink)

## what's netlink ?

The netlink package provides a simple netlink library for go. Netlink
is the interface a user-space program in linux uses to communicate with
the kernel.

more information **[here](https://github.com/eleme/sre/blob/master/linux/netlink-i.md)**

## example

```go
func (c *ExecCollector) prepareSocket() (*netlink.NetlinkSocket, error) {
	ns, err := netlink.NewNetlinkSocket(netlink.CONNECTOR, netlink.CN_IDX_PROC)
	if err != nil {
		return nil, fmt.Errorf("create netlink socket: %v", err)
	}
	req := netlink.NewNetlinkRequest()
	{
		msg := netlink.NewCnMsg()
		req.AddData(msg)
		op := netlink.PROC_CN_MCAST_LISTEN
		req.AddData(&op)
	}
	err = ns.Send(req)
	if err != nil {
		return nil, fmt.Errorf("Exec: %v", err)
	}
	go func() {
		<-c.shutdown
		ns.Close()
	}()
	return ns, nil
}
```

## future work

WIP
