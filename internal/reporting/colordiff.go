package reporting

import (
	"strings"

	"github.com/fatih/color"
)

func ColorizeDiff(diff string) string {
	coloredDiff := &strings.Builder{}
	lines := strings.Split(diff, "\n")
	colorBold := color.New(color.Bold)
	colorBold.EnableColor()
	colorCyan := color.New(color.FgCyan)
	colorCyan.EnableColor()
	colorRed := color.New(color.FgRed)
	colorRed.EnableColor()
	colorGreen := color.New(color.FgGreen)
	colorGreen.EnableColor()

	for i, line := range lines {
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") {
			colorBold.Fprint(coloredDiff, line)
		} else if strings.HasPrefix(line, "@@") {
			colorCyan.Fprint(coloredDiff, line)
		} else if strings.HasPrefix(line, "-") {
			colorRed.Fprint(coloredDiff, line)
		} else if strings.HasPrefix(line, "+") {
			colorGreen.Fprint(coloredDiff, line)
		} else {
			coloredDiff.WriteString(line)
		}
		if i < len(lines)-1 { // Avoid adding an extra newline at the very end
			coloredDiff.WriteString("\n")
		}
	}
	return coloredDiff.String()
}
