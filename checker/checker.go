package checker

import (
	"context"
	"log"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type Checker struct {
	TimeoutSecond int // 秒

	epfd   int // fd for epoll
	m      sync.Map
	cancel context.CancelFunc
}

func NewChecker() (*Checker, error) {
	c := &Checker{TimeoutSecond: 5}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	var err error
	c.epfd, err = unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	go c.checkLoop(ctx)
	return c, nil
}

func (c *Checker) Close() {
	c.cancel()
	unix.Close(c.epfd)
}

const maxEpollEvents = 102400

func (c *Checker) checkLoop(ctx context.Context) {
	var epollEvents [maxEpollEvents]unix.EpollEvent

	for {
		select {
		case <-ctx.Done():
			return
		default:
			nEvents, err := unix.EpollWait(c.epfd, epollEvents[:], 1000)
			if err != nil {
				log.Panicf("epoll_wait: %v", err)
			}
			for i := 0; i < nEvents; i++ {
				fd := int(epollEvents[i].Fd)
				errCode, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
				if errCode != 0 {
					err = unix.Errno(errCode)
				}
				if _ch, ok := c.m.Load(fd); ok {
					_ch.(chan error) <- err
				}
			}
		}
	}
}

func (c *Checker) wait(fd int) error {
	event := unix.EpollEvent{
		Events: unix.EPOLLOUT | unix.EPOLLIN | unix.EPOLLET,
		Fd:     int32(fd),
	}

	ch := make(chan error)
	c.m.Store(fd, ch)
	defer c.m.Delete(fd)

	if err := unix.EpollCtl(c.epfd, unix.EPOLL_CTL_ADD, fd, &event); err != nil {
		return err
	}

	select {
	case err := <-ch:
		return err
	case <-time.After(time.Second * time.Duration(c.TimeoutSecond)):
		return &timeoutError{}
	}
}

func (c *Checker) Check(addr string) error {
	fd, err := _Socket(unix.AF_INET)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	if err := _SetSockOpts(fd); err != nil {
		return err
	}
	tcpAddr, err := _ResolveAddress(addr)
	if err != nil {
		return err
	}
	if ok, err := doTcpCheck(fd, tcpAddr); err != nil {
		return err
	} else if ok {
		return nil
	}
	return c.wait(fd)
}
