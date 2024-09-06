package cli

const (
	CbarBell     = "\x07"
	CharNewline  = "\x0a"
	CharTab      = "\x09"
	CharSpace    = " "
	CharEllipsis = "…"
)

const (
	BoxTopLeft        = "┌"
	BoxTopRight       = "┐"
	BoxBottomLeft     = "└"
	BoxBottomRight    = "┘"
	BoxVertical       = "│"
	BoxVerticalRight  = "├"
	BoxVerticalLeft   = "┤"
	BoxHorizontal     = "─"
	BoxHorizontalDown = "┬"
	BoxHorizontalUp   = "┴"
	BoxCross          = "┼"
)

const (
	HeavyBoxTopLeft        = "┏"
	HeavyBoxTopRight       = "┓"
	HeavyBoxBottomLeft     = "┗"
	HeavyBoxBottomRight    = "┛"
	HeavyBoxVertical       = "┃"
	HeavyBoxVerticalRight  = "┣"
	HeavyBoxVerticalLeft   = "┫"
	HeavyBoxHorizontal     = "━"
	HeavyBoxHorizontalDown = "┳"
	HeavyBoxHorizontalUp   = "┻"
	HeavyBoxCross          = "╋"
)

const (
	DoubleBoxTopLeft        = "╔"
	DoubleBoxTopRight       = "╗"
	DoubleBoxBottomLeft     = "╚"
	DoubleBoxBottomRight    = "╝"
	DoubleBoxVertical       = "║"
	DoubleBoxVerticalRight  = "╠"
	DoubleBoxVerticalLeft   = "╣"
	DoubleBoxHorizontal     = "═"
	DoubleBoxHorizontalDown = "╦"
	DoubleBoxHorizontalUp   = "╩"
	DoubleBoxCross          = "╬"
)

const (
	RoundedBoxTopLeft        = "╭"
	RoundedBoxTopRight       = "╮"
	RoundedBoxBottomLeft     = "╰"
	RoundedBoxBottomRight    = "╯"
	RoundedBoxVertical       = "│"
	RoundedBoxVerticalRight  = "├"
	RoundedBoxVerticalLeft   = "┤"
	RoundedBoxHorizontal     = "─"
	RoundedBoxHorizontalDown = "┬"
	RoundedBoxHorizontalUp   = "┴"
	RoundedBoxCross          = "┼"
)

const (
	ArrowRight     = "→"
	ArrowLeft      = "←"
	ArrowUp        = "↑"
	ArrowDown      = "↓"
	ArrowLeftRight = "↔"
	ArrowUpDown    = "↕"
)

const (
	TriangleRight = "▶"
	TriangleLeft  = "◀"
	TriangleUp    = "▲"
	TriangleDown  = "▼"
)

const (
	SmallTriangleRight = "▸"
	SmallTriangleLeft  = "◂"
	SmallTriangleUp    = "▴"
	SmallTriangleDown  = "▾"
)

const (
	LineDiagonalCross         = "╳"
	LineDiagonalbackwards     = "╲"
	LineDiagonal              = "╱"
	LineVerticalDashed        = "┆"
	LineVerticalDashedHeavy   = "┇"
	LineVertical              = "│"
	LineVerticalHeavy         = "┃"
	LineHorizontalDashed      = "┄"
	LineHorizontalDashedHeavy = "┅"
	LineHorizontal            = "─"
	LineHorizontalHeavy       = "━"
)

const (
	CircleFilled        = "●"
	CircleOutline       = "◯"
	CircleOutlineFilled = "◉"
	CircleDotted        = "◌"
	CircleDoubled       = "◎"
	CircleSmall         = "•"
	CircleHalfLeft      = "◐"
	CircleHalfTop       = "◓"
	CircleHalfRight     = "◑"
	CircleHalfBottom    = "◒"
)

const (
	ChevronDefault = ""
	ChevronSmall   = "›"
	ChevronHeavy   = "❯"
)

const (
	DiamondDefault = "◆"
	DiamondOutline = "◇"
)

const (
	SquareDefault = "■"
	SquareOutline = "☐"
	SquareCrossed = "☒"
)

const (
	HeartDefault = "❤︎"
	HeartOutline = "♥"
)

const (
	IconTick        = "✓"
	IconTickSwoosh  = "✔"
	IconCross       = "✖"
	IconCrossSwoosh = "✘"
	IconHome        = "⌂"
	IconNote        = "♪"
	IconWarning     = "⚠"
	IconInfo        = "ℹ"
	IconStar        = "★"
)

const (
	ShadeLight  = "░"
	ShadeMedium = "▒"
	ShadeHeavy  = "▓"
)

var DotSpinner []string = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
var CircleSpinner []string = []string{"◐", "◓", "◑", "◒"}
