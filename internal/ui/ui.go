package ui

import (
	"errors"
	"fmt"
	"time"
	"tt/internal/tt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func t(msg string) string {
	return msg // TODO, handle locale
}

const (
	maxDescLen         = 40
	maxTagsLen         = 40
	dateFormat         = "2006-01-02"
	timeFormat         = "15:04"
	dateTimeFormat     = dateFormat + " " + timeFormat
	formDateTimeFormat = dateTimeFormat + ":05"
)

/* Layout:
    mainFlex {
    mainPages [
        taskFlex {taskTable, taskForm}
        configForm
    ]
    mainFooter
}
*/

// UI holds the state for the TUI.
type UI struct {
	tt  *tt.TT
	app *tview.Application

	mainFlex   *tview.Flex
	mainPages  *tview.Pages
	mainFooter *tview.TextView

	// Task view.
	taskFlex  *tview.Flex
	taskTable *tview.Table
	taskForm  *tview.Form

	configForm *tview.Form

	tasks             []tt.Task
	selectedTaskIndex int // index in tasks slice
}

func New(tt *tt.TT) *UI {
	ui := UI{
		tt:  tt,
		app: tview.NewApplication(),

		mainPages:  tview.NewPages(),
		mainFlex:   tview.NewFlex(),
		mainFooter: tview.NewTextView(),

		taskFlex:  tview.NewFlex(),
		taskTable: tview.NewTable(),
		taskForm:  tview.NewForm(),

		configForm: tview.NewForm(),
	}

	ui.init()

	return &ui
}

func (ui *UI) Run() error {
	return ui.app.Run()
}

const (
	pageTasks  = "tasks"
	pageConfig = "config"
)

func (ui *UI) init() {
	ui.initMainLayout()
	ui.initTaskPageLayout()

	ui.mainPages.AddPage(pageTasks, ui.taskFlex, true, true)
	ui.mainPages.AddPage(pageConfig, ui.configForm, true, false)

	ui.mainPages.SetInputCapture(ui.inputCapture)
	ui.app.SetRoot(ui.mainFlex, true).SetFocus(ui.taskTable)
	ui.setActivePage(pageTasks)
}

func (ui *UI) setActivePage(name string) {
	ui.mainPages.SwitchToPage(name)
	ui.mainFooter.Highlight(name)
}

func (ui *UI) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() { // nolint:exhaustive
	case tcell.KeyF1:
		ui.setActivePage(pageTasks)
		return nil
	case tcell.KeyF2:
		ui.setActivePage(pageConfig)
		return nil
	}

	switch {
	case ui.taskForm.HasFocus():
		return ui.taskFormInputCapture(event)
	case ui.taskTable.HasFocus():
		return ui.taskTableInputCapture(event)
	}

	return nil
}

func (ui *UI) printError(msg string, args ...interface{}) {
	// TODO proper error display
	// nolint:forbidigo
	fmt.Print("\x1b[0m" + fmt.Sprintf(msg, args...))
}

var errNoSelection = errors.New("no selection")

func formDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(formDateTimeFormat)
}

func clampString(v string, max int) string {
	if len(v) > max {
		return v[:max-1] + "…"
	}

	return v
}

func (ui *UI) initMainLayout() {
	ui.mainFlex.
		SetDirection(tview.FlexRow).
		AddItem(ui.mainPages, 0, 1, true).
		AddItem(ui.mainFooter, 1, 1, false)

	ui.mainFooter.
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	pages := []struct{ name, title, key string }{
		{pageTasks, t("Tasks"), "F1"},
		{pageConfig, t("Config"), "F2"},
	}
	for _, v := range pages {
		fmt.Fprintf(ui.mainFooter, `%s ["%s"][darkcyan]%s[white][""]  `, v.key, v.name, v.title)
	}
}
