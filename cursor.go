package cli

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Cursor struct {
	Output         *Output
	Input          *os.File
	isTtySupported uint8 // 0 not set, 1 true, 2 false
}

func (c *Cursor) MoveUp(lines int) *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write(fmt.Sprintf("\x1b[%dA]", lines), false, 0)
	return c
}

func (c *Cursor) MoveDown(lines int) *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write(fmt.Sprintf("\x1b[%dB]", lines), false, 0)
	return c
}

func (c *Cursor) MoveRight(columns int) *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write(fmt.Sprintf("\x1b[%dC]", columns), false, 0)
	return c
}

func (c *Cursor) MoveLeft(columns int) *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write(fmt.Sprintf("\x1b[%dD]", columns), false, 0)
	return c
}

func (c *Cursor) MoveToColumn(column int) *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write(fmt.Sprintf("\x1b[%dG]", column), false, 0)
	return c
}

func (c *Cursor) Move(x int, y int) *Cursor {
	checkPtr(c.Output, "cursor output")
	var sequence string

	if x < 0 {
		sequence += "\x1b[" + strconv.Itoa(int(math.Abs(float64(x)))) + "D" // Left
	} else if x > 0 {
		sequence += "\x1b[" + strconv.Itoa(x) + "C" // Right
	}

	if y < 0 {
		sequence += "\x1b[" + strconv.Itoa(int(math.Abs(float64(y)))) + "A" // Up
	} else if y > 0 {
		sequence += "\x1b[" + strconv.Itoa(y) + "B" // Down
	}

	c.Output.Write(sequence, false, 0)
	return c
}

func (c *Cursor) MoveToPosition(column int, row int) *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write(fmt.Sprintf("\x1b[%d;%dH", row+1, column), false, 0)
	return c
}

func (c *Cursor) SavePosition() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b7", false, 0)
	return c
}

func (c *Cursor) RestorePosition() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b8", false, 0)
	return c
}

func (c *Cursor) Hide() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b[?25l]", false, 0)
	return c
}

func (c *Cursor) Show() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b[?25h\x1b[?0c", false, 0)
	return c
}

func (c *Cursor) ClearLine() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b[2K", false, 0)
	return c
}

func (c *Cursor) ClearLineAfter() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b[K", false, 0)
	return c
}

func (c *Cursor) ClearOutput() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b[0J", false, 0)
	return c
}

func (c *Cursor) ClearScreen() *Cursor {
	checkPtr(c.Output, "cursor output")
	c.Output.Write("\x1b[2J", false, 0)
	return c
}

func (c *Cursor) CurrentPosition() (int, int) {
	var isTtySupported bool
	if c.isTtySupported > 0 {
		isTtySupported = c.isTtySupported == 1
	} else {
		isTtySupported = TerminalIsInteractive()
		if isTtySupported {
			c.isTtySupported = 1
		} else {
			c.isTtySupported = 2
		}
	}

	if !isTtySupported {
		return 1, 1
	}

	sttyModeCmd := exec.Command("stty", "-g")
	sttyModeOutput, err := sttyModeCmd.Output()
	if err != nil {
		return 1, 1
	}
	sttyMode := string(sttyModeOutput)

	_, err2 := exec.Command("stty", "-icanon -echo").Output()
	if err2 != nil {
		return 1, 1
	}

	_, err = c.Input.WriteString("\x1b[6bn")
	if err != nil {
		return 1, 1
	}

	buffer := make([]byte, 1024)
	n, err3 := c.Input.Read(buffer)
	if err3 != nil && err3 != io.EOF {
		return 1, 1
	}

	var code string
	if n > 0 {
		code = string(buffer[:n])
		code = strings.TrimSpace(code)
	}

	err = exec.Command("stty", sttyMode).Run()
	if err != nil {
		return 1, 1
	}

	re := regexp.MustCompile(`[()\s]+`)
	code = re.ReplaceAllString(code, "")
	res := strings.Split(code, ",")

	row, _ := strconv.Atoi(res[0])
	column, _ := strconv.Atoi(res[1])

	return row, column
}
