package cli

import (
	"fmt"
	"regexp"
)

func ScrollBar(visible []string, firstVisible int, height int, total int, width int, color string) []string {
	if height >= total {
		return visible
	}

	scrollPosition := scrollPosition(firstVisible, height, total)

	list := make([]string, 0, len(visible))
	for i, v := range visible {
		line := Pad(v, width, " ")
		if i != scrollPosition {
			color = "gray"
		}

		re := regexp.MustCompile(`\.`)
		list = append(list, re.ReplaceAllString(line, fmt.Sprintf("<fg=%s>|</>", color)))
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
