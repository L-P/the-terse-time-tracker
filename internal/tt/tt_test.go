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

func testBasicsCorrectedTask(t *testing.T, app *tt.TT) {
	created, updated, err := app.Start("desc @tag")
	if err != nil {
		t.Error(err)
	}
	if created != nil {
		t.Errorf("expected created to be nil, got %v", created)
	}
	if updated != nil {
		t.Errorf("expected created to be nil, got %v", updated)
	}

	created, updated, err = app.Start("desc @other-tag")
	if err != nil {
		t.Error(err)
	}
	if created != nil {
		t.Errorf("expected created to be nil, got %v", created)
	}
	if updated == nil {
		t.Fatal("expected updated to be non-nil")
	}
	if !reflect.DeepEqual(updated.Tags, []string{"@other-tag"}) {
		t.Errorf("expected tags to be [@other-tag], got %v", updated.Tags)
	}

	if _, _, err := app.Start("desc @tag"); err != nil {
		t.Error(err)
	}
}

func testBasicsNewTask(t *testing.T, app *tt.TT) {
	created, updated, err := app.Start("desc @tag")
	if err != nil {
		t.Error(err)
	}

	if updated != nil {
		t.Error("expected updated to be nil")
	}

	if created == nil {
		t.Fatal("expected created to be non-nil")
	}

	if created.Description != "desc" {
		t.Errorf("invalid Description, expected 'desc' got '%s'", created.Description)
	}
	if !reflect.DeepEqual(created.Tags, []string{"@tag"}) {
		t.Errorf("invalid Tags, expected [@tag] got %v", created.Tags)
	}
	if !created.StoppedAt.IsZero() {
		t.Errorf("expected StoppedAt to be zero, got %v", created.StoppedAt)
	}
	if time.Since(created.StartedAt) > 1*time.Second {
		t.Errorf("expected StartedAt to be just now, got %v", created.StartedAt)
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
