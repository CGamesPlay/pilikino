package tui

import (
	"testing"

	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKey(t *testing.T) {
	cases := []struct {
		in  string
		key *tcell.EventKey
	}{
		{in: "space", key: tcell.NewEventKey(tcell.KeyRune, ' ', 0)},
		{in: "a", key: tcell.NewEventKey(tcell.KeyRune, 'a', 0)},
		{in: "f1", key: tcell.NewEventKey(tcell.KeyF1, 0, 0)},
		{in: "shift-a", key: tcell.NewEventKey(tcell.KeyRune, 'A', 0)},
		{in: "alt-t", key: tcell.NewEventKey(tcell.KeyRune, 't', tcell.ModAlt)},
		{in: "ctrl-v", key: tcell.NewEventKey(tcell.KeyCtrlV, 22, tcell.ModCtrl)},
		{in: "ctrl-\\", key: tcell.NewEventKey(tcell.KeyCtrlBackslash, 28, tcell.ModCtrl)},
		{in: "ctrl-]", key: tcell.NewEventKey(tcell.KeyCtrlRightSq, 29, tcell.ModCtrl)},
		{in: "ctrl-^", key: tcell.NewEventKey(tcell.KeyCtrlCarat, 30, tcell.ModCtrl)},
		{in: "ctrl-_", key: tcell.NewEventKey(tcell.KeyCtrlUnderscore, 31, tcell.ModCtrl)},
		{in: "ctrl-space", key: tcell.NewEventKey(tcell.KeyCtrlSpace, 0, tcell.ModCtrl)},
		{in: "right", key: tcell.NewEventKey(tcell.KeyRight, 0, 0)},
		{in: "shift-right", key: tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift)},
		{in: "alt-shift-right", key: tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModShift|tcell.ModAlt)},
		{in: "ctrl-alt-space", key: tcell.NewEventKey(tcell.KeyCtrlSpace, 0, tcell.ModCtrl|tcell.ModAlt)},
	}
	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			res, err := ParseKey(c.in)
			require.NoError(t, err)
			ev := (*tcell.EventKey)(res)
			assert.Equal(t, c.key.Modifiers(), ev.Modifiers(), "modifiers should be equal")
			assert.Equal(t, c.key.Key(), ev.Key(), "key should be equal")
			assert.Equal(t, c.key.Rune(), ev.Rune(), "rune should be equal")
		})
	}
}
