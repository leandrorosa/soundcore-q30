package main

import (
	"fmt"
	"time"

	"github.com/leandrorosa/soundcore-q30/q30"
	"github.com/leandrorosa/soundcore-q30/rfcomm"
)

const replyTimeout = 3 * time.Second

type device struct {
	conn *rfcomm.Conn
	dec  q30.Decoder
}

func open(addr string) (*device, error) {
	c, err := rfcomm.Dial(addr)
	if err != nil {
		return nil, err
	}
	return &device{conn: c}, nil
}

func (d *device) Close() { d.conn.Close() }

// call sends a command and waits for the inbound packet carrying the same command id.
func (d *device) call(cmd q30.Command, body []byte) ([]byte, error) {
	if err := d.conn.Write(q30.Encode(cmd, body)); err != nil {
		return nil, err
	}
	deadline := time.Now().Add(replyTimeout)
	for left := replyTimeout; left > 0; left = time.Until(deadline) {
		chunk, err := d.conn.Read(left)
		if err != nil {
			return nil, err
		}
		for _, p := range d.dec.Feed(chunk) {
			if p.Command == cmd {
				return p.Body, nil
			}
		}
	}
	return nil, fmt.Errorf("no reply to command %x", cmd)
}

func (d *device) state() (q30.State, error) {
	b, err := d.call(q30.CmdState, nil)
	if err != nil {
		return q30.State{}, err
	}
	return q30.ParseState(b)
}
