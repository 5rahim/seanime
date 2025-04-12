package core

import (
	"fmt"
	"os"
	"seanime/internal/constants"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

func PrintHeader() {
	// Get terminal width
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	// Color scheme
	// primary := lipgloss.Color("#7B61FF")
	// secondary := lipgloss.Color("#5243CB")
	// highlight := lipgloss.Color("#14F9D5")
	// versionBgColor := lipgloss.Color("#8A2BE2")
	subtle := lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}

	// Base styles
	docStyle := lipgloss.NewStyle().Padding(1, 2)
	if physicalWidth > 0 {
		docStyle = docStyle.MaxWidth(physicalWidth)
	}

	// Build the header
	doc := strings.Builder{}

	// Logo with gradient effect
	logoStyle := lipgloss.NewStyle().Bold(true)
	logoLines := strings.Split(asciiLogo(), "\n")

	// Create a gradient effect for the logo
	gradientColors := []string{"#9370DB", "#8A2BE2", "#7B68EE", "#6A5ACD", "#5243CB"}
	for i, line := range logoLines {
		colorIdx := i % len(gradientColors)
		coloredLine := logoStyle.Foreground(lipgloss.Color(gradientColors[colorIdx])).Render(line)
		doc.WriteString(coloredLine + "\n")
	}

	// App name and version with box
	titleBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(subtle).
		Foreground(lipgloss.Color("#FFF7DB")).
		// Background(secondary).
		Padding(0, 1).
		Bold(true).
		Render("Seanime")

	versionBox := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(subtle).
		Foreground(lipgloss.Color("#ed4760")).
		// Background(versionBgColor).
		Padding(0, 1).
		Bold(true).
		Render(constants.Version)

	// Version name with different style
	versionName := lipgloss.NewStyle().
		Italic(true).
		Border(lipgloss.NormalBorder()).
		BorderForeground(subtle).
		Foreground(lipgloss.Color("#FFF7DB")).
		// Background(versionBgColor).
		Padding(0, 1).
		Render(constants.VersionName)

	// Combine title elements
	titleRow := lipgloss.JoinHorizontal(lipgloss.Center, titleBox, versionBox, versionName)

	// Add a decorative line
	// lineWidth := min(80, physicalWidth-4)
	// line := lipgloss.NewStyle().
	// 	Foreground(subtle).
	// 	Render(strings.Repeat("─", lineWidth))

	// Put it all together
	doc.WriteString("\n" +
		lipgloss.NewStyle().Align(lipgloss.Center).Render(titleRow))

	// Print the result
	fmt.Println(docStyle.Render(doc.String()))
}

// func asciiLogo() string {
// 	return `⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
// ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⣿⣿⠀⠀⠀⢠⣾⣧⣤⡖⠀⠀⠀⠀⠀⠀⠀
// ⠀⠀⠀⠀⠀⠀⠀⠀⢀⣼⠋⠀⠉⠀⢄⣸⣿⣿⣿⣿⣿⣥⡤⢶⣿⣦⣀⡀
// ⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⡆⠀⠀⠀⣙⣛⣿⣿⣿⣿⡏⠀⠀⣀⣿⣿⣿⡟
// ⠀⠀⠀⠀⠀⠀⠀⠀⠙⠻⠷⣦⣤⣤⣬⣽⣿⣿⣿⣿⣿⣿⣿⣟⠛⠿⠋⠀
// ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣴⠋⣿⣿⣿⣿⣿⣿⣿⣿⢿⣿⣿⡆⠀⠀
// ⠀⠀⠀⠀⣠⣶⣶⣶⣿⣦⡀⠘⣿⣿⣿⣿⣿⣿⣿⣿⠿⠋⠈⢹⡏⠁⠀⠀
// ⠀⠀⠀⢀⣿⡏⠉⠿⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⡆⠀⢀⣿⡇⠀⠀⠀
// ⠀⠀⠀⢸⣿⠀⠀⠀⠀⠀⠙⢿⣿⣿⣿⣿⣿⣿⣿⣿⣟⡘⣿⣿⣃⠀⠀⠀
// ⣴⣷⣀⣸⣿⠀⠀⠀⠀⠀⠀⠘⣿⣿⣿⣿⠹⣿⣯⣤⣾⠏⠉⠉⠉⠙⠢⠀
// ⠈⠙⢿⣿⡟⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣄⠛⠉⢩⣷⣴⡆⠀⠀⠀⠀⠀
// ⠀⠀⠀⠋⠀⠀⠀⠀⠀⠀⠀⠀⠈⣿⣿⣿⣿⣀⡠⠋⠈⢿⣇⠀⠀⠀⠀⠀
// ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠿⠿⠛⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀
// ⠀⠀⠀⠀⠀`
// }

func asciiLogo() string {
	return `⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣠⣴⡇⠀⠀⠀
⠀⢸⣿⣿⣶⣦⣤⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣴⣶⣿⣿⣿⣿⣿⡇⠀⠀⠀
⠀⠘⣿⣿⣿⣿⣿⣿⣿⣷⣦⣄⠀⠀⠀⣠⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀
⠀⠀⠹⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣄⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠀⠀⠀⠀
⠀⠀⠀⠘⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠏⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠉⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⡻⣿⣿⣿⠟⠋⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠻⣿⣿⣿⣿⣿⡌⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣿⣿⣿⣿⣿⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⢀⣠⣤⣴⣶⣶⣶⣦⣤⣤⣄⣉⡉⠛⠷⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⢀⣴⣾⣿⣿⣿⣿⡿⠿⠿⠿⣿⣿⣿⣿⣿⣿⣶⣦⣤⣀⡀⠀⠀⠀⠀⠀⠀
⠀⠀ ⠉⠉⠀⠀⠉⠉⠀⠀  ⠉ ⠉⠉⠉⠉⠉⠉⠉⠛⠛⠛⠲⠦⠄`
}
