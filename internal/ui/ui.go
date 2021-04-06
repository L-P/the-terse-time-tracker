package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"tt/internal/tt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	maxDescLen         = 40
	maxTagsLen         = 40
	dateFormat         = "2006-01-02"
	timeFormat         = "15:04"
	dateTimeFormat     = dateFormat + " " + timeFormat
	formDateTimeFormat = dateTimeFormat + ":05"
)

type UI struct {
	tt  *tt.TT
	app *tview.Application

	flex  *tview.Flex
	table *tview.Table
	form  *tview.Form

	tasks             []tt.Task
	selectedTaskIndex int // index in tasks slice
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
	ui.app.SetRoot(ui.flex, true).SetFocus(ui.flex)

	if len(ui.tasks) > 0 {
		ui.table.Select(1, 0)
	}
}

func (ui *UI) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	if ui.form.HasFocus() {
		return ui.formInputCapture(event)
	}

	return ui.tableInputCapture(event)
}

func (ui *UI) formInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() { // nolint:exhaustive
	case tcell.KeyEscape:
		ui.app.SetFocus(ui.table)
		return nil
	default:
		return event
	}
}

func (ui *UI) tableInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() { // nolint:exhaustive
	case tcell.KeyRune:
		switch event.Rune() {
		case 'q':
			ui.app.Stop()
		default:
			return event
		}
	case tcell.KeyEscape:
		ui.app.Stop()
	case tcell.KeyDelete:
		ui.deleteSelectedTask()
	case tcell.KeyF5:
		ui.updateTasksTable(ui.tasks)
	default:
		return event
	}

	return nil
}

func (ui *UI) printError(msg string, args ...interface{}) {
	// TODO proper error display
	// nolint:forbidigo
	fmt.Print("\x1b[0m" + fmt.Sprintf(msg, args...))
}

func (ui *UI) initTable() {
	ui.table.SetSelectable(true, false).SetFixed(1, 0)
	ui.table.SetSeparator(tview.Borders.Vertical)

	tasks, err := ui.tt.GetTasks()
	if err != nil {
		ui.printError("error: unable to read tasks: %s", err)
		tasks = nil
	}

	ui.table.SetSelectionChangedFunc(func(row, col int) {
		ui.selectedTaskIndex = row - 1
		selected, _ := ui.selectedTask()
		ui.updateForm(selected)
	})

	ui.table.SetSelectedFunc(func(row, col int) {
		ui.app.SetFocus(ui.form)
	})

	ui.updateTasksTable(tasks)
}

const ( // must match the AddInputField order below
	formFieldIndexDescription = iota
	formFieldIndexStartedAt
	formFieldIndexStoppedAt
	formFieldIndexTags
)

func (ui *UI) updateForm(task tt.Task) {
	ui.form.Clear(true)
	ui.form.
		AddInputField(t("Description"), task.Description, 0, nil, nil).
		AddInputField(t("Started at"), formDate(task.StartedAt), len(formDateTimeFormat), nil, nil).
		AddInputField(t("Stopped at"), formDate(task.StoppedAt), len(formDateTimeFormat), nil, nil).
		AddInputField(t("Tags"), strings.Join(task.Tags, " "), 0, nil, nil).
		AddButton(t("Save"), ui.saveFormTask).
		AddButton(t("Delete"), ui.deleteSelectedTask)
}

func (ui *UI) saveFormTask() {
	ui.app.SetFocus(ui.table)
	if len(ui.tasks) == 0 {
		return
	}

	task, err := ui.getFormTask()
	if err != nil {
		ui.printError("error: invalid task:  %s", err) // TODO proper error display
		return
	}

	if err := ui.tt.UpdateTask(task); err != nil {
		ui.printError("error: can't update: %s", err) // TODO proper error display
		return
	}

	ui.tasks[ui.selectedTaskIndex] = task
	ui.updateTasksTable(ui.tasks)
}

var errNoSelection = errors.New("no selection")

func (ui *UI) selectedTask() (tt.Task, error) {
	if len(ui.tasks) == 0 || ui.selectedTaskIndex < 0 {
		return tt.Task{}, errNoSelection
	}

	return ui.tasks[ui.selectedTaskIndex], nil
}

func (ui *UI) deleteSelectedTask() {
	ui.app.SetFocus(ui.table)
	selected, err := ui.selectedTask()
	if err != nil {
		return
	}

	if err := ui.tt.DeleteTask(selected.ID); err != nil {
		ui.printError("error: can't delete: %s", err) // TODO proper error display
		return
	}

	tasks := append(ui.tasks[:ui.selectedTaskIndex], ui.tasks[ui.selectedTaskIndex+1:]...)

	// TODO: starts to feel sluggish after 100k tasks, but RemoveRow incurs
	// additional complexity.
	ui.updateTasksTable(tasks)

	// when removing the last row the selection is lost
	if max := len(tasks) - 1; ui.selectedTaskIndex >= max {
		ui.selectedTaskIndex = max
	}

	// let the selection callback update the rest
	ui.table.Select(ui.selectedTaskIndex+1, 0)
}

func (ui *UI) getFormTask() (tt.Task, error) {
	selected, err := ui.selectedTask()
	if err != nil {
		return tt.Task{}, err
	}

	// Indices are from the input order in the form definition in updateForm.
	value := func(i int) string {
		return ui.form.GetFormItem(i).(*tview.InputField).GetText()
	}

	_, tags := tt.ParseRawDesc(value(formFieldIndexTags))

	location := time.Now().Location()
	startedAt, err := time.ParseInLocation(formDateTimeFormat, value(formFieldIndexStartedAt), location)
	if err != nil {
		return tt.Task{}, tt.ErrInvalidInput(err.Error())
	}

	var stoppedAt time.Time
	if str := value(formFieldIndexStoppedAt); str != "" {
		stoppedAt, err = time.ParseInLocation(dateTimeFormat, str, location)
		if err != nil {
			return tt.Task{}, tt.ErrInvalidInput(err.Error())
		}
	}

	if !stoppedAt.IsZero() && startedAt.After(stoppedAt) {
		startedAt, stoppedAt = stoppedAt, startedAt
	}

	return tt.Task{
		ID:          selected.ID,
		Description: value(formFieldIndexDescription),
		StartedAt:   startedAt,
		StoppedAt:   stoppedAt,
		Tags:        tags,
	}, nil
}

func formDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}

	return t.Format(formDateTimeFormat)
}

func (ui *UI) updateTasksTable(tasks []tt.Task) {
	ui.tasks = tasks

	ui.table.Clear()
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
		startedAt := v.StartedAt.Format(timeFormat)
		curLineDay := v.StartedAt.Format(dateFormat)
		if lastLineDay != curLineDay {
			startedAt = v.StartedAt.Format(dateTimeFormat)
			lastLineDay = curLineDay
		}

		var stoppedAt string
		if !v.StoppedAt.IsZero() {
			stoppedAt = v.StoppedAt.Format(timeFormat)
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
	ui.table.SetBorder(true).SetTitle(t("Past tasks"))
	ui.form.SetBorder(true).SetTitle(t("Selected task"))

	ui.flex = tview.NewFlex().
		AddItem(ui.table, 0, 2, true).
		AddItem(ui.form, 0, 1, true)
}

func (ui *UI) Run() error {
	return ui.app.Run()
}

func t(msg string) string {
	return msg // TODO, handle locale
}
