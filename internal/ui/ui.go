package ui

import (
	"log"
	"strings"
	"tt/internal/tt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	maxDescLen = 40
	maxTagsLen = 40
)

type UI struct {
	tt  *tt.TT
	app *tview.Application

	flex  *tview.Flex
	table *tview.Table
	form  *tview.Form
}

func New(tt *tt.TT) *UI {
	ui := UI{
		tt:    tt,
		app:   tview.NewApplication(),
		table: tview.NewTable(),
		form:  tview.NewForm(),
	}

	ui.init()

	return &ui
}

func (ui *UI) init() {
	ui.initLayout()
	ui.initTable()

	ui.flex.SetInputCapture(ui.inputCapture)
}

func (ui *UI) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch {
	case event.Rune() == 'q', event.Key() == tcell.KeyEscape:
		ui.app.Stop()
		return nil
	case event.Rune() == ' ':
		if ui.table.HasFocus() {
			ui.app.SetFocus(ui.form)
		} else {
			ui.app.SetFocus(ui.table)
		}
		return nil
	default:
		return event
	}
}

func (ui *UI) initTable() {
	ui.table.SetSelectable(true, false).SetFixed(1, 0)
	ui.table.SetSeparator(tview.Borders.Vertical)

	tasks, err := ui.tt.Tasks()
	if err != nil {
		log.Printf("error: unable to read tasks: %s", err)
		return
	}

	ui.updateTasks(tasks)
}

func (ui *UI) updateTasks(tasks []tt.Task) {
	for i, v := range []string{
		t("Description"), t("Started at"), t("Stopped at"), t("Tags"),
	} {
		cell := tview.NewTableCell(v).SetAttributes(tcell.AttrBold)
		cell.NotSelectable = true
		ui.table.SetCell(0, i, cell)
	}

	var lastLineDay string
	rowID := 1

	for _, v := range tasks {
		startedAt := v.StoppedAt.Format("15:04")
		curLineDay := v.StartedAt.Format("2006-01-02")
		if lastLineDay != curLineDay {
			startedAt = v.StartedAt.Format("2006-01-02 15:04")
			lastLineDay = curLineDay
		}

		var stoppedAt string
		if !v.StoppedAt.IsZero() {
			stoppedAt = v.StoppedAt.Format("15:04")
		}

		ui.table.SetCell(rowID, 0, tview.NewTableCell(clampString(v.Description, maxDescLen)))
		ui.table.SetCell(rowID, 1, tview.NewTableCell(startedAt).SetAlign(tview.AlignRight))
		ui.table.SetCell(rowID, 2, tview.NewTableCell(stoppedAt).SetAlign(tview.AlignRight))
		ui.table.SetCell(rowID, 3, tview.NewTableCell(clampString(strings.Join(v.Tags, " "), maxTagsLen)))
		rowID++
	}
}

func clampString(v string, max int) string {
	if len(v) > max {
		return v[:max-1] + "…"
	}

	return v
}

func (ui *UI) initLayout() {
	ui.table.SetBorder(true)
	ui.form.SetBorder(true)

	ui.flex = tview.NewFlex().
		AddItem(ui.table, 0, 2, true).
		AddItem(ui.form, 0, 1, true)
}

func (ui *UI) Run() error {
	ui.app.SetRoot(ui.flex, true).SetFocus(ui.flex)

	return ui.app.Run()
}

func t(msg string) string {
	return msg // TODO, handle locale
}
