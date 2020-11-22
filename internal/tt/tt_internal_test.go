package tt

import (
	"reflect"
	"testing"
)

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
		desc, tags := parseRawDesc(v.input)
		if desc != v.expectedDesc {
			t.Errorf("case #%d description does not match: expected '%s' got '%s'", k, v.expectedDesc, desc)
		}
		if !reflect.DeepEqual(tags, v.expectedTags) {
			t.Errorf("case #%d tags do not match: expected %v got %v", k, v.expectedTags, tags)
		}
	}
}
