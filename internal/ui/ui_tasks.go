package ui

import (
	"strings"
	"time"
	"tt/internal/tt"
	"tt/internal/util"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (ui *UI) initTaskPageLayout() {
	ui.taskFlex.
		AddItem(ui.taskTable, 0, 2, true).
		AddItem(ui.taskForm, 0, 1, true)

	ui.initTaskTable()
	ui.taskForm.SetBorder(true).SetTitle(t("Selected task"))
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
		ui.updateTaskForm(selected)
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
	taskFormFieldIndexDescription = iota
	taskFormFieldIndexStartedAt
	taskFormFieldIndexStoppedAt
	taskFormFieldIndexTags
)

func (ui *UI) updateTaskForm(task tt.Task) {
	ui.taskForm.
		Clear(true).
		AddInputField(t("Description"), task.Description, 0, nil, nil).
		AddInputField(t("Started at"), formDate(task.StartedAt), len(formDateTimeFormat), nil, nil).
		AddInputField(t("Stopped at"), formDate(task.StoppedAt), len(formDateTimeFormat), nil, nil).
		AddInputField(t("Tags"), strings.Join(task.Tags, " "), 0, nil, nil).
		AddButton(t("Save"), ui.saveTaskFormTask).
		AddButton(t("Delete"), ui.deleteSelectedTask)
}

func (ui *UI) getTaskFormTask() (tt.Task, error) {
	selected, err := ui.selectedTask()
	if err != nil {
		return tt.Task{}, err
	}

	// Indices are from the input order in the form definition in updateTaskForm.
	value := func(i int) string {
		return ui.taskForm.GetFormItem(i).(*tview.InputField).GetText()
	}

	_, tags := tt.ParseRawDesc(value(taskFormFieldIndexTags))

	location := time.Now().Location()
	startedAt, err := time.ParseInLocation(formDateTimeFormat, value(taskFormFieldIndexStartedAt), location)
	if err != nil {
		return tt.Task{}, tt.InvalidInputError(err.Error())
	}

	var stoppedAt time.Time
	if str := value(taskFormFieldIndexStoppedAt); str != "" {
		stoppedAt, err = time.ParseInLocation(formDateTimeFormat, str, location)
		if err != nil {
			return tt.Task{}, tt.InvalidInputError(err.Error())
		}
	}

	if !stoppedAt.IsZero() && startedAt.After(stoppedAt) {
		startedAt, stoppedAt = stoppedAt, startedAt
	}

	return tt.Task{
		ID:          selected.ID,
		Description: value(taskFormFieldIndexDescription),
		StartedAt:   startedAt,
		StoppedAt:   stoppedAt,
		Tags:        tags,
	}, nil
}

func (ui *UI) saveTaskFormTask() {
	ui.app.SetFocus(ui.taskTable)
	if len(ui.tasks) == 0 {
		return
	}

	task, err := ui.getTaskFormTask()
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

	// nolint: gocritic // on purpose, temp representation
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
