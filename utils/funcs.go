package utils

import (
  "net"
  "strings"
)

func RemoteIpConn(conn net.Conn) string {

  // ip, port, err := net.SplitHostPort(conn.RemoteAddr)
  remoteAddr := conn.RemoteAddr();
  ip := strings.Split(remoteAddr.String(), ":")[0]

  return ip
}
