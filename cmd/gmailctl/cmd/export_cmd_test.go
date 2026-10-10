package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gmailv1 "google.golang.org/api/gmail/v1"

	"github.com/mbrt/gmailctl/internal/engine/api"
	"github.com/mbrt/gmailctl/internal/engine/filter"
	"github.com/mbrt/gmailctl/internal/fakegmail"
)

type testAPIProvider struct {
	svc *gmailv1.Service
	err error
}

func (p testAPIProvider) Service(ctx context.Context, cfgDir string) (*gmailv1.Service, error) {
	if p.err != nil {
		return nil, p.err
	}
	return p.svc, nil
}

func (p testAPIProvider) ResetConfig(cfgDir string) error {
	return nil
}

func (p testAPIProvider) InitConfig(cfgDir string, port int) error {
	return nil
}

func createTestConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.jsonnet")

	content := `{
  version: "v1alpha3",
  author: {
    name: "Test User",
    email: "test@example.com"
  },
  rules: [
    {
      filter: {
        from: "alice@example.com"
      },
      actions: {
        markRead: true
      }
    },
    {
      filter: {
        from: "bob@example.com"
      },
      actions: {
        star: true
      }
    }
  ]
}`
	err := os.WriteFile(cfgPath, []byte(content), 0600)
	require.NoError(t, err)
	return cfgPath
}

func TestExportCmdFlags(t *testing.T) {
	newFlag := exportCmd.Flags().Lookup("new")
	require.NotNil(t, newFlag)
	assert.Equal(t, "false", newFlag.DefValue)

	changesOnlyFlag := exportCmd.Flags().Lookup("changes-only")
	require.NotNil(t, changesOnlyFlag)
	assert.Equal(t, "false", changesOnlyFlag.DefValue)
}

func TestExportAllRules(t *testing.T) {
	cfgPath := createTestConfig(t)
	var buf bytes.Buffer

	err := exportWithOut(cfgPath, &buf, false, false)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "<title>Mail Filters</title>")
	assert.Contains(t, out, `name="from" value="alice@example.com"`)
	assert.Contains(t, out, `name="shouldMarkAsRead" value="true"`)
	assert.Contains(t, out, `name="from" value="bob@example.com"`)
	assert.Contains(t, out, `name="shouldStar" value="true"`)
}

func TestExportNewRulesWithFakeAPI(t *testing.T) {
	cfgPath := createTestConfig(t)
	ctx := context.Background()
	svc := fakegmail.NewService(ctx, t)
	gapi := api.NewFromService(svc)

	prevProvider := APIProvider
	t.Cleanup(func() {
		APIProvider = prevProvider
	})
	APIProvider = testAPIProvider{svc: svc}

	t.Run("no upstream filters exports all local as new", func(t *testing.T) {
		var buf bytes.Buffer
		err := exportWithOut(cfgPath, &buf, false, true)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, `name="from" value="alice@example.com"`)
		assert.Contains(t, out, `name="from" value="bob@example.com"`)
	})

	t.Run("upstream contains alice exports only bob", func(t *testing.T) {
		aliceFilter := filter.Filter{
			Criteria: filter.Criteria{From: "alice@example.com"},
			Action:   filter.Actions{MarkRead: true},
		}
		err := gapi.AddFilters(filter.Filters{aliceFilter})
		require.NoError(t, err)

		var buf bytes.Buffer
		err = exportWithOut(cfgPath, &buf, false, true)
		require.NoError(t, err)

		out := buf.String()
		assert.NotContains(t, out, `name="from" value="alice@example.com"`)
		assert.Contains(t, out, `name="from" value="bob@example.com"`)
		assert.Contains(t, out, `name="shouldStar" value="true"`)
	})

	t.Run("upstream contains both exports empty feed", func(t *testing.T) {
		bobFilter := filter.Filter{
			Criteria: filter.Criteria{From: "bob@example.com"},
			Action:   filter.Actions{Star: true},
		}
		err := gapi.AddFilters(filter.Filters{bobFilter})
		require.NoError(t, err)

		var buf bytes.Buffer
		err = exportWithOut(cfgPath, &buf, false, true)
		require.NoError(t, err)

		out := buf.String()
		assert.Contains(t, out, "<title>Mail Filters</title>")
		assert.NotContains(t, out, "<entry>")
	})
}

func TestExportNewRulesAPIError(t *testing.T) {
	cfgPath := createTestConfig(t)
	prevProvider := APIProvider
	t.Cleanup(func() {
		APIProvider = prevProvider
	})
	APIProvider = testAPIProvider{err: fmt.Errorf("connection refused")}

	var buf bytes.Buffer
	err := exportWithOut(cfgPath, &buf, false, true)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot connect to Gmail")
}
