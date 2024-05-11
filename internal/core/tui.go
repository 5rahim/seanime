package core

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/seanime-app/seanime/internal/constants"
	"golang.org/x/term"
	"os"
	"strings"
)

func PrintHeader() {

	//const (
	//	width       = 96
	//	columnWidth = 30
	//)

	var (
		subtle     = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
		titleStyle = lipgloss.NewStyle().
				MarginLeft(1).
				MarginRight(5).
				Padding(0, 1).
				Italic(true).
				Foreground(lipgloss.Color("#FFF7DB")).
				SetString("Seanime")
		//list = lipgloss.NewStyle().
		//	Border(lipgloss.NormalBorder(), false, true, false, false).
		//	BorderForeground(subtle).
		//	MarginRight(2).
		//	Height(8).
		//	Width(columnWidth + 1)
		docStyle = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	)

	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	doc := strings.Builder{}

	var (
		logo strings.Builder
	)

	//col := color.New(color.FgHiMagenta)
	fmt.Println()
	//col.Printf("\n        .-----.    \n       /    _ /  \n       \\_..`--.  \n       .-._)   \\ \n       \\ ")
	//col.Printf("      / \n        `-----'  \n")
	fmt.Fprint(&logo, lipgloss.NewStyle().Foreground(lipgloss.Color("#9f92ff")).SetString("\n      .-----.    \n     /    _ /  \n     \\_..`--.  \n     .-._)   \\ \n     \\       / \n      `-----'  \n"))
	doc.WriteString(logo.String() + "\n")

	{
		var (
			title  strings.Builder
			titles = []string{"Seanime", constants.Version, constants.VersionName}
			colors = []string{"#5243cb", "#5243cb", "#312887", "#14F9D5"}
		)

		for i, v := range titles {
			const offset = 4
			c := lipgloss.Color(colors[i])
			s := titleStyle.SetString(v).MarginLeft(i * offset).Background(c)
			fmt.Fprint(&title, s)
			if i < len(titles)-1 {
				title.WriteRune('\n')
			}
		}

		row := lipgloss.NewStyle().
			//BorderStyle(lipgloss.NormalBorder()).BorderTop(true).
			Padding(0, 1).BorderForeground(subtle).Render(title.String())
		doc.WriteString(row)
	}

	if physicalWidth > 0 {
		docStyle = docStyle.MaxWidth(physicalWidth)
	}

	fmt.Println(docStyle.Render(doc.String()))

}
