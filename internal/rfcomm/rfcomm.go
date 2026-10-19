// Package rfcomm opens a serial (RFCOMM) channel to a paired Bluetooth device.
package rfcomm

import "errors"

var ErrNotFound = errors.New("no paired device found")
