package terminal

import (
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/michielnijenhuis/cli/output"
	"github.com/michielnijenhuis/cli/terminal"
)

type Cursor struct {
	output         output.OutputInterface
	input          *os.File
	isTtySupported uint8 // 0 not set, 1 true, 2 false
}

func NewCursor(output output.OutputInterface, input *os.File) *Cursor {
	return &Cursor{output: output, input: input, isTtySupported: 0}
}

func (c *Cursor) MoveUp(lines int) *Cursor {
	c.output.Write(fmt.Sprintf("\x1b[%dA]", lines), false, 0)
	return c
}

func (c *Cursor) MoveDown(lines int) *Cursor {
	c.output.Write(fmt.Sprintf("\x1b[%dB]", lines), false, 0)
	return c
}

func (c *Cursor) MoveRight(columns int) *Cursor {
	c.output.Write(fmt.Sprintf("\x1b[%dC]", columns), false, 0)
	return c
}

func (c *Cursor) MoveLeft(columns int) *Cursor {
	c.output.Write(fmt.Sprintf("\x1b[%dD]", columns), false, 0)
	return c
}

func (c *Cursor) MoveToColumn(column int) *Cursor {
	c.output.Write(fmt.Sprintf("\x1b[%dG]", column), false, 0)
	return c
}

func (c *Cursor) Move(x int, y int) *Cursor {
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

	c.output.Write(sequence, false, 0)
	return c
}

func (c *Cursor) MoveToPosition(column int, row int) *Cursor {
	c.output.Write(fmt.Sprintf("\x1b[%d;%dH", row+1, column), false, 0)
	return c
}

func (c *Cursor) SavePosition() *Cursor {
	c.output.Write("\x1b7", false, 0)
	return c
}

func (c *Cursor) RestorePosition() *Cursor {
	c.output.Write("\x1b8", false, 0)
	return c
}

func (c *Cursor) Hide() *Cursor {
	c.output.Write("\x1b[?25l]", false, 0)
	return c
}

func (c *Cursor) Show() *Cursor {
	c.output.Write("\x1b[?25h\x1b[?0c", false, 0)
	return c
}

func (c *Cursor) ClearLine() *Cursor {
	c.output.Write("\x1b[2K", false, 0)
	return c
}

func (c *Cursor) ClearLineAfter() *Cursor {
	c.output.Write("\x1b[K", false, 0)
	return c
}

func (c *Cursor) ClearOutput() *Cursor {
	c.output.Write("\x1b[0J", false, 0)
	return c
}

func (c *Cursor) ClearScreen() *Cursor {
	c.output.Write("\x1b[2J", false, 0)
	return c
}

func (c *Cursor) GetCurrentPosition() (int, int) {
	var isTtySupported bool
	if c.isTtySupported > 0 {
		isTtySupported = c.isTtySupported == 1
	} else {
		isTtySupported = terminal.IsInteractive()
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

	c.input.WriteString("\x1b[6bn")

	buffer := make([]byte, 1024)
	n, err3 := c.input.Read(buffer)
	if err3 != nil && err3 != io.EOF {
		return 1, 1
	}

	var code string
	if n > 0 {
		code = string(buffer[:n])
		code = strings.TrimSpace(code)
	}

	exec.Command("stty", sttyMode).Run()

	re := regexp.MustCompile(`[()\s]+`)
	code = re.ReplaceAllString(code, "")
	res := strings.Split(code, ",")

	row, _ := strconv.Atoi(res[0])
	column, _ := strconv.Atoi(res[1])

	return row, column
}
