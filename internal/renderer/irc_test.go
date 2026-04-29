package renderer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIRC_Bold(t *testing.T) {
	input := "normal\x02bold\x02normal"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-bold`)
	assert.Contains(t, result, "bold")
}

func TestIRC_Italic(t *testing.T) {
	input := "normal\x1ditalic\x1dnormal"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-italic`)
}

func TestIRC_Underline(t *testing.T) {
	input := "normal\x1funderline\x1fnormal"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-underline`)
}

func TestIRC_Strikethrough(t *testing.T) {
	input := "normal\x1estrike\x1enormal"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-strikethrough`)
}

func TestIRC_Monospace(t *testing.T) {
	input := "normal\x11mono\x11normal"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-monospace`)
}

func TestIRC_ColorFg(t *testing.T) {
	input := "normal\x034red text"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-fg-4`)
}

func TestIRC_ColorFgBg(t *testing.T) {
	input := "\x034,12red on blue"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-fg-4`)
	assert.Contains(t, result, `irc-bg-12`)
}

func TestIRC_ColorReset(t *testing.T) {
	input := "\x034red\x03 plain"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-fg-4`)
}

func TestIRC_ColorTwoDigit(t *testing.T) {
	input := "\x0312light blue"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-fg-12`)
}

func TestIRC_Reverse(t *testing.T) {
	input := "\x034,12\x16reversed"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-fg-12`)
	assert.Contains(t, result, `irc-bg-4`)
}

func TestIRC_Reset(t *testing.T) {
	input := "\x02\x034bold red\x0f plain"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "irc-bold")
	assert.Contains(t, result, "irc-fg-4")
	body := result[strings.Index(result, "</span>")+len("</span>"):]
	assert.NotContains(t, body, "irc-bold")
}

func TestIRC_Spoiler(t *testing.T) {
	input := "\x034,4hidden text"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, `irc-spoiler`)
}

func TestIRC_CodeSpanUntouched(t *testing.T) {
	input := "normal `\x02bold\x02 code` normal"
	result, err := Render(input)
	assert.NoError(t, err)
	codeStart := strings.Index(result, "<code>")
	codeEnd := strings.Index(result, "</code>")
	codeContent := result[codeStart+len("<code>"):codeEnd]
	assert.NotContains(t, codeContent, "irc-bold")
	assert.Contains(t, codeContent, "\x02")
}

func TestIRC_FencedCodeUntouched(t *testing.T) {
	input := "```\n\x02bold\x02 in code\n```"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.NotContains(t, result, "irc-bold")
	assert.Contains(t, result, "\x02")
}

func TestIRC_MultipleFormats(t *testing.T) {
	input := "\x02\x034bold red\x0f plain"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "irc-bold")
	assert.Contains(t, result, "irc-fg-4")
}

func TestIRC_HexColorStripped(t *testing.T) {
	input := "\x04FF0000red text"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.NotContains(t, result, "FF0000")
	assert.Contains(t, result, "red text")
}

func TestIRC_PlainTextPassthrough(t *testing.T) {
	input := "just normal text"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.NotContains(t, result, "irc-")
}

func TestIRC_Color99(t *testing.T) {
	input := "\x034colored\x0399 default"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "irc-fg-4")
}

func TestIRC_ExtendedColor16(t *testing.T) {
	input := "\x0316extended"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "irc-fg-16")
}

func TestIRC_ExtendedColorBg(t *testing.T) {
	input := "\x034,16red on ext"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.Contains(t, result, "irc-fg-4")
	assert.Contains(t, result, "irc-bg-16")
}

func TestParseIRCSegments_BoldToggle(t *testing.T) {
	data := []byte("a\x02b\x02c")
	segs := parseIRCSegments(data)
	assert.Equal(t, 3, len(segs))
	assert.Equal(t, "a", string(segs[0].text))
	assert.False(t, segs[0].state.bold)
	assert.Equal(t, "b", string(segs[1].text))
	assert.True(t, segs[1].state.bold)
	assert.Equal(t, "c", string(segs[2].text))
	assert.False(t, segs[2].state.bold)
}

func TestParseIRCSegments_ColorParsing(t *testing.T) {
	data := []byte("\x034,12hello")
	segs := parseIRCSegments(data)
	assert.Equal(t, 1, len(segs))
	assert.Equal(t, "hello", string(segs[0].text))
	assert.Equal(t, 4, segs[0].state.fg)
	assert.Equal(t, 12, segs[0].state.bg)
}

func TestIRC_ControlCharsStripped(t *testing.T) {
	input := "text\x01\x07\x1bmore"
	result, err := Render(input)
	assert.NoError(t, err)
	assert.NotContains(t, result, "\x01")
	assert.NotContains(t, result, "\x07")
	assert.NotContains(t, result, "\x1b")
	assert.Contains(t, result, "text")
	assert.Contains(t, result, "more")
}
