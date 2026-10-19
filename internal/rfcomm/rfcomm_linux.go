// BlueZ backend: a raw RFCOMM socket for data, bluetoothctl for connection management.

package rfcomm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os/exec"
	"slices"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// The Q30 serves its control protocol on SPP, which it always exposes on channel 12.
const sppChannel = 12

type Conn struct{ fd int }

func Dial(addr string) (*Conn, error) {
	mac, err := net.ParseMAC(normalize(addr))
	if err != nil {
		return nil, err
	}
	fd, err := unix.Socket(unix.AF_BLUETOOTH, unix.SOCK_STREAM, unix.BTPROTO_RFCOMM)
	if err != nil {
		return nil, fmt.Errorf("rfcomm socket: %w", err)
	}
	// bdaddr_t is little-endian.
	sa := &unix.SockaddrRFCOMM{Channel: sppChannel}
	copy(sa.Addr[:], mac)
	slices.Reverse(sa.Addr[:])
	if err := unix.Connect(fd, sa); err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("open rfcomm: %w", err)
	}
	return &Conn{fd}, nil
}

func (c *Conn) Write(b []byte) error {
	for len(b) > 0 {
		n, err := unix.Write(c.fd, b)
		if err != nil {
			return fmt.Errorf("write: %w", err)
		}
		b = b[n:]
	}
	return nil
}

// Read returns the next chunk of received bytes, or nil once timeout elapses.
func (c *Conn) Read(timeout time.Duration) ([]byte, error) {
	fds := []unix.PollFd{{Fd: int32(c.fd), Events: unix.POLLIN}}
	n, err := unix.Poll(fds, int(timeout.Milliseconds()))
	if err == unix.EINTR || n == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 1024)
	n, err = unix.Read(c.fd, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}
	return buf[:n], nil
}

func (c *Conn) Close() { unix.Close(c.fd) }

func bluetoothctl(args ...string) ([]byte, error) {
	out, err := exec.Command("bluetoothctl", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bluetoothctl %s: %w: %s", strings.Join(args, " "), err, bytes.TrimSpace(out))
	}
	return out, nil
}

func Connect(addr string) error {
	_, err := bluetoothctl("connect", normalize(addr))
	return err
}

func Disconnect(addr string) error {
	_, err := bluetoothctl("disconnect", normalize(addr))
	return err
}

func Connected(addr string) bool {
	out, err := bluetoothctl("info", normalize(addr))
	return err == nil && bytes.Contains(out, []byte("Connected: yes"))
}

// Find returns the address of the first known device whose name contains name.
func Find(name string) (string, error) {
	out, err := bluetoothctl("devices")
	if err != nil {
		return "", err
	}
	s := bufio.NewScanner(bytes.NewReader(out))
	for s.Scan() {
		// "Device 00:11:22:33:44:55 Soundcore Life Q30"
		f := strings.SplitN(s.Text(), " ", 3)
		if len(f) == 3 && f[0] == "Device" && strings.Contains(strings.ToLower(f[2]), strings.ToLower(name)) {
			return f[1], nil
		}
	}
	return "", fmt.Errorf("%w matching %q", ErrNotFound, name)
}

// normalize accepts the macOS "00-11-22-..." form too, so Q30_ADDR works on both.
func normalize(addr string) string {
	return strings.ToUpper(strings.ReplaceAll(addr, "-", ":"))
}
