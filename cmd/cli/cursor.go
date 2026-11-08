package main

import (
	"github.com/michielnijenhuis/cli"
)

var CursorCmd = &cli.Command{
	Name:        "cursor",
	Description: "Command that makes managing the cursor easier",
	Commands: []*cli.Command{
		{
			Name:        "clear-line",
			Description: "Clear the current line",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.ClearLine()
			},
		},
		{
			Name:        "clear-line-after",
			Description: "Clear the line after the cursor",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.ClearLineAfter()
			},
		},
		{
			Name:        "clear-output",
			Description: "Clear the output",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.ClearOutput()
			},
		},
		{
			Name:        "clear-screen",
			Description: "Clear the screen",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.ClearScreen()
			},
		},
		{
			Name:        "current-position",
			Description: "Get the current cursor position",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				row, column := c.CurrentPosition()
				ctx.IO.Writelnf("%d %d", row, column)
			},
		},
		{
			Name:        "hide",
			Description: "Hide the cursor",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.Hide()
			},
		},
		{
			Name:        "show",
			Description: "Show the cursor",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.Show()
			},
		},
		{
			Name:        "move",
			Description: "Move the cursor to a specific position",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "row",
					Description: "Row to move to",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
				&cli.StringArg{
					Name:        "column",
					Description: "Column to move to",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				row := ctx.IO.Int("row")
				column := ctx.IO.Int("column")
				c.Move(column, row)
			},
		},
		{
			Name:        "move-down",
			Description: "Move the cursor down by a specific number of rows",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "rows",
					Description: "Amount of rows to move down",
					Value:       "1",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				rows := ctx.IO.Int("rows")
				c.MoveDown(rows)
			},
		},
		{
			Name:        "move-up",
			Description: "Move the cursor up by a specific number of rows",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "rows",
					Description: "Amount of rows to move up",
					Value:       "1",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				rows := ctx.IO.Int("rows")
				c.MoveUp(rows)
			},
		},
		{
			Name:        "move-left",
			Description: "Move the cursor left by a specific number of columns",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "columns",
					Description: "Amount of columns to move left",
					Value:       "1",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				columns := ctx.IO.Int("columns")
				c.MoveLeft(columns)
			},
		},
		{
			Name:        "move-right",
			Description: "Move the cursor right by a specific number of columns",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "columns",
					Description: "Amount of columns to move right",
					Value:       "1",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				columns := ctx.IO.Int("columns")
				c.MoveRight(columns)
			},
		},
		{
			Name:        "move-to-column",
			Description: "Move the cursor to a specific column",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "column",
					Description: "Column to move to",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				column := ctx.IO.Int("column")
				c.MoveToColumn(column)
			},
		},
		{
			Name:        "move-to-position",
			Description: "Move the cursor to a specific row and column",
			Arguments: []cli.Arg{
				&cli.StringArg{
					Name:        "column",
					Description: "Column to move to",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
				&cli.StringArg{
					Name:        "row",
					Description: "Row to move to",
					Validator:   cli.IsPositiveInt,
					Required:    true,
				},
			},
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				row := ctx.IO.Int("row")
				column := ctx.IO.Int("column")
				c.MoveToPosition(column, row)
			},
		},
		{
			Name:        "restore-position",
			Description: "Restore the cursor to its previous position",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.RestorePosition()
			},
		},
		{
			Name:        "save-position",
			Description: "Save the current cursor position",
			Run: func(args []string, ctx *cli.CommandCtx) {
				c := makeCursor(ctx)
				c.SavePosition()
			},
		},
	},
}

func makeCursor(ctx *cli.CommandCtx) cli.Cursor {
	return cli.Cursor{
		Output: ctx.IO.Output,
		Input:  ctx.IO.Input.Stream,
	}
}
