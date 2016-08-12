#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <unistd.h>
#include <asm/types.h>
#include <sys/socket.h>
#include <linux/netlink.h>
#include <linux/rtnetlink.h>
#include <netinet/in.h>
#include <linux/tcp.h>
#include <linux/sock_diag.h>
#include <linux/inet_diag.h>
#include <arpa/inet.h>
#include <pwd.h>
#include <libmnl/libmnl.h>

// This code is based on https://github.com/kristrev/inet-diag-example/blob/master/inet_monitor.c,
// which is in the public domain.

// Copied from libmnl source
#define SOCKET_BUFFER_SIZE (getpagesize() < 8192L ? getpagesize() : 8192L)

// Callback back into Go code
void SocketInfoCallback(unsigned int uid, unsigned int inode,
  unsigned int src, unsigned int dst,
  unsigned int sport, unsigned int dport);

int send_diag_msg(int sockfd, int family){
    struct msghdr msg;
    struct nlmsghdr nlh;
    struct inet_diag_req_v2 conn_req;
    struct sockaddr_nl sa;
    struct iovec iov[4];
    int retval = 0;

    //For the filter
    struct rtattr rta;
    int filter_len = 0;

    memset(&msg, 0, sizeof(msg));
    memset(&sa, 0, sizeof(sa));
    memset(&nlh, 0, sizeof(nlh));
    memset(&conn_req, 0, sizeof(conn_req));

    //No need to specify groups or pid. This message only has one receiver and
    //pid 0 is kernel
    sa.nl_family = AF_NETLINK;

    //Address family and protocol we are interested in. sock_diag can also be
    //used with UDP sockets, DCCP sockets and Unix sockets, to mention a few.
    //This code requests information about TCP sockets bound to IPv4
    //or IPv6 addresses.
    conn_req.sdiag_family = family;
    conn_req.sdiag_protocol = IPPROTO_TCP;

    conn_req.idiag_states = 0xFFF;

    nlh.nlmsg_len = NLMSG_LENGTH(sizeof(conn_req));
    //In order to request a socket bound to a specific IP/port, remove
    //NLM_F_DUMP and specify the required information in conn_req.id
    nlh.nlmsg_flags = NLM_F_DUMP | NLM_F_REQUEST;

    //Avoid using compat by specifying family + protocol in header
    nlh.nlmsg_type = SOCK_DIAG_BY_FAMILY;
    iov[0].iov_base = (void*) &nlh;
    iov[0].iov_len = sizeof(nlh);
    iov[1].iov_base = (void*) &conn_req;
    iov[1].iov_len = sizeof(conn_req);

    //Set message correctly
    msg.msg_name = (void*) &sa;
    msg.msg_namelen = sizeof(sa);
    msg.msg_iov = iov;
    msg.msg_iovlen = 2;

    return sendmsg(sockfd, &msg, 0);
}

void parse_diag_msg(struct inet_diag_msg *diag_msg, int rtalen){
    struct rtattr *attr;
    struct tcp_info *tcpi;
    char local_addr_buf[INET6_ADDRSTRLEN];
    char remote_addr_buf[INET6_ADDRSTRLEN];
    struct passwd *uid_info = NULL;

    if(diag_msg->idiag_family == AF_INET){
        SocketInfoCallback(diag_msg->idiag_uid, diag_msg->idiag_inode,
          *(diag_msg->id.idiag_src), *(diag_msg->id.idiag_dst),
          ntohs(diag_msg->id.idiag_sport), ntohs(diag_msg->id.idiag_dport));
    } else if(diag_msg->idiag_family == AF_INET6){
        // TODO add IPv6 support here as soon as connbeat supports it
        // also probably needs a change to conn_req.sdiag_family
    } else {
        fprintf(stderr, "Unknown family %d\n", diag_msg->idiag_family);
        return;
    }
}

static int cb(const struct nlmsghdr *nlh, void *data) {
  struct inet_diag_msg *diag_msg = (struct inet_diag_msg*) NLMSG_DATA(nlh);
  int rtalen = nlh->nlmsg_len - NLMSG_LENGTH(sizeof(*diag_msg));

  parse_diag_msg(diag_msg, rtalen);
  return MNL_CB_OK;
}

void recvinfo(const struct mnl_socket * sock) {
  int ret;
  uint8_t recv_buf[SOCKET_BUFFER_SIZE];
  while(1) {
    ret = mnl_socket_recvfrom(sock, recv_buf, sizeof(recv_buf));
    if (ret < 0) {
      perror("receiving");
      return;
    }
    ret = mnl_cb_run(recv_buf, ret, 0, mnl_socket_get_portid(sock), cb, NULL);
    if (ret == -1) {
      perror("running callbacks");
    } else if (ret <= MNL_CB_STOP) {
      return;
    }
  }
}

int poll(int sockfd, const struct mnl_socket * sock) {
  int res = send_diag_msg(sockfd, AF_INET);
  if (res < 0) {
    return res;
  }
  recvinfo(sock);

  res = send_diag_msg(sockfd, AF_INET6);
  if (res < 0) {
    return res;
  }
  recvinfo(sock);

  return 0;
}
