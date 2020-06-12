package main

import (
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// ListItem holds information about the contents of the main list box.
type ListItem struct {
	// ID is a unique identifier that will be returned from the RunInteractive method.
	ID string
	// Label is the text shown to the user.
	Label string
	// Score is used to sort the items in the list.
	Score float32
}

// SearchFunc is implemented by the called to implement searching. The query is
// the content of the search box, and the function should return the first num
// results associated with the query, or an error.
type SearchFunc func(query string, num int) ([]ListItem, error)

type searchRequest struct {
	query      string
	numResults int
}

type interactiveState struct {
	app        *tview.Application
	input      *tview.InputField
	results    *tview.List
	searchFunc SearchFunc
	err        error
	// nextRequest is a channel which holds the next query that should be
	// evaluated by the searcher.
	nextRequest chan searchRequest
}

type threadData struct {
	nextRequest   chan searchRequest
	handleError   func(err error)
	handleResults func(items []ListItem)
}

// RunInteractive handles the full interactive mode lifecycle, returning the
// selected item.
func RunInteractive(searchFunc SearchFunc) (string, error) {
	is := interactiveState{
		searchFunc:  searchFunc,
		nextRequest: make(chan searchRequest, 1),
	}
	td := threadData{
		nextRequest: is.nextRequest,
		handleError: func(err error) {
			is.err = err
			is.app.Stop()
		},
		handleResults: is.setItems,
	}
	go goroutine(td, searchFunc)

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
	is.input.Box.SetBackgroundColor(-1)

	is.results = tview.NewList().
		ShowSecondaryText(false).
		SetWrapAround(false)
	is.results.Box.SetBackgroundColor(-1)
	resultsViewInput := is.results.InputHandler()

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(is.results, 0, 1, false).
		AddItem(is.input, 1, 0, true)

	is.input.SetDoneFunc(func(key tcell.Key) {
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
		}
		return event
	})
	is.input.SetChangedFunc(func(query string) {
		is.update(query)
	})

	is.app.SetRoot(flex, true).SetFocus(flex)
	if err := is.app.Run(); err != nil {
		is.err = err
	}
	close(is.nextRequest)
	if is.err != nil {
		return "", is.err
	}
	item := is.results.GetCurrentItem()
	text, _ := is.results.GetItemText(item)
	return text, nil
}

func (is *interactiveState) update(query string) {
	select {
	case <-is.nextRequest:
		// drop the old unstarted request
	default:
	}
	_, _, _, numResults := is.results.GetInnerRect()
	is.nextRequest <- searchRequest{query: query, numResults: numResults}
}

func (is *interactiveState) setItems(items []ListItem) {
	is.app.QueueUpdateDraw(func() {
		pos := is.results.GetCurrentItem()
		is.results.Clear()
		for _, i := range items {
			is.results.AddItem(tview.Escape(i.Label), "", 0, nil)
		}
		is.results.SetCurrentItem(pos)
	})
}

func goroutine(td threadData, searchFunc SearchFunc) {
	for {
		request, more := <-td.nextRequest
		if !more {
			break
		}
		items, err := searchFunc(request.query, request.numResults)
		if err != nil {
			td.handleError(err)
		} else {
			td.handleResults(items)
		}
	}
}
