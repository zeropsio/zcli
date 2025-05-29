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
	"github.com/zeropsio/zcli/src/uxBlock"
	"github.com/zeropsio/zcli/src/uxBlock/models/prompt"
	"github.com/zeropsio/zcli/src/uxBlock/models/selector"
	"github.com/zeropsio/zcli/src/uxBlock/models/table"
	. "github.com/zeropsio/zcli/src/uxBlock/styles"
	"golang.org/x/term"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	regSignals(cancel)

	blocks := createBlocks(cancel)

	do(ctx, blocks)
}

func do(ctx context.Context, blocks uxBlock.UxBlocks) {
	prompts(ctx, blocks)
	spinners(ctx, blocks)
	texts(ctx, blocks)
	tables(ctx, blocks)
}

func spinners(ctx context.Context, blocks uxBlock.UxBlocks) {
	{
		fmt.Println("========= spinners block =========")

		spinner1 := uxBlock.NewSpinner(NewLine("Running 1").String())
		spinner2 := uxBlock.NewSpinner(NewLine("Running 2").String())
		spinner3 := uxBlock.NewSpinner(NewLine("Running 3").String())

		stop, send := blocks.RunSpinners(ctx, []*uxBlock.Spinner{spinner1, spinner2, spinner3})

		counter := 0
		tick := time.NewTicker(time.Millisecond * 300)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				counter++
				if counter == 2 {
					send(spinner2.FinishWithLine(SuccessLine("Finished successfully").String()))
				}
				if counter == 4 {
					send(spinner1.FinishWithLine(ErrorLine("finished with error").String()))
				}
				if counter == 6 {
					send(spinner3.FinishWithLine(WarningLine("Finish with warning").String()))
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

func prompts(ctx context.Context, blocks uxBlock.UxBlocks) {
	fmt.Println("========= prompt block =========")
	choices := []string{"yes", "no", "maybe"}
	choice, err := uxBlock.Run(
		prompt.NewRoot(
			ctx,
			"Question?",
			choices,
		),
		prompt.GetChoiceCursor,
	)
	if err != nil {
		return
	}

	blocks.PrintInfo(InfoWithValueLine("selected", choices[choice]))

	fmt.Println("========= prompt block end =========")
}

func texts(_ context.Context, blocks uxBlock.UxBlocks) {
	fmt.Println("========= texts block =========")
	blocks.PrintInfo(NewLine("line without style"))
	blocks.PrintInfo(InfoLine("info line"))
	blocks.PrintInfo(InfoWithValueLine("info line", "value"))
	blocks.PrintInfo(SuccessLine("success line"))
	blocks.PrintWarning(WarningLine("warning line"))
	blocks.PrintError(ErrorLine("Error line"))
	blocks.PrintInfo(SelectLine("Select line"))
	blocks.PrintInfo(NewLine(WarningText("NewLine"), " with ", InfoIcon, InfoText("mixed"), " ", ErrorText("styles")))
	fmt.Println("========= texts block end =========")
}

func tables(ctx context.Context, blocks uxBlock.UxBlocks) {
	tableData := [][]string{
		{"lorem", "ipsum", "dolor", "sit"},
		{
			"amet",
			"lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua",
			"adipiscing",
			"elit",
		},
		{"sed", "do", "eiusmod", "tempor"},
		{"incididunt", "ut", "labore", "et"},
	}
	body := table.NewBody().AddStringsRows(tableData...)
	header := table.NewRowFromStrings("header1", "header2", "header3", "header4")

	fmt.Println("========= table single selection block =========")

	line, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel("Single select"),
			selector.WithHeader(header),
			selector.WithEnableFiltering(),
		),
		selector.GetOneSelectedFunc,
	)
	if err != nil {
		return
	}

	blocks.PrintInfo(InfoWithValueLine("selected", tableData[line][0]))

	fmt.Println("========= table single selection end =========")
	fmt.Println("========= table multi selection block =========")

	lines, err := uxBlock.Run(
		selector.NewRoot(
			ctx,
			body,
			selector.WithLabel("Multi select"),
			selector.WithHeader(header),
			selector.WithEnableMultiSelect(),
			selector.WithEnableFiltering(),
		),
		selector.GetMultipleSelectedFunc,
	)
	if err != nil {
		return
	}

	for _, line := range lines {
		blocks.PrintInfo(InfoWithValueLine("selected", tableData[line][0]))
	}

	fmt.Println("========= table multi selection end =========")
	fmt.Println("========= print table block =========")

	blocks.PrintInfo(InfoLine("printing table"))
	fmt.Println(table.Render(body, table.WithHeader(header)))

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

func createBlocks(contextCancelFunc func()) uxBlock.UxBlocks {
	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	outputLogger := logger.NewOutputLogger(logger.OutputConfig{
		IsTerminal: isTerminal,
	})

	width, height, err := term.GetSize(0)
	if err != nil {
		width = 100
		height = 25
	}

	debugFileLogger := logger.NewDebugFileLogger(logger.DebugFileConfig{
		FilePath: "/tmp/zerops-showcase.log",
	})

	blocks := uxBlock.NewBlocks(outputLogger, debugFileLogger, isTerminal, width, height, contextCancelFunc)

	return blocks
}
