package kongCmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/alecthomas/kong"

	"github.com/zeropsio/zcli/src/uxBlock/styles"
)

// Help is the kong.HelpPrinter for zcli. Mirrors the cobra-era root template
// shape (Usage / Available Commands / Flags / Global Flags / "Use ... --help"
// hint) with the same lipgloss styling.
func Help(_ kong.HelpOptions, ctx *kong.Context) error {
	out := ctx.Stdout

	app := ctx.Model
	selected := ctx.Selected()
	node := app.Node
	if selected != nil {
		node = selected
	}

	writeUsage(out, app, node)

	if children := visibleChildren(node); len(children) > 0 {
		fmt.Fprintln(out)
		writeSection(out, "Available Commands:")
		writeChildren(out, children)
	}

	localFlags, globalFlags := splitFlags(node)
	if len(localFlags) > 0 {
		fmt.Fprintln(out)
		writeSection(out, "Flags:")
		writeFlags(out, localFlags)
	}
	if len(globalFlags) > 0 {
		fmt.Fprintln(out)
		writeSection(out, "Global Flags:")
		writeFlags(out, globalFlags)
	}

	if hasChildren(node) {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Use \"%s [command] --help\" for more information about a command.\n", node.Path())
	}
	return nil
}

func writeSection(out io.Writer, label string) {
	fmt.Fprintln(out, styles.CobraSectionStyle().SetString(label).String())
}

func writeUsage(out io.Writer, app *kong.Application, node *kong.Node) {
	writeSection(out, "Usage:")
	summary := strings.TrimSpace(node.Summary())
	if summary == "" {
		fmt.Fprintf(out, "%s\n", app.Name)
		return
	}
	fmt.Fprintf(out, "%s %s\n", app.Name, summary)
}

func visibleChildren(node *kong.Node) []*kong.Node {
	out := make([]*kong.Node, 0, len(node.Children))
	for _, c := range node.Children {
		if c.Hidden {
			continue
		}
		out = append(out, c)
	}
	return out
}

func writeChildren(out io.Writer, children []*kong.Node) {
	pad := 0
	for _, c := range children {
		if l := len(c.Name); l > pad {
			pad = l
		}
	}
	for _, c := range children {
		name := styles.CobraItemNameStyle().SetString(c.Name).String()
		fmt.Fprintf(out, "%s%s  %s\n", name, strings.Repeat(" ", pad-len(c.Name)), c.Help)
	}
}

func hasChildren(node *kong.Node) bool {
	for _, c := range node.Children {
		if !c.Hidden {
			return true
		}
	}
	return false
}

// splitFlags separates flags into local (declared on the active node) and
// global (inherited from ancestors). The cobra template distinguished the two
// visually; we preserve that. Hidden flags are dropped.
//
// Uses kong's Node.AllFlags which returns flag levels root-first, leaf-last.
func splitFlags(node *kong.Node) (local, global []*kong.Flag) {
	levels := node.AllFlags(true)
	if len(levels) == 0 {
		return nil, nil
	}
	local = levels[len(levels)-1]
	for _, lvl := range levels[:len(levels)-1] {
		global = append(global, lvl...)
	}
	return local, global
}

func writeFlags(out io.Writer, flags []*kong.Flag) {
	pad := 0
	labels := make([]string, len(flags))
	for i, f := range flags {
		labels[i] = flagLabel(f)
		if l := len(labels[i]); l > pad {
			pad = l
		}
	}
	for i, f := range flags {
		fmt.Fprintf(out, "  %s%s  %s\n", labels[i], strings.Repeat(" ", pad-len(labels[i])), f.Help)
	}
}

func flagLabel(f *kong.Flag) string {
	switch {
	case f.Short != 0:
		return fmt.Sprintf("-%c, --%s", f.Short, f.Name)
	default:
		return fmt.Sprintf("    --%s", f.Name)
	}
}
