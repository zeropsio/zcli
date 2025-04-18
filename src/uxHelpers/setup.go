package uxHelpers

import (
	"context"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
)

func PrintSetupSelector(
	ctx context.Context,
	setups []string,
) (string, error) {
	header, rows := createSetupTableRows(setups)

	selected, err := uxBlock.RunR(
		selector.NewRoot(
			ctx,
			rows,
			selector.WithLabel("Select setup"),
			selector.WithHeader(header),
			selector.WithSetEnableFiltering(true),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return "", err
	}

	if selected > len(setups)-1 {
		return "", errors.New("invalid option")
	}

	return setups[selected], nil
}

func PrintSetupList(
	out io.Writer,
	setups []string,
) error {
	header, tableBody := createSetupTableRows(setups)

	t := table.Render(tableBody, table.WithHeader(header))

	_, err := fmt.Fprintln(out, t)
	return err
}

func createSetupTableRows(setups []string) (*table.Row, *table.Body) {
	header := table.NewRowFromStrings("Setup")

	tableBody := table.NewBody()
	for _, setup := range setups {
		tableBody.AddStringsRow(setup)
	}

	return header, tableBody
}
