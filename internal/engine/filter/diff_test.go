package filter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gmailctl/internal/engine/gmail"
)

func TestNoDiff(t *testing.T) {
	prev := Filters{
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
	curr := Filters{
		{
			Criteria: Criteria{
				From: "someone@gmail.com",
			},
			Action: Actions{
				MarkRead: true,
			},
		},
	}

	fd, err := Diff(prev, curr, false)
	assert.Nil(t, err)
	// No difference even if the ID is present in only one of them.
	assert.True(t, fd.Empty())
}

func TestDiffOutput(t *testing.T) {
	prev := Filters{
		{
			ID: "abcdefg",
			Criteria: Criteria{
				From:  "someone@gmail.com",
				Query: "(a b) subject:(foo bar)",
			},
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
	}
	curr := Filters{
		{
			Criteria: Criteria{
				From:  "{someone@gmail.com else@gmail.com}",
				Query: "(a c) subject:(foo baz)",
			},
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
	}

	fd, err := Diff(prev, curr, false)
	assert.Nil(t, err)

	expected := `
--- Current
+++ TO BE APPLIED
@@ -1,15 +1,15 @@
 * Criteria:
-    from: someone@gmail.com
+    from: {someone@gmail.com else@gmail.com}
     query: 
       (
         a
-        b
+        c
       )
       subject:(
         foo
-        bar
+        baz
       )
   Actions:
     mark as read
     categorize as: personal`
	assert.Equal(t, strings.TrimSpace(fd.String()), strings.TrimSpace(expected))
}

func TestDiffOutputWithGmailSearchURL(t *testing.T) {
	prev := Filters{
		{
			ID: "abcdefg",
			Criteria: Criteria{
				From:  "someone@gmail.com",
				Query: "(a b) subject:(foo bar)",
			},
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
	}
	curr := Filters{
		{
			Criteria: Criteria{
				From:  "{someone@gmail.com else@gmail.com}",
				Query: "(a c) subject:(foo baz)",
			},
			Action: Actions{
				MarkRead: true,
				Category: gmail.CategoryPersonal,
			},
		},
	}

	fd, err := Diff(prev, curr, true)
	assert.Nil(t, err)

	expected := `
--- Current
+++ TO BE APPLIED
@@ -1,17 +1,17 @@
-# Search: from:someone@gmail.com (a b) subject:(foo bar)
-# URL: https://mail.google.com/mail/u/0/#search/from%3Asomeone%40gmail.com+%28a+b%29+subject%3A%28foo+bar%29
+# Search: from:{someone@gmail.com else@gmail.com} (a c) subject:(foo baz)
+# URL: https://mail.google.com/mail/u/0/#search/from%3A%7Bsomeone%40gmail.com+else%40gmail.com%7D+%28a+c%29+subject%3A%28foo+baz%29
 * Criteria:
-    from: someone@gmail.com
+    from: {someone@gmail.com else@gmail.com}
     query: 
       (
         a
-        b
+        c
       )
       subject:(
         foo
-        bar
+        baz
       )
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
	prev := someFilters()
	curr := Filters{
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

	fd, err := Diff(prev, curr, false)
	expected := FiltersDiff{
		Added:   Filters{curr[0]},
		Removed: Filters{prev[1]},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDiffReorder(t *testing.T) {
	prev := someFilters()
	curr := Filters{
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

	fd, err := Diff(prev, curr, false)
	assert.Nil(t, err)
	assert.Len(t, fd.Added, 0)
	assert.Len(t, fd.Removed, 0)
}

func TestDiffModify(t *testing.T) {
	prev := someFilters()
	curr := Filters{
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

	fd, err := Diff(prev, curr, false)
	expected := FiltersDiff{
		Added:   Filters{curr[1]},
		Removed: Filters{prev[1]},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDiffAdd(t *testing.T) {
	prev := someFilters()
	curr := Filters{
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

	fd, err := Diff(prev, curr, false)
	expected := FiltersDiff{
		Added: Filters{curr[2]},
	}
	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDiffRemove(t *testing.T) {
	prev := someFilters()
	curr := Filters{
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

	fd, err := Diff(prev, curr, false)
	expected := FiltersDiff{
		Removed: Filters{prev[2], prev[0]},
	}

	assert.Nil(t, err)
	assert.Equal(t, expected, fd)
}

func TestDuplicate(t *testing.T) {
	prev := Filters{}
	curr := Filters{
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

	fd, err := Diff(prev, curr, false)
	assert.Nil(t, err)
	// Only one of the two identical filters is present
	assert.Equal(t, curr[1:], fd.Added)
}

func TestIndent(t *testing.T) {
	testCases := []struct{ name, query, want string }{
		{"no_newline_necessary", `from:"foo bar"`, `from:"foo bar"`},
		{"quotes", `from:"foo bar" "another foo bar" "good thing gmail doesn't support escaping" "re: something"`, `
  from:"foo bar"
  "another foo bar"
  "good thing gmail doesn't support escaping"
  "re: something"`},
		{"parens", `(a b) subject:(foo bar "re: something")`, `
  (
    a
    b
  )
  subject:(
    foo
    bar
    "re: something"
  )`},
		{"negations_and_braces", `-(x {y -z} l)`, `
  -(
    x
    {
      y
      -z
    }
    l
  )`},
		{"unicode", `日(本)語`, `
  日(
    本
  )
  語`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := indent(tc.query, 0)
			assert.Equal(t, tc.want, got)
		})
	}
}
