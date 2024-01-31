package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/zeropsio/zcli/src/logger"
	. "github.com/zeropsio/zcli/src/uxBlock"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	regSignals(cancel)

	blocks, err := createBlocks(cancel)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = do(ctx, blocks)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func do(ctx context.Context, blocks UxBlocks) error {
	prompts(ctx, blocks)
	spinners(ctx, blocks)
	texts(ctx, blocks)
	tables(ctx, blocks)

	return nil
}

func spinners(ctx context.Context, blocks UxBlocks) {
	{
		fmt.Println("========= spinners block =========")

		spinner1 := NewSpinner(InfoText("Running 1").String())
		spinner2 := NewSpinner(InfoText("Running 2").String())
		spinner3 := NewSpinner(InfoText("Running 3").String())

		stop := blocks.RunSpinners(ctx, []*Spinner{spinner1, spinner2, spinner3})

		counter := 0
		tick := time.NewTicker(time.Second * 1)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				counter++
				if counter == 2 {
					spinner2.Finish(NewLine(SuccessIcon, SuccessText("Finished successfully")).String())
				}
				if counter == 4 {
					spinner1.Finish(NewLine(ErrorIcon, ErrorText("finished with error")).String())
				}
				if counter == 6 {
					spinner3.Finish(NewLine(WarningIcon, WarningText("Finish with warning")).String())
				}
			}
			if counter == 6 {
				break
			}
		}

		stop()

		fmt.Println("========= spinners block end =========")
	}
}

func prompts(ctx context.Context, blocks UxBlocks) {
	fmt.Println("========= prompt block =========")
	choices := []string{"yes", "no", "maybe"}
	choice, err := blocks.Prompt(ctx, "Question?", choices)
	if err != nil {
		return
	}

	blocks.PrintInfoLine("selected", choices[choice])

	fmt.Println("========= prompt block end =========")
}

func texts(ctx context.Context, blocks UxBlocks) {
	fmt.Println("========= texts block =========")
	blocks.PrintInfoLine("info line")
	blocks.PrintWarningLine("warning line")
	blocks.PrintErrorLine("Error line")
	blocks.PrintLine(WarningText("Line"), " with ", InfoIcon, InfoText("mixed"), " ", ErrorText("styles"))
	fmt.Println("========= texts block end =========")
}

func tables(ctx context.Context, blocks UxBlocks) {
	fmt.Println("========= table selection block =========")

	tableData := [][]string{
		{"lorem", "ipsum", "dolor", "sit"},
		{"amet", "consectetur", "adipiscing", "elit"},
		{"sed", "do", "eiusmod", "tempor"},
		{"incididunt", "ut", "labore", "et"},
	}

	body := NewTableBody().AddStringsRows(tableData...)

	line, err := blocks.Select(
		ctx,
		body,
		SelectTableHeader(NewTableRow().AddStringCells("header1", "header2", "header3", "header4")),
		SelectLabel("Select line"),
	)
	if err != nil {
		return
	}

	blocks.PrintInfoLine("selected", tableData[line[0]][0])

	fmt.Println("========= table selection end =========")
	fmt.Println("========= print table block =========")

	blocks.PrintInfoLine("printing table")
	blocks.Table(body, WithTableHeader(NewTableRow().AddStringCells("header1", "header2", "header3", "header4")))

	fmt.Println("========= print table block end =========")
}

func regSignals(contextCancel func()) {
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		contextCancel()
	}()
}

func createBlocks(contextCancelFunc func()) (UxBlocks, error) {
	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
	})

	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: "zerops.log",
	})

	blocks := NewBlock(outputLogger, debugFileLogger, isTerminal, contextCancelFunc)

	return blocks, nil
}
