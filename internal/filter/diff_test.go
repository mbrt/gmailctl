package filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/gmail"
)

func TestNoDiff(t *testing.T) {
	old := Filters{
		{
			ID: "abcdefg",
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
				MarkRead: true,
			},
		},
	}
	new := Filters{
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
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
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
	}
	new := Filters{
		{
			Criteria: Criteria{
				From: "{someone@gmail.com else@gmail.com}",
			},
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
	}

	fd, err := Diff(old, new)
	assert.Nil(t, err)

	expected := `
--- Current
+++ TO BE APPLIED
@@ -1,6 +1,6 @@
 * Criteria:
-    from: someone@gmail.com
+    from: {someone@gmail.com else@gmail.com}
   Actions:
     mark as read
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
			Action: Actions{
				AddLabel: "label1",
			},
		},
		{
			ID: "qwerty",
			Criteria: Criteria{
				To: "me@gmail.com",
			},
			Action: Actions{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			ID: "zxcvb",
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Actions{
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
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Actions{
				MarkImportant: true,
			},
		},
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
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
			Action: Actions{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Actions{
				MarkImportant: true,
			},
		},
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
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
			Action: Actions{
				AddLabel: "label1",
			},
		},
		{
			Criteria: Criteria{
				To: "{me@gmail.com you@gmail.com}",
			},
			Action: Actions{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Actions{
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
			Action: Actions{
				AddLabel: "label1",
			},
		},
		{
			Criteria: Criteria{
				To: "me@gmail.com",
			},
			Action: Actions{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				To: "{me@gmail.com you@gmail.com}",
			},
			Action: Actions{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
		{
			Criteria: Criteria{
				Query: "-{foobar baz}",
			},
			Action: Actions{
				MarkImportant: true,
			},
		},
	}

	fd, err := Diff(old, new)
	expected := FiltersDiff{
		Added: Filters{new[2]},
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
			Action: Actions{
				MarkRead: true,
				AddLabel: "label2",
			},
		},
	}

	fd, err := Diff(old, new)
	expected := FiltersDiff{
		Removed: Filters{old[2], old[0]},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDuplicate(t *testing.T) {
	old := Filters{}
	new := Filters{
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
				MarkRead: true,
			},
		},
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
				MarkRead: true,
			},
		},
	}

	fd, err := Diff(old, new)
	assert.Nil(t, err)
	// Only one of the two identical filters is present
	assert.Equal(t, new[1:], fd.Added)
}
