package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/leandrorosa/soundcore-q30/internal/q30"
	"github.com/leandrorosa/soundcore-q30/internal/rfcomm"
)

func (a *app) statusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show battery, sound mode, eq and firmware",
		Args:  cobra.NoArgs,
		RunE: a.withClient(func(cmd *cobra.Command, c *q30.Client, _ []string) error {
			s, err := c.State()
			if err != nil {
				return err
			}
			battery := fmt.Sprintf("%d%%", s.Battery*20)
			if s.Charging {
				battery += " (charging)"
			}
			eq := s.EQName()
			if s.EQPreset == q30.CustomEQ {
				eq += fmt.Sprintf(" %v", s.EQBands)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintf(w, "battery\t%s\n", battery)
			fmt.Fprintf(w, "mode\t%s\n", describeMode(s.SoundModes))
			fmt.Fprintf(w, "eq\t%s\n", eq)
			fmt.Fprintf(w, "firmware\t%s\n", s.Firmware)
			fmt.Fprintf(w, "serial\t%s\n", s.Serial)
			return w.Flush()
		}),
	}
}

func (a *app) ancCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "anc [transport|indoor|outdoor]",
		Short:     "Turn on noise cancelling, optionally picking its mode",
		Args:      cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
		ValidArgs: q30.ANCModes,
		RunE: a.withClient(func(cmd *cobra.Command, c *q30.Client, args []string) error {
			s, err := c.State()
			if err != nil {
				return err
			}
			mode := s.SoundModes.ANC()
			if len(args) == 1 {
				if mode, err = q30.ParseANCMode(args[0]); err != nil {
					return err
				}
			}
			return setModes(cmd, c, s.SoundModes.With(q30.ANC, mode))
		}),
	}
}

func (a *app) modeCmd(use, short string, mode q30.AmbientMode) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: a.withClient(func(cmd *cobra.Command, c *q30.Client, _ []string) error {
			s, err := c.State()
			if err != nil {
				return err
			}
			return setModes(cmd, c, s.SoundModes.With(mode, s.SoundModes.ANC()))
		}),
	}
}

func (a *app) toggleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "toggle",
		Short: "Switch between anc and normal",
		Args:  cobra.NoArgs,
		RunE: a.withClient(func(cmd *cobra.Command, c *q30.Client, _ []string) error {
			s, err := c.State()
			if err != nil {
				return err
			}
			next := q30.ANC
			if s.SoundModes.Ambient() == q30.ANC {
				next = q30.Normal
			}
			return setModes(cmd, c, s.SoundModes.With(next, s.SoundModes.ANC()))
		}),
	}
}

func setModes(cmd *cobra.Command, c *q30.Client, m q30.SoundModes) error {
	if err := c.SetSoundModes(m); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), describeMode(m))
	return nil
}

func describeMode(m q30.SoundModes) string {
	if m.Ambient() == q30.ANC {
		return fmt.Sprintf("%s %s", m.Ambient(), m.ANC())
	}
	return m.Ambient().String()
}

func (a *app) eqCmd() *cobra.Command {
	apply := a.withClient(func(cmd *cobra.Command, c *q30.Client, args []string) error {
		if args[0] == "custom" {
			if len(args) != 2 {
				return fmt.Errorf("usage: q30 eq custom B1,...,B8")
			}
			bands, err := q30.ParseBands(args[1])
			if err != nil {
				return err
			}
			if err := c.SetEqualizer(q30.CustomEQ, bands); err != nil {
				return err
			}
		} else {
			id, ok := q30.PresetID(args[0])
			if !ok {
				return fmt.Errorf("unknown preset %q, run `q30 eq` to list them", args[0])
			}
			if err := c.SetEqualizer(id, q30.Presets[id].Bands); err != nil {
				return err
			}
		}
		fmt.Fprintln(cmd.OutOrStdout(), "eq", args[0])
		return nil
	})
	return &cobra.Command{
		Use:   "eq [PRESET | custom B1,...,B8]",
		Short: "List eq presets, apply one, or set 8 custom bands in dB (-12..12)",
		Example: "  q30 eq rock\n" +
			"  q30 eq custom 3,2,0,0,0,1,2,3",
		Args: cobra.MaximumNArgs(2),
		ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]cobra.Completion, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			names := []cobra.Completion{"custom"}
			for _, p := range q30.Presets {
				names = append(names, p.Name)
			}
			return names, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				for _, p := range q30.Presets {
					fmt.Fprintln(cmd.OutOrStdout(), p.Name)
				}
				return nil
			}
			return apply(cmd, args)
		},
	}
}

func (a *app) connectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connect",
		Short: "Connect bluetooth",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			addr, err := a.address()
			if err != nil {
				return err
			}
			return rfcomm.Connect(addr)
		},
	}
}

func (a *app) offCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "off",
		Aliases: []string{"disconnect"},
		Short:   "Disconnect bluetooth (the Q30 has no remote power off)",
		Args:    cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			addr, err := a.address()
			if err != nil {
				return err
			}
			return rfcomm.Disconnect(addr)
		},
	}
}
