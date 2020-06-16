package pilikino

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update .golden files")

func TestParseNoteContent(t *testing.T) {
	files, err := filepath.Glob("testdata/*.md")
	require.NoError(t, err)
	for _, caseName := range files {
		t.Run(caseName, func(t *testing.T) {
			bytes, err := ioutil.ReadFile(caseName)
			require.NoError(t, err)

			note := noteData{Content: string(bytes)}
			parseNoteContent(&note)
			note.Content = ""
			actual, err := json.Marshal(note)
			require.NoError(t, err)

			golden := caseName + ".golden"
			if *update {
				err = ioutil.WriteFile(golden, actual, 0644)
				require.NoError(t, err)
			}
			expected, err := ioutil.ReadFile(golden)
			require.NoError(t, err)

			var pActual, pExpected map[string]interface{}
			require.NoError(t, json.Unmarshal(actual, &pActual))
			require.NoError(t, json.Unmarshal(expected, &pExpected))
			require.Equal(t, pExpected, pActual)
		})
	}
}
