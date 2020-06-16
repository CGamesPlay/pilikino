package tui

import (
	"errors"
	"sync/atomic"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// ErrSearchAborted is returned when the user presses escape or ^C.
var ErrSearchAborted = errors.New("aborted")

// Document is the interface required by the GUI to display results. Also
// see Previewable.
type Document interface {
	// Label is the text shown to the user.
	Label() string
}

// Previewable is the interface required to display content in the preview
// window when a Document is selected. Optional.
type Previewable interface {
	Preview(target *tview.TextView)
}

// SearchResult holds data about a successful search.
type SearchResult struct {
	// Results holds the actual matching documents
	Results []Document
	// TotalCandidates represents the number of documents searched (size of the
	// index).
	TotalCandidates uint64
	// Status, if nonempty, will override what it displayed in the status line.
	Status string
	// QueryError represents a nonfatal errors in parsing the query that should
	// be displayed to the user without crashing the program.
	QueryError error
}

// SearchFunc is implemented by the caller to implement searching. The query is
// the content of the search box, and the function should return the first num
// results associated with the query, or an error. This function will be called
// from a goroutine, but will never be called concurrently with itself.
type SearchFunc func(query string, num int) (SearchResult, error)

type searchRequest struct {
	query      string
	numResults int
}

// Tui is the interface for controlling the interactive search. This package
// does not have any specific knowledge of pilikino, it just presents a generic
// live terminal search interface. The public members of this struct may be
// not be modified while an invocation of Run is in progress.
type Tui struct {
	// Set to 1 whenever Run is executing, and 0 otherwise.
	locked int32
	// Wraps a *interactiveState
	state atomic.Value
	// Wraps a string
	statusText atomic.Value
	// SearchFunc is the SearchFunc used by the next invocation of Run.
	SearchFunc SearchFunc
	// ShowPreview controls whether a preview window will be shown on the next
	// invocation of Run.
	ShowPreview bool
	// ExpectedKeys is the array of keys which will be accepted by the UI. The
	// default value of this field contains KeyEnter.
	ExpectedKeys []Key
}

// InteractiveResults holds the results of the interactive mode search.
type InteractiveResults struct {
	// Results holds all of the results selected by the user
	Results []Document
	// Action holds the index of the key in ExpectedKeys that the user pressed
	// to finalize the search.
	Action int
}

// This struct holds all values modified by an interactive mode search. It's
// just used to isolate the variables and make cleanup simple.
type interactiveState struct {
	// Pointer to Tui.statusText
	statusText   *atomic.Value
	app          *tview.Application
	input        *tview.InputField
	statusLine   *tview.TextView
	resultsView  *tview.List
	preview      *tview.TextView
	latestResult SearchResult
	err          error
	// nextRequest is a channel which holds the next query that should be
	// evaluated by the searcher.
	nextRequest chan searchRequest
}

type searcherThread struct {
	nextRequest   chan searchRequest
	handleError   func(err error)
	handleResults func(result SearchResult)
}

// NewTui creates a new interface for performing an interactive search.
func NewTui(searchFunc SearchFunc, showPreview bool) *Tui {
	res := &Tui{
		SearchFunc:   searchFunc,
		ShowPreview:  showPreview,
		ExpectedKeys: []Key{KeyEnter},
	}
	res.statusText.Store("")
	return res
}

// Stop cancels an in-progress interactive search, causing Run to return.
func (tui *Tui) Stop() {
	val := tui.state.Load()
	if val != nil {
		val.(*interactiveState).app.Stop()
	}
}

// Refresh causes an in-progress search to immediately refresh the currently
// displayed results. Use if the underlying results may have changed and this
// should be reflected in the UI.
func (tui *Tui) Refresh() {
	val := tui.state.Load()
	if val != nil {
		is := val.(*interactiveState)
		is.app.QueueUpdateDraw(func() {
			is.update(is.input.GetText())
		})
	}
}

// SetStatusText sets the value of the status line. A query error will always
// take precedence over the set value.
func (tui *Tui) SetStatusText(value string) {
	tui.statusText.Store(value)
}

// Run performs an interactive search, returning the selected item.
func (tui *Tui) Run() (*InteractiveResults, error) {
	if ok := atomic.CompareAndSwapInt32(&tui.locked, 0, 1); !ok {
		return nil, errors.New("attempt to perform interactive search concurrently")
	}
	defer func() {
		atomic.StoreInt32(&tui.locked, 0)
	}()
	is := &interactiveState{
		statusText:  &tui.statusText,
		nextRequest: make(chan searchRequest, 1),
	}
	st := searcherThread{
		nextRequest: is.nextRequest,
		handleError: func(err error) {
			is.err = err
			is.app.Stop()
		},
		handleResults: is.setResult,
	}
	go runSearcher(st, tui.SearchFunc)

	var initialQuery string
	var results = InteractiveResults{Action: -1}

	is.app = tview.NewApplication()
	is.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		// XXX this may cause flicker but it's the only way to reset the
		// background of a cell to transparent
		screen.Clear()
		return false
	})
	isFirstDraw := true
	is.app.SetAfterDrawFunc(func(screen tcell.Screen) {
		if isFirstDraw {
			// We have to do this after the first draw of the application to
			// know how many screen lines are available for results.
			is.update(initialQuery)
		}
		isFirstDraw = false
	})

	is.input = tview.NewInputField().
		SetLabel("> ").
		SetText(initialQuery).
		SetLabelColor(tcell.Color110).
		SetFieldBackgroundColor(tcell.ColorDefault).
		SetFieldTextColor(tcell.ColorDefault)
	is.input.SetBackgroundColor(tcell.ColorDefault)

	is.statusLine = tview.NewTextView().
		SetTextColor(tcell.ColorDefault)
	is.statusLine.SetBackgroundColor(tcell.ColorDefault)

	is.resultsView = tview.NewList().
		ShowSecondaryText(false).
		SetWrapAround(false).
		SetMainTextColor(tcell.ColorDefault).
		SetSelectedTextColor(tcell.Color254).
		SetSelectedBackgroundColor(tcell.Color236)
	is.resultsView.SetBackgroundColor(tcell.ColorDefault)
	resultsViewInput := is.resultsView.InputHandler()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(is.input, 1, 0, true).
		AddItem(is.statusLine, 1, 0, false).
		AddItem(is.resultsView, 0, 1, false)

	if tui.ShowPreview {
		is.preview = tview.NewTextView().
			SetTextColor(tcell.ColorDefault).
			SetChangedFunc(func() {
				is.app.Draw()
			})
		is.preview.
			SetBackgroundColor(tcell.ColorDefault).
			SetBorder(true).
			SetBorderColor(tcell.ColorDefault)
		flex = tview.NewFlex().
			AddItem(flex, 0, 1, true).
			AddItem(is.preview, 0, 1, false)
	}

	is.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		for i, exKey := range tui.ExpectedKeys {
			ex := (*tcell.EventKey)(exKey)
			if ex.Key() == event.Key() && ex.Modifiers() == event.Modifiers() && ex.Rune() == event.Rune() {
				results.Action = i
				is.app.Stop()
				return nil
			}
		}
		switch key := event.Key(); key {
		case tcell.KeyDown, tcell.KeyUp:
			if resultsViewInput != nil {
				resultsViewInput(event, func(p tview.Primitive) {
					is.app.SetFocus(p)
				})
			}
			return nil
		case tcell.KeyEscape:
			is.app.Stop()
			return nil
		}
		return event
	})
	is.input.SetChangedFunc(func(query string) {
		is.update(query)
	})

	is.resultsView.SetChangedFunc(func(idx int, _, _ string, _ rune) {
		if !tui.ShowPreview {
			return
		}
		is.preview.Clear()
		result := is.latestResult.Results[idx]
		if previewable, ok := result.(Previewable); ok {
			previewable.Preview(is.preview)
		}
	})

	is.app.SetRoot(flex, true).SetFocus(flex)

	tui.state.Store(is)
	defer func() {
		tui.state.Store((*interactiveState)(nil))
	}()

	if err := is.app.Run(); err != nil {
		is.err = err
	}
	close(is.nextRequest)
	if is.err != nil {
		return nil, is.err
	} else if results.Action == -1 {
		return nil, ErrSearchAborted
	}
	idx := is.resultsView.GetCurrentItem()
	if idx < len(is.latestResult.Results) {
		results.Results = []Document{is.latestResult.Results[idx]}
	}
	return &results, nil
}

func (is *interactiveState) update(query string) {
	select {
	case <-is.nextRequest:
		// drop the old unstarted request
	default:
	}
	_, _, _, numResults := is.resultsView.GetInnerRect()
	is.nextRequest <- searchRequest{query: query, numResults: numResults}
}

func (is *interactiveState) setResult(result SearchResult) {
	is.app.QueueUpdateDraw(func() {
		if result.QueryError != nil {
			is.latestResult.QueryError = result.QueryError
		} else {
			is.latestResult = result
			pos := is.resultsView.GetCurrentItem()
			is.resultsView.Clear()
			for _, i := range result.Results {
				is.resultsView.AddItem(tview.Escape(i.Label()), "", 0, nil)
			}
			is.resultsView.SetCurrentItem(pos)
		}
		is.updateStatusLine()
	})
}

func (is *interactiveState) updateStatusLine() {
	if is.latestResult.QueryError != nil {
		is.statusLine.SetText(is.latestResult.QueryError.Error())
	} else {
		is.statusLine.SetText(is.statusText.Load().(string))
	}
}

func runSearcher(st searcherThread, searchFunc SearchFunc) {
	for {
		request, more := <-st.nextRequest
		if !more {
			break
		}
		result, err := searchFunc(request.query, request.numResults)
		if err != nil {
			st.handleError(err)
		} else {
			st.handleResults(result)
		}
	}
}
