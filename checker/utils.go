package checker

import (
	"errors"
	"net"

	"golang.org/x/sys/unix"
)

func _Socket(family int) (int, error) {
	fd, err := unix.Socket(family, unix.SOCK_STREAM, 0)
	if err != nil {
		return -1, err
	}
	unix.CloseOnExec(fd)
	return fd, nil
}

func _SetSockOpts(fd int) error {
	if err := unix.SetNonblock(fd, true); err != nil {
		return err
	}
	if err := unix.SetsockoptInt(fd, unix.IPPROTO_TCP, unix.TCP_QUICKACK, 0); err != nil {
		return err
	}
	zeroLinger := unix.Linger{Onoff: 1, Linger: 0}
	return unix.SetsockoptLinger(fd, unix.SOL_SOCKET, unix.SO_LINGER, &zeroLinger)
}

func _ResolveAddress(addr string) (unix.Sockaddr, error) {
	tAddr, err := net.ResolveTCPAddr("tcp", addr) // FIXME: 没有办法设置超时时间？
	if err != nil {
		return nil, err
	}

	if ip := tAddr.IP.To4(); ip == nil {
		return nil, errors.New("not implement")
	} else {
		sockAddr := &unix.SockaddrInet4{Port: tAddr.Port, Addr: [net.IPv4len]byte{}}
		copy(sockAddr.Addr[:], ip)
		return sockAddr, nil
	}
}

func doTcpCheck(fd int, addr unix.Sockaddr) (bool, error) {
	switch err := unix.Connect(fd, addr); err {
	case unix.EALREADY, unix.EINPROGRESS, unix.EINTR:
		return false, nil
	case nil, unix.EISCONN:
		return true, nil
	case unix.EINVAL:
		return false, err
	default:
		return false, err
	}
}
