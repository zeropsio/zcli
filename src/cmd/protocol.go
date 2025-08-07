package cmd

import (
	"context"
	"net/url"
	"strings"

	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/protocol"
)

func protocolCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("protocol").
		Short("Manage custom protocol registration for zcli://").
		AddChildrenCmd(protocolRegisterCmd()).
		AddChildrenCmd(protocolUnregisterCmd()).
		AddChildrenCmd(protocolStatusCmd()).
		AddChildrenCmd(protocolHandleCmd()).
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			cmdData.PrintHelp()
			return nil
		})
}

func protocolRegisterCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("register").
		Short("Register the zcli:// custom protocol handler").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			registration := protocol.NewRegistration()

			isRegistered, err := registration.IsRegistered(ctx)
			if err != nil {
				return err
			}

			if isRegistered {
				cmdData.Stdout.PrintLines("Protocol handler is already registered.")
				return nil
			}

			if err := registration.Register(ctx); err != nil {
				return err
			}

			cmdData.Stdout.PrintLines("Protocol handler registered successfully!")
			cmdData.Stdout.PrintLines("You can now use zcli:// URLs to open the CLI.")
			return nil
		})
}

func protocolUnregisterCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("unregister").
		Short("Unregister the zcli:// custom protocol handler").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			registration := protocol.NewRegistration()

			isRegistered, err := registration.IsRegistered(ctx)
			if err != nil {
				return err
			}

			if !isRegistered {
				cmdData.Stdout.PrintLines("Protocol handler is not registered.")
				return nil
			}

			if err := registration.Unregister(ctx); err != nil {
				return err
			}

			cmdData.Stdout.PrintLines("Protocol handler unregistered successfully!")
			return nil
		})
}

func protocolStatusCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("status").
		Short("Check the status of the zcli:// protocol registration").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			registration := protocol.NewRegistration()

			isRegistered, err := registration.IsRegistered(ctx)
			if err != nil {
				return err
			}

			if isRegistered {
				cmdData.Stdout.PrintLines("Protocol handler is registered.")
			} else {
				cmdData.Stdout.PrintLines("Protocol handler is not registered.")
			}

			return nil
		})
}

func protocolHandleCmd() *cmdBuilder.Cmd {
	return cmdBuilder.NewCmd().
		Use("handle <url>").
		Short("Handle a zcli:// protocol URL (internal use)").
		Arg("url").
		GuestRunFunc(func(ctx context.Context, cmdData *cmdBuilder.GuestCmdData) error {
			args := cmdData.Args["url"]
			if len(args) == 0 {
				cmdData.PrintHelp()
				return nil
			}

			rawURL := args[0]

			// Clean up the URL if it's wrapped in quotes or has extra characters
			rawURL = strings.Trim(rawURL, "\"'")

			protocolURL, err := url.Parse(rawURL)
			if err != nil {
				cmdData.Stderr.Printf("Invalid URL format: %v\n", err)
				return nil
			}

			if protocolURL.Scheme != "zcli" {
				cmdData.Stderr.Printf("Unsupported protocol scheme: %s\n", protocolURL.Scheme)
				return nil
			}

			handler := protocol.NewDefaultHandler()
			if err := handler.Handle(ctx, protocolURL); err != nil {
				cmdData.Stderr.Printf("Error handling protocol URL: %v\n", err)
				return nil
			}

			return nil
		})
}
