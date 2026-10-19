package q30

import (
	"encoding/binary"
	"fmt"
	"time"
)

// Conn is a byte stream to the headphones, such as an RFCOMM channel.
type Conn interface {
	Write(b []byte) error
	// Read returns the next chunk of bytes, or an empty slice once timeout elapses.
	Read(timeout time.Duration) ([]byte, error)
}

type Client struct {
	conn    Conn
	dec     Decoder
	timeout time.Duration
}

func NewClient(conn Conn) *Client {
	return &Client{conn: conn, timeout: 3 * time.Second}
}

func (c *Client) State() (State, error) {
	b, err := c.call(CmdState, nil)
	if err != nil {
		return State{}, err
	}
	return ParseState(b)
}

func (c *Client) SetSoundModes(m SoundModes) error {
	_, err := c.call(CmdSoundModes, m[:])
	return err
}

func (c *Client) SetEqualizer(preset uint16, bands [8]float64) error {
	body := binary.LittleEndian.AppendUint16(nil, preset)
	_, err := c.call(CmdEqualizer, append(body, EncodeBands(bands)...))
	return err
}

// call sends a command and waits for the inbound packet carrying the same
// command id, skipping unrelated packets the headphones push on their own.
func (c *Client) call(cmd Command, body []byte) ([]byte, error) {
	if err := c.conn.Write(Encode(cmd, body)); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(c.timeout)
	for left := c.timeout; left > 0; left = time.Until(deadline) {
		chunk, err := c.conn.Read(left)
		if err != nil {
			return nil, err
		}
		for _, p := range c.dec.Feed(chunk) {
			if p.Command == cmd {
				return p.Body, nil
			}
		}
	}
	return nil, fmt.Errorf("no reply to command %x", cmd)
}
