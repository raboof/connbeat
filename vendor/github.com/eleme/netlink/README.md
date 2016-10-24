# netlink - a simple Golang netlink library 

[![Build Status](https://travis-ci.org/eleme/netlink.png?branch=master)](https://travis-ci.org/eleme/netlink)
[![GoDoc](https://godoc.org/github.com/eleme/netlink?status.svg)](https://godoc.org/github.com/eleme/netlink)

## what's netlink ?

The netlink package provides a communication protocol which allows process exchanging 
information no matter ther are in user or kernel space. 

## The advantage of netlink
Most kernal's process need communicate with user's process in `Linux`, but traditional 
`Unix`'s IPC (`pipe`, `message queue`, `shared memory` and `singal`) can not offer a 
strong support for the communication between user's process and kernel. `Linux` provides 
a lot other methods which allow user's process can communicate with kernel, but they are 
very hard to use. To make these method easier to user for user, especially for 
Operational Engineer is the reason why we develop `netlink`. 

Compared with native method, `netlink` has following advantages:
- The prerequistie of using `netlink` is just adding a new type in `include/linux/netlink.h`
- `nektlink` is asynchronous. An instance will not be blocked after sending messages. 
The messages is actually saved in socket's messages buffer. 
- `netlink` is modularizeed. `netlink`'application and kernel parts are independent. They do not have any compile 
dependency. But system call does. Whenever a new system all want to be used. It has to be compiled statically into kernel. 
- `netlink` supports multicast. Kernel modules or process can pass messages into multiple `netlink` group. 
- `netlink` can be called from kernel, unlike `ioctl`.
- `netlink` uses standard socket API, hence it is pretty simple to use unlike `system call` or `iotcl`. 
But these traditional protocol can not give enough support of the communication
between process(user space) and kernel. `Linux` provides a lot of methods to 
fix such problem, but they are hard to learn. A new protocol `netlink` is
easy to learn. 

## netlink feature
`netlink` only provides basic communication protocol between user's process and kernel's process. These specific tasks are
based on the subprotocol in `netlink`. Additionally, inside `Linux`'s kernel, it already has a protocol which is:

```c
#define NETLINK_ROUTE       0   /* Routing/device hook              */
#define NETLINK_UNUSED      1   /* Unused number                */
#define NETLINK_USERSOCK    2   /* Reserved for user mode socket protocols  */
#define NETLINK_FIREWALL    3   /* Unused number, formerly ip_queue     */
#define NETLINK_SOCK_DIAG   4   /* socket monitoring                */
#define NETLINK_NFLOG       5   /* netfilter/iptables ULOG */
#define NETLINK_XFRM        6   /* ipsec */
#define NETLINK_SELINUX     7   /* SELinux event notifications */
#define NETLINK_ISCSI       8   /* Open-iSCSI */
#define NETLINK_AUDIT       9   /* auditing */
#define NETLINK_FIB_LOOKUP  10  
#define NETLINK_CONNECTOR   11
#define NETLINK_NETFILTER   12  /* netfilter subsystem */
#define NETLINK_IP6_FW      13
#define NETLINK_DNRTMSG     14  /* DECnet routing messages */
#define NETLINK_KOBJECT_UEVENT  15  /* Kernel messages to userspace */
#define NETLINK_GENERIC     16
/* leave room for NETLINK_DM (DM Events) */
#define NETLINK_SCSITRANSPORT   18  /* SCSI Transports */
#define NETLINK_ECRYPTFS    19
#define NETLINK_RDMA        20
#define NETLINK_CRYPTO      21  /* Crypto layer */
#define NETLINK_INET_DIAG   NETLINK_SOCK_DIAG
```


## How to use netlink?
### user space
If you plan use `netlink` in user space, then you should consider standard socker API. For creating a new `netlink` sockert, user need use
the following arguments to call `socker()`:

```go
socket(AF_NETLINK, SOCK_RAW, netlink_type)
```
You can find more details [here](https://www.infradead.org/~tgr/libnl/)

### kernel space
Any kernel's module wants to use `netlink`(in Linux, not this project) need to include header file `linux/netlink.h`. Compared with the usage of `netlink`, kernel has to 
use particular API defined in `netlink`. For now, `netlink` alraedy implemented a generic protocol `NETLINK_GENERIC` which reduces the extra work.

### example

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
- [ ] need add all test for this project


## miscellaneous
more information **[here](https://github.com/eleme/sre/blob/master/linux/netlink-i.md)**
