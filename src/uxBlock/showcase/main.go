package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/zeropsio/zcli/src/logfile"
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

func do(ctx context.Context, blocks *uxBlock.Blocks) {
	prompts(ctx, blocks)
	spinners(ctx, blocks)
	withSpinner(ctx, blocks)
	texts(ctx, blocks)
	tables(ctx, blocks)
}

func withSpinner(ctx context.Context, blocks *uxBlock.Blocks) {
	fmt.Println("========= WithSpinner block =========")

	// Set a persistent header that survives every spinner in this block.
	blocks.SetHeader(InfoWithValueLine("scope", "project=demo  service=api  region=prg1").String())
	defer blocks.SetHeader("")

	// 1) Successful phase with SetMessage transitions.
	_ = blocks.WithSpinner(ctx, uxBlock.SpinnerSpec{
		Message:        "preparing",
		SuccessMessage: "deployed",
	}, func(h uxBlock.SpinnerHandle) error {
		time.Sleep(800 * time.Millisecond)
		h.SetMessage("uploading package")
		time.Sleep(800 * time.Millisecond)
		h.SetMessage("running deploy")
		time.Sleep(800 * time.Millisecond)
		return nil
	})

	// 2) Error path with custom ErrorMessage.
	_ = blocks.WithSpinner(ctx, uxBlock.SpinnerSpec{
		Message:      "running risky thing",
		ErrorMessage: "risky thing failed",
	}, func(_ uxBlock.SpinnerHandle) error {
		time.Sleep(800 * time.Millisecond)
		return fmt.Errorf("oh no")
	})

	// 3) Error path with no ErrorMessage — falls back to err.Error().
	_ = blocks.WithSpinner(ctx, uxBlock.SpinnerSpec{
		Message: "trying something",
	}, func(_ uxBlock.SpinnerHandle) error {
		time.Sleep(800 * time.Millisecond)
		return fmt.Errorf("something broke at step 42")
	})

	// 4) Live byte counter via ProgressReader. Simulates a 4 MiB
	// upload by reading from a slow source; the spinner line ticks
	// every 500ms with bytes consumed.
	_ = blocks.WithSpinner(ctx, uxBlock.SpinnerSpec{
		Message:        "uploading package",
		SuccessMessage: "uploaded",
	}, func(h uxBlock.SpinnerHandle) error {
		const total = 4 * 1024 * 1024
		source := &slowReader{remaining: total, perChunk: 64 * 1024, delay: 80 * time.Millisecond}
		reader, stop := uxBlock.ProgressReader(h, "uploading package", source)
		defer stop()
		_, err := io.Copy(io.Discard, reader)
		return err
	})

	fmt.Println("========= WithSpinner block end =========")
}

// slowReader simulates a network read: hands out perChunk bytes every
// delay until remaining is exhausted.
type slowReader struct {
	remaining int
	perChunk  int
	delay     time.Duration
}

func (s *slowReader) Read(p []byte) (int, error) {
	if s.remaining <= 0 {
		return 0, io.EOF
	}
	time.Sleep(s.delay)
	n := s.perChunk
	if n > len(p) {
		n = len(p)
	}
	if n > s.remaining {
		n = s.remaining
	}
	for i := range n {
		p[i] = 'x'
	}
	s.remaining -= n
	return n, nil
}

func spinners(ctx context.Context, blocks *uxBlock.Blocks) {
	fmt.Println("========= spinners block =========")

	// Three concurrent jobs, each with its own outcome rendered in
	// place on its spinner. Order of completion is independent.
	_ = blocks.WithSpinners(ctx, []uxBlock.SpinnerJob{
		{
			Spec: uxBlock.SpinnerSpec{
				Message:      "Running 1",
				ErrorMessage: "finished with error",
			},
			Run: func(_ uxBlock.SpinnerHandle) error {
				time.Sleep(1200 * time.Millisecond)
				return fmt.Errorf("simulated failure")
			},
		},
		{
			Spec: uxBlock.SpinnerSpec{
				Message:        "Running 2",
				SuccessMessage: "Finished successfully",
			},
			Run: func(_ uxBlock.SpinnerHandle) error {
				time.Sleep(600 * time.Millisecond)
				return nil
			},
		},
		{
			Spec: uxBlock.SpinnerSpec{
				Message:        "Running 3",
				SuccessMessage: "Finish with success", // rendered green; warning-styled outcome would require a custom hook.
			},
			Run: func(_ uxBlock.SpinnerHandle) error {
				time.Sleep(1800 * time.Millisecond)
				return nil
			},
		},
	})

	fmt.Println("========= spinners block end =========")
}

func prompts(ctx context.Context, blocks *uxBlock.Blocks) {
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

func texts(_ context.Context, blocks *uxBlock.Blocks) {
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

func tables(ctx context.Context, blocks *uxBlock.Blocks) {
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

func createBlocks(contextCancelFunc func()) *uxBlock.Blocks {
	isTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())

	width, height, err := term.GetSize(0)
	if err != nil {
		width = 100
		height = 25
	}

	slog.SetDefault(slog.New(logfile.New(logfile.Config{
		Path: "/tmp/zerops-showcase.log",
	})))

	blocks := uxBlock.NewBlocks(isTerminal, width, height, contextCancelFunc)

	return blocks
}
