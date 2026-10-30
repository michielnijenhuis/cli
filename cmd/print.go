package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/michielnijenhuis/cli"
)

var PrintCmd = &cli.Command{
	Name:        "print",
	Description: "Print something",
	Aliases:     []string{"p"},
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:        "critical",
			Description: "Print as critical",
		},
		&cli.BoolFlag{
			Name:        "error",
			Description: "Print as error",
		},
		&cli.BoolFlag{
			Name:        "warning",
			Description: "Print as warning",
		},
		&cli.BoolFlag{
			Name:        "warn",
			Description: "Print as warning",
		},
		&cli.BoolFlag{
			Name:        "caution",
			Description: "Print as caution",
		},
		&cli.BoolFlag{
			Name:        "success",
			Description: "Print as success",
		},
		&cli.BoolFlag{
			Name:        "debug",
			Description: "Print as debug",
		},
		&cli.BoolFlag{
			Name:        "info",
			Description: "Print as info",
		},
		&cli.BoolFlag{
			Name:        "ok",
			Description: "Print as ok",
		},
		&cli.BoolFlag{
			Name:        "check",
			Description: "Print with checkmark symbol",
		},
		&cli.BoolFlag{
			Name:        "alert",
			Description: "Print with alert symbol",
		},
		&cli.StringFlag{
			Name:        "symbol",
			Description: "The symbol to print with",
		},
		&cli.StringFlag{
			Name:        "color",
			Description: "The color to print a given symbol in",
		},
		&cli.BoolFlag{
			Name:        "timestamp",
			Description: "Print with timestamp",
			Shortcuts:   []string{"t"},
		},
		&cli.StringFlag{
			Name:        "mode",
			Description: "The mode to print as",
			Shortcuts:   []string{"m"},
			Options:     []string{"critical", "error", "warning", "caution", "success", "debug", "info", "ok", "check"},
		},
	},
	Arguments: []cli.Arg{
		&cli.ArrayArg{
			Name:        "args",
			Description: "The arguments to print",
		},
	},
	Run: func(args []string, ctx *cli.CommandCtx) {
		printArgs := ctx.IO.Array("args")
		if len(printArgs) == 0 {
			return
		}

		timestamp := ""
		if ctx.IO.Bool("timestamp") {
			timestamp = time.Now().Format("15:04:05")
			printArgs[0] = fmt.Sprintf("<fg=gray>%s</> %s", timestamp, printArgs[0])
		}

		linesSlice := make([]string, 0, len(printArgs))

		for _, arg := range printArgs {
			subLines := strings.Split(arg, "\\n")
			linesSlice = append(linesSlice, subLines...)
		}

		lines := strings.Join(linesSlice, "\n")

		mode := ctx.IO.String("mode")

		if ctx.IO.Bool("critical") || mode == "critical" {
			ctx.IO.Critical(lines)
		} else if ctx.IO.Bool("error") || mode == "error" {
			ctx.IO.Error(lines)
		} else if ctx.IO.Bool("warning") || mode == "warning" {
			ctx.IO.Warning(lines)
		} else if ctx.IO.Bool("warn") || mode == "warn" {
			ctx.IO.Warning(lines)
		} else if ctx.IO.Bool("caution") || mode == "caution" {
			ctx.IO.Caution(lines)
		} else if ctx.IO.Bool("success") || mode == "success" {
			ctx.IO.Success(lines)
		} else if ctx.IO.Bool("debug") || mode == "debug" {
			ctx.IO.Debug(lines)
		} else if ctx.IO.Bool("info") || mode == "info" {
			ctx.IO.Info(lines)
		} else if ctx.IO.Bool("ok") || mode == "ok" {
			ctx.IO.Ok(lines)
		} else if ctx.IO.Bool("check") || mode == "check" {
			ctx.IO.Symbol("✓", "green", lines)
		} else if ctx.IO.Bool("alert") || mode == "alert" {
			ctx.IO.Symbol("⚠︎", "red", lines)
		} else {
			symbol := ctx.IO.String("symbol")
			color := ctx.IO.String("color")

			if symbol != "" && color != "" {
				ctx.IO.Symbol(symbol, color, lines)
			} else {
				ctx.IO.Writeln(lines)
			}
		}
	},
}
