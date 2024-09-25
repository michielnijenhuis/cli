package cli

import (
	"fmt"
)

func ScrollBar(visible []string, firstVisible int, height int, total int, width int, color string) []string {
	if height >= total {
		return visible
	}

	scrollPosition := scrollPosition(firstVisible, height, total)

	list := make([]string, 0, len(visible))
	for i, v := range visible {
		line := Pad(v, width, " ")
		lineColor := color

		if i != scrollPosition {
			lineColor = "gray"
		} else if lineColor == "" {
			lineColor = "cyan"
		}

		length := len(line)
		line = line[:length-1] + fmt.Sprintf("<fg=%s>%s</>", lineColor, LineVertical)
		list = append(list, line)
	}

	return list
}

func scrollPosition(firstVisible int, height int, total int) int {
	if firstVisible == 0 {
		return 0
	}

	maxPos := total - height

	if firstVisible == maxPos {
		return height - 1
	}

	if height <= 2 {
		return -1
	}

	percent := firstVisible / maxPos

	return (percent * (height - 3)) + 1
}
