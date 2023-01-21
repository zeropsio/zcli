package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/zeropsio/zcli/src/i18n"
	"github.com/zeropsio/zcli/src/proto"
	"github.com/zeropsio/zcli/src/proto/daemon"
	"github.com/zeropsio/zcli/src/utils/generic"
)

func sshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ssh serviceName",
		Short:        i18n.CmdSsh,
		Args:         ExactNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			regSignals(cancel)

			zeropsDaemonClient, closeFn, err := daemon.CreateClient(ctx)
			if err != nil {
				return err
			}
			defer closeFn()

			_, err = zeropsDaemonClient.StatusVpn(ctx, &daemon.StatusVpnRequest{})
			daemonInstalled, err := proto.DaemonError(err)
			if err != nil {
				return err
			}
			if !daemonInstalled {
				return errors.New(i18n.VpnDaemonUnavailable)
			}

			host := fmt.Sprintf("all.runtime.%s.zerops", args[0])
			m := &sshModel{
				ctx:  ctx,
				host: host,
			}
			err = m.CreateChoices()
			if err != nil {
				return err
			}

			_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
			if err != nil {
				return err
			}
			return m.Err()
		},
	}
	return cmd
}

type choice struct {
	hostname string
	ip       string
}

func (c choice) render() string {
	return fmt.Sprintf("host: %s, ip: %s", c.hostname, c.ip)
}

type sshModel struct {
	ctx     context.Context
	host    string
	err     error
	choices []choice
	cursor  int
}

func (s *sshModel) CreateChoices() error {
	ips, err := net.DefaultResolver.LookupIP(s.ctx, "ip6", s.host)
	if err != nil {
		return err
	}

	choices := generic.TransformSlice(ips, func(i net.IP) choice {
		addr, err := net.DefaultResolver.LookupAddr(s.ctx, i.String())
		if err != nil {
			return choice{ip: i.String()}
		}
		c := addr[0]
		c = c[:len(c)-1]
		return choice{hostname: c, ip: i.String()}
	})
	s.choices = choices
	return nil
}

func (s *sshModel) Err() error {
	return s.err
}

func (s *sshModel) Init() tea.Cmd {
	return tea.ClearScreen
}

func (s *sshModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "ctrl+c", "q":
			return s, tea.Sequence(
				tea.ExitAltScreen,
				tea.ClearScreen,
				tea.Quit,
			)
		case "up":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down":
			if s.cursor < len(s.choices)-1 {
				s.cursor++
			}
		case "enter", " ":
			c := exec.Command("ssh", "-p", "65437", s.choices[s.cursor].hostname)
			c.Stdout = os.Stdout
			return s, tea.Sequence(
				tea.ExitAltScreen,
				tea.ClearScreen,
				tea.ExecProcess(c, nil),
				tea.EnterAltScreen,
				func() tea.Msg {
					s.err = s.CreateChoices()
					if s.err != nil {
						return tea.Quit()
					}
					return nil
				},
			)
		}
	}
	return s, nil
}

func (s *sshModel) View() string {
	// The header
	d := "Select container to join:\n\n"

	// Iterate over our choices
	for i, choice := range s.choices {
		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if s.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		d += fmt.Sprintf("%s %s\n", cursor, choice.render())
	}

	// The footer
	d += "\nPress q to quit.\n"

	// Send the UI for rendering
	return d
}
