package windows

import (
	"fmt"
	"strings"
	"utils"
	"tui"
)

type Suggester = func(string)([]string)

type InputField struct {
	X, Y, Length int
	IsModal bool
	EnterToConfirm bool
	Offset int
	Cursor int
	IsActive bool
	Value []rune
	Limit int
	Charset string
	Suggester Suggester
	Suggestion *string
	Manager *WindowManager
	Parent tui.Drawable
	OnResult func(*InputField, InputFieldResult)
}

type InputFieldResult int

const (
	FOCUS_FORWARD InputFieldResult = 1
	FOCUS_BACKWARD InputFieldResult = -1
	UPDATE InputFieldResult = 0
	CONFIRM InputFieldResult = 2
	CANCEL InputFieldResult = 3
)

func (field *InputField) Draw() {
	window := field.Parent
	runes := field.Value
	start, end := 0, 0

	// Draw current value.
	if len(runes) > 0 {
		window.WithColor(tui.COLOR_4BIT_BLACK, tui.COLOR_4BIT_CYAN, func() {
			start, end = utils.RuneSlice(field.Value, field.Offset, field.Length)
			window.MovePrint(field.Y, field.X, string(runes[start:end]))
		})
	}

	// Draw suggestion.
	visibleCells := tui.CellLength(runes[start:end])
	if field.IsActive && field.Length > visibleCells && field.Suggestion != nil {
		suggestionRunes := []rune(*field.Suggestion)
		tailStart, tailEnd := utils.RuneSlice(
			suggestionRunes,
			len(field.Value),
			field.Length - visibleCells)
		tail := suggestionRunes[tailStart:tailEnd]
		window.WithColor(tui.COLOR_4BIT_WHITE, tui.COLOR_4BIT_CYAN, func() {
			window.MovePrint(
				field.Y,
				field.X + visibleCells,
				string(suggestionRunes[tailStart:tailEnd]))
		})
		visibleCells += tui.CellLength(tail)
	}

	// Clear the rest of the field.
	if field.Length > visibleCells {
		window.WithColor(tui.COLOR_4BIT_BLACK, tui.COLOR_4BIT_CYAN, func() {
			window.MovePrint(
				field.Y,
				field.X + visibleCells,
				strings.Repeat(" ", field.Length - visibleCells))
		})
	}
}

func (field *InputField) OnInput(c tui.Key) {
	// Focus helper.
	moveFocus := func(result InputFieldResult) {
		field.IsActive = false
		field.OnResult(field, result)
	}

	if c.Rune != nil {
		field.NewChar(*c.Rune)
		field.OnResult(field, UPDATE)
	} else if c.EscapeSeq != nil {
		switch *c.EscapeSeq {
		case tui.ESC_DOWN:
			if field.IsModal { return }
			moveFocus(FOCUS_FORWARD)
		case tui.ESC_UP:
			if field.IsModal { return }
			moveFocus(FOCUS_BACKWARD)
		case tui.ESC_LEFT:
			field.Cursor = utils.MaxInt(0, field.Cursor - 1)
			if field.Offset > field.Cursor {
				field.Offset -= 1
			}
			field.OnResult(field, UPDATE)
		case tui.ESC_RIGHT:
			if field.Cursor == len(field.Value) && field.Suggestion != nil {
				field.ConfirmSuggestion()
			} else {
				field.Cursor = utils.MinInt(len(field.Value), field.Cursor + 1)
				if (field.Offset + field.Length) < field.Cursor {
					field.Offset += 1
				}
				field.OnResult(field, UPDATE)
			}
		case tui.ESC_HOME, tui.ESC_PGUP:
			field.Offset = 0
			field.Cursor = 0
			field.OnResult(field, UPDATE)
		case tui.ESC_END, tui.ESC_PGDOWN:
			field.Offset = utils.OffsetToFit(field.Value, field.Length - 1) + 1
			field.Cursor = len(field.Value)
			field.OnResult(field, UPDATE)
		}
	} else if c.ControlCode != 0 {
		switch c.ControlCode {
		case tui.ASC_ESC:
			if field.IsModal {
				field.OnResult(field, CANCEL)
			}
		case tui.ASC_TAB:
			var suggestions []string

			// Get suggestions for current value.
			if (field.Suggester != nil) && len(field.Value) > 0 {
				suggestions = field.Suggester(string(field.Value));
			} else {
				suggestions = []string{}
			}

			if len(suggestions) > 0 {
				if field.Suggestion != nil {
					if len(suggestions) > 1 {
						// Check if all suggestions start with common prefix.
						prefix := commonPrefix(suggestions)
						prefixStr := string(prefix)
						if prefixStr != "" && prefixStr != string(field.Value) {
							field.Suggestion = &suggestions[0]
							field.Value = prefix
							field.Cursor = len(field.Value)
						} else {
							// Cycle suggestion.
							suggestIndex := (utils.IndexOf(suggestions, *field.Suggestion) + 1) % len(suggestions)
							field.Suggestion = &suggestions[suggestIndex];
						}
						field.Offset = utils.MinInt(
							utils.MaxInt(0, len(*field.Suggestion) - field.Length + 1),
							len(field.Value),
						)
						field.OnResult(field, UPDATE)
					} else if suggestions[0] == string(field.Value) {
						moveFocus(FOCUS_FORWARD)
					} else {
						field.ConfirmSuggestion()
					}
				} else {
					// Nothing to suggest, just move focus.
					moveFocus(FOCUS_FORWARD)
				}
			} else {
				// Nothing to suggest, just move focus.
				moveFocus(FOCUS_FORWARD)
			}
		case tui.ASC_BACKSPACE, tui.ASC_DELETE:
			field.ControlCode(c.ControlCode)
			field.OnResult(field, UPDATE)
		case tui.ASC_ENTER:
			if field.EnterToConfirm {
				field.IsActive = false
				if field.IsModal {
					field.OnResult(field, CONFIRM)
				} else {
					field.OnResult(field, FOCUS_FORWARD)
				}
			} else if field.Suggestion != nil {
				field.ConfirmSuggestion()
			} else {
				field.IsActive = false
				if field.IsModal {
					field.OnResult(field, CONFIRM)
				} else {
					field.OnResult(field, FOCUS_FORWARD)
				}
			}
		}
	}
}

func (field *InputField) NewChar(c rune) {
	var trimmed []rune
	var runes []rune = field.Value

	if (field.Limit > 0) && (len(field.Value) >= field.Limit) {
		return
	}

	if len(field.Charset) > 0 && !strings.Contains(field.Charset, fmt.Sprintf("%c", c)) {
		return
	}

	prefix, suffix := runes[:field.Cursor], runes[field.Cursor:]
	output := fmt.Sprintf("%s%c%s", string(prefix), c, string(suffix))

	trimmed = []rune(output)
	field.Cursor += 1

	field.Value = trimmed
	field.Cursor = utils.MaxInt(utils.MinInt(len(trimmed), field.Cursor), 0)
	if field.Cursor == len(trimmed) {
		field.Offset = utils.OffsetToFit(field.Value, field.Length - 1) + 1
	}

	field.UpdateSuggestion()
}

func (field *InputField) ControlCode(code int) {
	var trimmed []rune
	var runes []rune = field.Value
	var index int

	if len(field.Value) == 0 {
		return
	}

	switch code {
	case tui.ASC_BACKSPACE:
		if field.Cursor == 0 {
			return
		} else {
			index = field.Cursor - 1
		}
		trimmed = utils.Remove(runes, index)
		field.Cursor -= 1
	case tui.ASC_DELETE:
		if field.Cursor == len(runes) {
			index = field.Cursor - 1
		} else {
			index = field.Cursor
		}
		trimmed = utils.Remove(runes, index)
	}

	field.Value = trimmed
	field.Cursor = utils.MaxInt(utils.MinInt(len(trimmed), field.Cursor), 0)
	if field.Cursor == len(trimmed) {
		field.Offset = utils.OffsetToFit(field.Value, field.Length - 1) + 1
	}

	field.UpdateSuggestion()
}

func (field *InputField) UpdateSuggestion() {
	var suggestion *string
	if (field.Suggester != nil) && len(field.Value) > 0 {
		if suggestions := field.Suggester(string(field.Value)); len(suggestions) > 0 {
			suggestion = &suggestions[0]
		}
	}
	field.Suggestion = suggestion
}

func (field *InputField) SetCursor(window tui.Drawable) {
	window.MoveTo(
		field.Y,
		field.X + utils.CellsTo(field.Value, field.Offset, field.Cursor))
}

func (field *InputField) ConfirmSuggestion() {
	field.Value = []rune(*field.Suggestion)
	field.Offset = utils.OffsetToFit(field.Value, field.Length - 1) + 1
	field.Cursor = len(field.Value)
	field.UpdateSuggestion()
	field.OnResult(field, UPDATE)
}

func commonPrefix(list []string) []rune {
	if len(list) < 2 {
		return nil
	}

	runeList := make([][]rune, len(list))
	for i, _ := range list {
		runeList[i] = []rune(list[i])
	}

	for idx, c := range runeList[0] {
		for _, s := range runeList[1:] {
			if s[idx] != c {
				return runeList[0][0:idx]
			}
		}
	}

	return nil
}
