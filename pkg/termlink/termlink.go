/*
This file and this file only is subject to the following license.

MIT License

Copyright (c) [year] [fullname]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

Provided by STRV.
*/
package termlink

import (
	"fmt"
	"os"
	"strings"
)

func parseVersion(version string) (int, int, int) {
	var major, minor, patch int
	fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
	return major, minor, patch
}

func supportsHyperlinks() bool {
	if os.Getenv("FORCE_HYPERLINK") != "" {
		return true
	}

	if os.Getenv("DOMTERM") != "" {
		// DomTerm
		return true
	}

	if os.Getenv("VTE_VERSION") != "" {
		// VTE-based terminals above v0.50 (Gnome Terminal, Guake, ROXTerm, etc)
		major, minor, patch := parseVersion(os.Getenv("VTE_VERSION"))
		if major >= 5000 && minor >= 50 && patch >= 50 {
			return true
		}
	}

	if os.Getenv("TERM_PROGRAM") != "" {
		if os.Getenv("TERM_PROGRAM") == "Hyper" ||
			os.Getenv("TERM_PROGRAM") == "iTerm.app" ||
			os.Getenv("TERM_PROGRAM") == "terminology" ||
			os.Getenv("TERM_PROGRAM") == "WezTerm" {
			return true
		}
	}

	if os.Getenv("TERM") != "" {
		// Kitty
		if os.Getenv("TERM") == "xterm-kitty" {
			return true
		}
	}

	// Windows Terminal and Konsole
	if os.Getenv("WT_SESSION") != "" || os.Getenv("KONSOLE_VERSION") != "" {
		return true
	}

	return false
}

var colorsList = map[string]string{
	"black":     "30",
	"red":       "31",
	"green":     "32",
	"yellow":    "33",
	"blue":      "34",
	"magenta":   "35",
	"cyan":      "36",
	"white":     "37",
	"bold":      "1",
	"italic":    "3",
	"bgBlack":   "40",
	"bgRed":     "41",
	"bgGreen":   "42",
	"bgYellow":  "43",
	"bgBlue":    "44",
	"bgMagenta": "45",
	"bgCyan":    "46",
	"bgWhite":   "47",
}

func isInList(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func parseColor(color string) string {
	acceptedForegroundColors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	acceptedBackgroundColors := []string{"bgBlack", "bgRed", "bgGreen", "bgYellow", "bgBlue", "bgMagenta", "bgCyan", "bgWhite"}

	if color == "" {
		return ""
	}

	var colors []string
	for _, c := range strings.Split(color, " ") {
		if c == "" {
			continue
		}
		if c == "bold" {
			colors = append(colors, colorsList["bold"])
		} else if c == "italic" {
			colors = append(colors, colorsList["italic"])
		} else if c == "underline" {
			colors = append(colors, colorsList["underline"])
		} else if c == "blink" {
			colors = append(colors, colorsList["blink"])
		} else if c == "reverse" {
			colors = append(colors, colorsList["reverse"])
		} else if c == "hidden" {
			colors = append(colors, colorsList["hidden"])
		} else if c == "strike" {
			colors = append(colors, colorsList["strike"])
		} else if isInList(acceptedForegroundColors, c) {
			colors = append(colors, colorsList[c])
		} else if isInList(acceptedBackgroundColors, c) {
			colors = append(colors, colorsList[c])
		} else if c == "reset" {
			colors = append(colors, colorsList["reset"])
		} else {
			return ""
		}
	}

	if len(colors) == 0 {
		return ""
	}

	return "\u001b[" + strings.Join(colors, ";") + "m"
}

// Function Link creates a clickable link in the terminal's stdout.
//
// The function takes two parameters: text and url.
//
// The text parameter is the text to be displayed.
// The url parameter is the URL to be opened when the link is clicked.
//
// The function returns the clickable link.
func Link(text string, url string) string {
	if supportsHyperlinks() {
		return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07" + parseColor("reset")
	} else {
		return text + " (\u200B" + url + ")" + parseColor("reset")
	}
}

// Function LinkColor creates a colored clickable link in the terminal's stdout.
//
// The function takes three parameters: text, url and color.
//
// The text parameter is the text to be displayed.
// The url parameter is the URL to be opened when the link is clicked.
// The color parameter is the color of the link.
//
// The function returns the clickable link.
func ColorLink(text string, url string, color string) string {
	textColor := parseColor(color)

	if supportsHyperlinks() {
		return "\x1b]8;;" + url + "\x07" + textColor + text + "\x1b]8;;\x07" + parseColor("reset")
	} else {
		return textColor + text + " (\u200B" + url + ")" + parseColor("reset")
	}
}

// Function SupportsHyperlinks returns true if the terminal supports hyperlinks.
//
// The function returns true if the terminal supports hyperlinks, false otherwise.
func SupportsHyperlinks() bool {
	return supportsHyperlinks()
}
