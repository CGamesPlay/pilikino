package main

import (
	"errors"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// ErrSearchAborted is returned when the user presses escape or ^C.
var ErrSearchAborted = errors.New("aborted")

// SearchResult is the interface required by the GUI to display results. Also
// see Previewable.
type SearchResult interface {
	// Label is the text shown to the user.
	Label() string
}

// Previewable is the interface required to display content in the preview
// window when a search result is selected. Optional.
type Previewable interface {
	Preview(target *tview.TextView)
}

// SearchResults is an array of SearchResult.
type SearchResults []SearchResult

// SearchFunc is implemented by the caller to implement searching. The query is
// the content of the search box, and the function should return the first num
// results associated with the query, or an error.
type SearchFunc func(query string, num int) (SearchResults, error)

type searchRequest struct {
	query      string
	numResults int
}

type interactiveState struct {
	app         *tview.Application
	input       *tview.InputField
	resultsView *tview.List
	preview     *tview.TextView
	items       SearchResults
	searchFunc  SearchFunc
	err         error
	// nextRequest is a channel which holds the next query that should be
	// evaluated by the searcher.
	nextRequest chan searchRequest
}

type searcherThread struct {
	nextRequest   chan searchRequest
	handleError   func(err error)
	handleResults func(items SearchResults)
}

// RunInteractive handles the full interactive mode lifecycle, returning the
// selected item.
//
// `searchFunc` is called from a separate goroutine. It will never be called
// concurrently with itself.
func RunInteractive(searchFunc SearchFunc, showPreview bool) (SearchResult, error) {
	is := interactiveState{
		searchFunc:  searchFunc,
		nextRequest: make(chan searchRequest, 1),
	}
	st := searcherThread{
		nextRequest: is.nextRequest,
		handleError: func(err error) {
			is.err = err
			is.app.Stop()
		},
		handleResults: is.setItems,
	}
	go runSearcher(st, searchFunc)

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
			is.update("")
		}
		isFirstDraw = false
	})

	is.input = tview.NewInputField()
	is.input.SetFieldBackgroundColor(-1)
	is.input.SetBackgroundColor(-1)

	is.resultsView = tview.NewList().
		ShowSecondaryText(false).
		SetWrapAround(false)
	is.resultsView.Box.SetBackgroundColor(-1)
	resultsViewInput := is.resultsView.InputHandler()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(is.resultsView, 0, 1, false).
		AddItem(is.input, 1, 0, true)

	if showPreview {
		is.preview = tview.NewTextView().
			SetChangedFunc(func() {
				is.app.Draw()
			})
		is.preview.
			SetBackgroundColor(-1).
			SetBorder(true).
			SetBorderColor(tcell.ColorBlack)
		flex = tview.NewFlex().
			AddItem(flex, 0, 1, true).
			AddItem(is.preview, 0, 1, false)
	}

	is.input.SetDoneFunc(func(key tcell.Key) {
		is.err = ErrSearchAborted
		is.app.Stop()
	})
	is.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch key := event.Key(); key {
		case tcell.KeyDown, tcell.KeyUp:
			if resultsViewInput != nil {
				resultsViewInput(event, func(p tview.Primitive) {
					is.app.SetFocus(p)
				})
			}
			return nil
		case tcell.KeyEnter:
			is.app.Stop()
			return nil
		}
		return event
	})
	is.input.SetChangedFunc(func(query string) {
		is.update(query)
	})

	is.resultsView.SetChangedFunc(func(idx int, _, _ string, _ rune) {
		if !showPreview {
			return
		}
		is.preview.Clear()
		result := is.items[idx]
		if previewable, ok := result.(Previewable); ok {
			previewable.Preview(is.preview)
		}
	})

	is.app.SetRoot(flex, true).SetFocus(flex)
	if err := is.app.Run(); err != nil {
		is.err = err
	}
	close(is.nextRequest)
	if is.err != nil {
		return nil, is.err
	}
	idx := is.resultsView.GetCurrentItem()
	if idx >= len(is.items) {
		return nil, nil
	}
	return is.items[idx], nil
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

func (is *interactiveState) setItems(items SearchResults) {
	is.app.QueueUpdateDraw(func() {
		is.items = items
		pos := is.resultsView.GetCurrentItem()
		is.resultsView.Clear()
		for _, i := range items {
			is.resultsView.AddItem(tview.Escape(i.Label()), "", 0, nil)
		}
		is.resultsView.SetCurrentItem(pos)
	})
}

func runSearcher(st searcherThread, searchFunc SearchFunc) {
	for {
		request, more := <-st.nextRequest
		if !more {
			break
		}
		items, err := searchFunc(request.query, request.numResults)
		if err != nil {
			st.handleError(err)
		} else {
			st.handleResults(items)
		}
	}
}
