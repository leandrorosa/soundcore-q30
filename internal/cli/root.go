// Package cli implements the q30 command line.
package cli

import (
	"errors"
	"os"

	"github.com/spf13/cobra"

	"github.com/leandrorosa/soundcore-q30/internal/q30"
	"github.com/leandrorosa/soundcore-q30/internal/rfcomm"
)

const deviceName = "Q30"

var errNotConnected = errors.New("headphones not connected, run: q30 connect")

type app struct{ addr string }

func NewRoot(version string) *cobra.Command {
	a := &app{}
	root := &cobra.Command{
		Use:               "q30",
		Short:             "Control Soundcore Life Q30 headphones",
		Version:           version,
		SilenceUsage:      true,
		ValidArgsFunction: cobra.NoFileCompletions,
	}
	root.PersistentFlags().StringVarP(&a.addr, "addr", "a", os.Getenv("Q30_ADDR"),
		`bluetooth address (default: first paired device named "`+deviceName+`", or $Q30_ADDR)`)
	root.AddCommand(
		a.statusCmd(),
		a.ancCmd(),
		a.modeCmd("transparency", "Let ambient sound in", q30.Transparency),
		a.modeCmd("normal", "Turn off anc and transparency", q30.Normal),
		a.toggleCmd(),
		a.eqCmd(),
		a.connectCmd(),
		a.offCmd(),
	)
	return root
}

func (a *app) address() (string, error) {
	if a.addr != "" {
		return a.addr, nil
	}
	return rfcomm.Find(deviceName)
}

// withClient opens an RFCOMM channel to the headphones for the duration of run.
func (a *app) withClient(run func(*cobra.Command, *q30.Client, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		addr, err := a.address()
		if err != nil {
			return err
		}
		if !rfcomm.Connected(addr) {
			return errNotConnected
		}
		conn, err := rfcomm.Dial(addr)
		if err != nil {
			return err
		}
		defer conn.Close()
		return run(cmd, q30.NewClient(conn), args)
	}
}
