package aoc

import "github.com/fatih/color"

var (
	passLabel       = color.New(color.FgHiGreen).Sprint("pass")
	failLabel       = color.New(color.FgHiRed).Sprint("fail")
	incompleteLabel = color.New(color.BgHiYellow).Sprint("did not complete")
	missingLabel    = color.New(color.FgHiYellow, color.Italic).Sprint("empty")

	bold       = color.New(color.Bold)
	dimmed     = color.New(color.FgHiBlack, color.Italic)
	brightBlue = color.New(color.FgHiBlue)
	boldBlue   = color.New(color.Bold, color.FgHiBlue)
	boldYellow = color.New(color.Bold, color.FgHiYellow)
)
