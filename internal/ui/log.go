package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	blue    = lipgloss.NewStyle().Foreground(lipgloss.Color("4"))
	green   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	yellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	red     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	magenta = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	cyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	bold    = lipgloss.NewStyle().Bold(true)
)

func Info(msg string) {
	fmt.Printf("%s %s\n", blue.Render("[INFO]"), msg)
}

func Success(msg string) {
	fmt.Printf("%s %s\n", green.Render("[OK]"), msg)
}

func Warn(msg string) {
	fmt.Printf("%s %s\n", yellow.Render("[WARN]"), msg)
}

func Error(msg string) {
	fmt.Printf("%s %s\n", red.Render("[ERROR]"), msg)
}

func Debug(verbose bool, msg string) {
	if verbose {
		fmt.Printf("%s\n", dim.Render("[DEBUG] "+msg))
	}
}

func Banner(engine, source, file string) {
	fmt.Println(bold.Render("============================================"))
	fmt.Println(bold.Render("Ralphy") + " - Running until PRD is complete")
	fmt.Printf("Engine: %s\n", engineColor(engine).Render(engine))
	fmt.Printf("Source: %s (%s)\n", cyan.Render(source), file)
	fmt.Println(bold.Render("============================================"))
}

func engineColor(engine string) lipgloss.Style {
	switch engine {
	case "opencode":
		return cyan
	case "cursor":
		return yellow
	case "codex":
		return blue
	default:
		return magenta
	}
}

func TaskHeader(iteration, completed, remaining int) {
	fmt.Println()
	fmt.Println(bold.Render(fmt.Sprintf(">>> Task %d", iteration)))
	fmt.Printf("%s\n", dim.Render(fmt.Sprintf("    Completed: %d | Remaining: %d", completed, remaining)))
	fmt.Println("--------------------------------------------")
}

func TaskDone(task string) {
	fmt.Printf("  %s %-16s │ %s\n", green.Render("✓"), "Done", truncate(task, 40))
}

func TaskFailed(task string) {
	fmt.Printf("  %s %-16s │ %s\n", red.Render("✗"), "Failed", truncate(task, 40))
}

func Summary(iteration, inputTokens, outputTokens int, cost float64) {
	fmt.Println()
	fmt.Println(bold.Render("============================================"))
	fmt.Printf("%s Finished %d task(s).\n", green.Render("PRD complete!"), iteration)
	fmt.Println(bold.Render("============================================"))
	fmt.Println()
	fmt.Println(bold.Render(">>> Cost Summary"))
	fmt.Printf("Input tokens:  %d\n", inputTokens)
	fmt.Printf("Output tokens: %d\n", outputTokens)
	fmt.Printf("Total tokens:  %d\n", inputTokens+outputTokens)
	if cost > 0 {
		fmt.Printf("Est. cost:     $%.4f\n", cost)
	}
	fmt.Println(bold.Render("============================================"))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
