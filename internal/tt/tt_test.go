// nolint:thelper
package tt_test

import (
	"errors"
	"reflect"
	"testing"
	"time"
	"tt/internal/tt"
)

func TestBasics(t *testing.T) {
	app := newTestApp(t)

	testBasicsNewTask(t, app)       // Creates brand new task from scratch.
	testBasicsCorrectedTask(t, app) // Changes the current task tags (and revert the change).
	testBasicsStopTask(t, app)      // Stops the current running task.
}

func testBasicsStopTask(t *testing.T, app *tt.TT) {
	stopped, err := app.Stop()
	if err != nil {
		t.Error(err)
	}

	if stopped == nil {
		t.Fatal("expected stopped to be non-nil")
	}
	if stopped.Description != "desc" {
		t.Errorf("invalid Description, expected 'desc' got '%s'", stopped.Description)
	}
	if !reflect.DeepEqual(stopped.Tags, []string{"@tag"}) {
		t.Errorf("invalid Tags, expected [@tag] got %v", stopped.Tags)
	}
	if time.Since(stopped.StoppedAt) > 1*time.Second {
		t.Errorf("expected StoppedAt to be just now, got %v", stopped.StartedAt)
	}
	if time.Since(stopped.StartedAt) > 1*time.Second {
		t.Errorf("expected StartedAt to be just now, got %v", stopped.StartedAt)
	}

	stopped, err = app.Stop()
	if !errors.Is(err, tt.ErrNoCurrentTask) {
		t.Errorf("expected ErrNoCurrentTask, got %v", err)
	}
	if stopped != nil {
		t.Errorf("expected stopped to be nil, got %v", stopped)
	}
}

// nolint: cyclop
func testBasicsCorrectedTask(t *testing.T, app *tt.TT) {
	// No change
	prev, next, err := app.Start("desc @tag")
	if !errors.Is(err, tt.ErrContinue) {
		t.Errorf("expected error to be ErrContinue, got %v", err)
	}
	if prev == nil {
		t.Fatal("expected prev to non nil")
	}
	if next == nil {
		t.Fatal("expected next to non nil")
	}
	if prev.ID != next.ID {
		t.Error("expected identical task ID")
	}

	// Tag change
	prev, next, err = app.Start("desc @other-tag")
	if err != nil {
		t.Error(err)
	}
	if prev == nil {
		t.Fatal("expected prev to non nil")
	}
	if next == nil {
		t.Fatal("expected next to non nil")
	}
	if prev.ID != next.ID {
		t.Error("expected identical task ID")
	}
	if !reflect.DeepEqual(next.Tags, []string{"@other-tag"}) {
		t.Errorf("expected tags to be [@other-tag], got %v", next.Tags)
	}

	// Task change
	prev, next, err = app.Start("otherdesc")
	if err != nil {
		t.Error(err)
	}
	if prev == nil {
		t.Fatal("expected prev to non nil")
	}
	if next == nil {
		t.Fatal("expected next to non nil")
	}
	if prev.ID == next.ID {
		t.Error("expected task IDs to be different")
	}
	if !reflect.DeepEqual(next.Tags, []string{"@other-tag"}) { // tag reuse when no tag specified
		t.Errorf("expected tags to be [@other-tag], got %v", next.Tags)
	}

	if _, _, err := app.Start("desc @tag"); err != nil {
		t.Error(err)
	}
}

func testBasicsNewTask(t *testing.T, app *tt.TT) {
	prev, next, err := app.Start("desc @tag")
	if err != nil {
		t.Error(err)
	}

	if prev != nil {
		t.Error("expected prev to be nil")
	}

	if next == nil {
		t.Fatal("expected next to be non-nil")
	}

	if next.Description != "desc" {
		t.Errorf("invalid Description, expected 'desc' got '%s'", next.Description)
	}
	if !reflect.DeepEqual(next.Tags, []string{"@tag"}) {
		t.Errorf("invalid Tags, expected [@tag] got %v", next.Tags)
	}
	if !next.StoppedAt.IsZero() {
		t.Errorf("expected StoppedAt to be zero, got %v", next.StoppedAt)
	}
	if time.Since(next.StartedAt) > 1*time.Second {
		t.Errorf("expected StartedAt to be just now, got %v", next.StartedAt)
	}
}

func newTestApp(t *testing.T) *tt.TT {
	app, err := tt.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := app.Close(); err != nil {
			t.Error(err)
		}
	})

	return app
}

func TestParseRawDesc(t *testing.T) {
	cases := []struct {
		input        string
		expectedDesc string
		expectedTags []string
	}{
		{"", "", nil},
		{"@", "", []string{"@"}},
		{"foo", "foo", nil},
		{"foo @bar", "foo", []string{"@bar"}},
		{"foo @bar @baz", "foo", []string{"@bar", "@baz"}},
		{"foo @bar baz @booze", "foo @bar baz", []string{"@booze"}},
		{
			"stupidly long description with many tags just to see how the line will break on the grid, here comes the tag which @are @not @poundtags @since @you @cannot @type @them @in @cli @without @escaping", // nolint:lll
			"stupidly long description with many tags just to see how the line will break on the grid, here comes the tag which",
			[]string{"@are", "@cannot", "@cli", "@escaping", "@in", "@not", "@poundtags", "@since", "@them", "@type", "@without", "@you"}, // nolint:lll
		},
	}

	for k, v := range cases {
		desc, tags := tt.ParseRawDesc(v.input)
		if desc != v.expectedDesc {
			t.Errorf("case #%d description does not match: expected '%s' got '%s'", k, v.expectedDesc, desc)
		}
		if !reflect.DeepEqual(tags, v.expectedTags) {
			t.Errorf("case #%d tags do not match: expected %v got %v", k, v.expectedTags, tags)
		}
	}
}
