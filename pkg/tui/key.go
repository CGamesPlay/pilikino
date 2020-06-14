package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
)

// Key represents a parsed key sequence.
type Key *tcell.EventKey

// KeyEnter represents the enter key. It is the default value for ExpectedKeys.
var KeyEnter = Key(tcell.NewEventKey(tcell.KeyRune, '\r', 0))

var keyPrefixes = []struct {
	str string
	mod tcell.ModMask
}{
	{str: "shift-", mod: tcell.ModShift},
	{str: "alt-", mod: tcell.ModAlt},
	// According to tcell documentation, meta is never produced as an event,
	// alt is used instead. So we will support meta as an alias for alt.
	{str: "meta-", mod: tcell.ModAlt},
	{str: "ctrl-", mod: tcell.ModCtrl},
}

// ParseKey parses the given key name into a EventKey which would be produced
// by that key.
func ParseKey(input string) (Key, error) {
	var (
		mod tcell.ModMask
		key tcell.Key
		ch  rune
	)
	foundModifier := true
	for foundModifier {
		foundModifier = false
		for _, prefix := range keyPrefixes {
			if len(input) >= len(prefix.str) && strings.EqualFold(input[0:len(prefix.str)], prefix.str) {
				foundModifier = true
				mod |= prefix.mod
				input = input[len(prefix.str):]
			}
		}
	}
	isSpace := strings.EqualFold(input, "space")
	if len(input) == 1 || isSpace {
		key = tcell.KeyRune
		if isSpace {
			ch = ' '
		} else {
			ch = rune(input[0])
		}
		if mod&tcell.ModShift != 0 {
			mod &^= tcell.ModShift
			ch -= 32
		}
		if mod&tcell.ModCtrl != 0 {
			key = tcell.Key(ch &^ 0x60)
			ch = rune(key)
		}
	} else {
		found := false
		for testKey, testName := range tcell.KeyNames {
			if strings.EqualFold(input, testName) {
				key = testKey
				found = true
				break
			}
		}
		if !found {
			return &tcell.EventKey{}, fmt.Errorf("unknown key: %#v", input)
		}
	}
	result := tcell.NewEventKey(key, ch, mod)
	return Key(result), nil
}
