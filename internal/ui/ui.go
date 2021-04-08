package ui

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"tt/internal/tt"
	"tt/internal/util"

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

func (ui *UI) taskFormInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() { // nolint:exhaustive
	case tcell.KeyEscape:
		ui.app.SetFocus(ui.taskTable)
		return nil
	default:
		return event
	}
}

func (ui *UI) taskTableInputCapture(event *tcell.EventKey) *tcell.EventKey {
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

func (ui *UI) initTaskTable() {
	ui.taskTable.SetBorder(true).SetTitle(t("Past tasks"))
	ui.taskTable.SetSelectable(true, false).SetFixed(1, 0)
	ui.taskTable.SetSeparator(tview.Borders.Vertical)

	tasks, err := ui.tt.GetTasks()
	if err != nil {
		ui.printError("error: unable to read tasks: %s", err)
		tasks = nil
	}

	ui.taskTable.SetSelectionChangedFunc(func(row, col int) {
		ui.selectedTaskIndex = row - 1
		selected, _ := ui.selectedTask()
		ui.updateForm(selected)
	})

	ui.taskTable.SetSelectedFunc(func(row, col int) {
		ui.app.SetFocus(ui.taskForm)
	})

	ui.updateTasksTable(tasks)

	if len(tasks) > 0 {
		ui.taskTable.Select(1, 0)
	}
}

const ( // must match the AddInputField order below
	formFieldIndexDescription = iota
	formFieldIndexStartedAt
	formFieldIndexStoppedAt
	formFieldIndexTags
)

func (ui *UI) updateForm(task tt.Task) {
	ui.taskForm.Clear(true)
	ui.taskForm.
		AddInputField(t("Description"), task.Description, 0, nil, nil).
		AddInputField(t("Started at"), formDate(task.StartedAt), len(formDateTimeFormat), nil, nil).
		AddInputField(t("Stopped at"), formDate(task.StoppedAt), len(formDateTimeFormat), nil, nil).
		AddInputField(t("Tags"), strings.Join(task.Tags, " "), 0, nil, nil).
		AddButton(t("Save"), ui.saveFormTask).
		AddButton(t("Delete"), ui.deleteSelectedTask)
}

func (ui *UI) saveFormTask() {
	ui.app.SetFocus(ui.taskTable)
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
	ui.app.SetFocus(ui.taskTable)
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
	ui.taskTable.Select(ui.selectedTaskIndex+1, 0)
}

func (ui *UI) getFormTask() (tt.Task, error) {
	selected, err := ui.selectedTask()
	if err != nil {
		return tt.Task{}, err
	}

	// Indices are from the input order in the form definition in updateForm.
	value := func(i int) string {
		return ui.taskForm.GetFormItem(i).(*tview.InputField).GetText()
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

	ui.taskTable.Clear()
	for i, v := range []string{
		t("Description"), t("Started at"), t("Stopped at"), t("Duration"), t("Tags"),
	} {
		cell := tview.NewTableCell(v).SetAttributes(tcell.AttrBold)
		cell.NotSelectable = true
		ui.taskTable.SetCell(0, i, cell)
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

		duration := util.FormatDuration(v.Duration())

		ui.taskTable.SetCell(rowID, 0, tview.NewTableCell(clampString(v.Description, maxDescLen)))
		ui.taskTable.SetCell(rowID, 1, tview.NewTableCell(startedAt).SetAlign(tview.AlignRight))
		ui.taskTable.SetCell(rowID, 2, tview.NewTableCell(stoppedAt).SetAlign(tview.AlignRight))
		ui.taskTable.SetCell(rowID, 3, tview.NewTableCell(duration).SetAlign(tview.AlignRight))
		ui.taskTable.SetCell(rowID, 4, tview.NewTableCell(clampString(strings.Join(v.Tags, " "), maxTagsLen)))
		rowID++
	}
}

func clampString(v string, max int) string {
	if len(v) > max {
		return v[:max-1] + "…"
	}

	return v
}

func (ui *UI) initTaskPageLayout() {
	ui.taskFlex.
		AddItem(ui.taskTable, 0, 2, true).
		AddItem(ui.taskForm, 0, 1, true)

	ui.initTaskTable()
	ui.taskForm.SetBorder(true).SetTitle(t("Selected task"))
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
