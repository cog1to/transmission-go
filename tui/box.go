package tui

import (
	"fmt"
	"strings"
)

const(
	ASC_HLINE string			 = "q"
	ASC_VLINE							 = "x"
	ASC_TOPLEFT_CORNER		 = "l"
	ASC_TOPRIGHT_CORNER		 = "k"
	ASC_BOTTOMLEFT_CORNER	 = "m"
	ASC_BOTTOMRIGHT_CORNER = "j"
)

func HLine(x, y, length int) {
	MovePrintf(x, y, "%s%s%s", ESC_BEGIN_BOXDRAW, strings.Repeat(ASC_HLINE, length), ESC_END_BOXDRAW)
}

func VLine(x, y, length int) {
	fmt.Printf(ESC_BEGIN_BOXDRAW)
	for line := x; line < x + length; line++ {
		MovePrintf(line, y, ASC_VLINE)
	}
	fmt.Printf(ESC_END_BOXDRAW)
}

func Corner(x, y int, left, top bool) {
	symbol := ASC_TOPLEFT_CORNER
	if (!left && top) {
		symbol = ASC_TOPRIGHT_CORNER
	} else if (!left && !top) {
		symbol = ASC_BOTTOMRIGHT_CORNER
	} else if (left && !top) {
		symbol = ASC_BOTTOMLEFT_CORNER
	}

	MovePrintf(x, y, "%s%s%s", ESC_BEGIN_BOXDRAW, symbol, ESC_END_BOXDRAW)
}

func Box(x, y, height, width int) {
	repeatY := func(startx, starty, count int, value string) string {
		var input string

		for line := 0; line < count; line++ {
			input += fmt.Sprintf(string(ESC_MOVE_TO), startx + line, starty)
			input += value
		}

		return input
	}

	fmt.Printf(
		ESC_BEGIN_BOXDRAW +
		fmt.Sprintf(string(ESC_MOVE_TO), x + 1, y + 1) +
		ASC_TOPLEFT_CORNER + strings.Repeat(ASC_HLINE, width - 2) + ASC_TOPRIGHT_CORNER + "\n" +
		repeatY(x + 2, y + 1, height - 2, ASC_VLINE + strings.Repeat(" ", width - 2) + ASC_VLINE + "\n") +
		fmt.Sprintf(string(ESC_MOVE_TO), x + height, y + 1) +
		ASC_BOTTOMLEFT_CORNER + strings.Repeat(ASC_HLINE, width - 2) + ASC_BOTTOMRIGHT_CORNER +
		ESC_END_BOXDRAW)
}

func ClearBox(x, y, height, width int) {
	repeatY := func(startx, starty, count int, value string) string {
		var input string

		for line := 0; line < count; line++ {
			input += fmt.Sprintf(string(ESC_MOVE_TO), startx + line, starty)
			input += value
		}

		return input
	}

	fmt.Printf(repeatY(x + 1, y + 1, height, strings.Repeat(" ", width) + ASC_VLINE + "\n"))
}
