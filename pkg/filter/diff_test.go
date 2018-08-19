package filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailfilter/pkg/config"
)

func TestNoDiff(t *testing.T) {
	old := Filters{
		{
			ID: "abcdefg",
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				MarkRead: true,
			},
		},
	}
	new := Filters{
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				MarkRead: true,
			},
		},
	}

	fd, err := Diff(old, new)
	assert.Nil(t, err)
	// No difference even if the ID is present in only one of them.
	assert.True(t, fd.Empty())
}

func TestDiffOutput(t *testing.T) {
	old := Filters{
		{
			ID: "abcdefg",
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				MarkRead: true,
				Category: config.CategoryPersonal,
			},
		},
	}
	new := Filters{
		{
			Criteria: Criteria{
				From: "{someone@gmail.com else@gmail.com}",
			},
			Action: Action{
				MarkRead: true,
				Category: config.CategoryPersonal,
			},
		},
	}

	fd, err := Diff(old, new)
	assert.Nil(t, err)

	expected := `
--- Original
+++ Current
@@ -1,6 +1,6 @@
 * Criteria:
-    from: someone@gmail.com
+    from: {someone@gmail.com else@gmail.com}
   Actions:
     categorize as: personal`
	assert.Equal(t, strings.TrimSpace(fd.String()), strings.TrimSpace(expected))
}

func someFilters() Filters {
	return Filters{
		{
			ID: "abcdefg",
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				AddLabel: "label1",
			},
		},
		{
			ID: "qwerty",
			Criteria: Criteria{
				To: "me@gmail.com",
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			ID: "zxcvb",
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
	}
}

func TestDiffAddRemove(t *testing.T) {
	old := someFilters()
	new := Filters{
		{
			Criteria: Criteria{
				From: "{someone@gmail.com else@gmail.com}",
			},
			Action: Action{
				MarkRead: true,
				Category: config.CategoryPersonal,
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				AddLabel: "label1",
			},
		},
	}

	fd, err := Diff(old, new)
	expected := FiltersDiff{
		Added:   Filters{new[0]},
		Removed: Filters{old[1]},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDiffReorder(t *testing.T) {
	old := someFilters()
	new := Filters{
		{
			Criteria: Criteria{
				To: "me@gmail.com",
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				AddLabel: "label1",
			},
		},
	}

	fd, err := Diff(old, new)
	assert.Nil(t, err)
	assert.Len(t, fd.Added, 0)
	assert.Len(t, fd.Removed, 0)
}

func TestDiffModify(t *testing.T) {
	old := someFilters()
	new := Filters{
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				AddLabel: "label1",
			},
		},
		{
			Criteria: Criteria{
				To: "{me@gmail.com you@gmail.com}",
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
	}

	fd, err := Diff(old, new)
	expected := FiltersDiff{
		Added:   Filters{new[1]},
		Removed: Filters{old[1]},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDiffAdd(t *testing.T) {
	old := someFilters()
	new := Filters{
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Action{
				AddLabel: "label1",
			},
		},
		{
			Criteria: Criteria{
				To: "me@gmail.com",
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				To: "{me@gmail.com you@gmail.com}",
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Action{
				MarkImportant: true,
			},
		},
	}

	fd, err := Diff(old, new)
	expected := FiltersDiff{
		Added:   Filters{new[2]},
		Removed: Filters{},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDiffRemove(t *testing.T) {
	old := someFilters()
	new := Filters{
		{
			Criteria: Criteria{
				To: "me@gmail.com",
			},
			Action: Action{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
	}

	fd, err := Diff(old, new)
	expected := FiltersDiff{
		Added:   Filters{},
		Removed: Filters{old[0], old[2]},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}
