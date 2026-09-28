# Issue #372: compaction policy experiment

The measured comparison favors an explicit `settings.compact` boolean, default
`false`, accompanied by a suggestion near the filter limit. Hysteresis preserves
the installed representation during ordinary edits, but inferring history from
the cheapest diff introduces additional transitions and different online/offline
output. A toggle makes the large migration an intentional configuration change.

This is a runnable, read-only prototype. It uses gmailctl's real Jsonnet parser,
AST simplification, filter generation, diff rendering, and XML exporter. It does
not modify the production CLI or contact Gmail. Tests use the repository's fake
Gmail server. `settings.compact` is accepted only by this executable; the current
production CLI rejects the new field.

See [results.md](results.md) for generated measurements and sample notices.

## Policies implemented

Both policies share exactly the same compactor:

- Generate normal filters using the production compiler.
- Keep a structured top-level OR together when its complete AST has fewer than
  20 nodes/arguments and its generated query has at most 1,000 UTF-8 bytes.
- Exclude raw queries and `isEscaped` leaves from new grouping. Preserve the
  existing handling of larger expressions, nested splitting, and action expansion.
- Compact every eligible rule in compact mode. Deduplicate the resulting filters
  before counting them, as the production diff does.
- Accept the compact candidate only if its final distinct count is smaller.
  Shared split filters across different rules can make independent compaction
  increase the total. Equal-size candidates also offer no benefit for the churn.

The node budget matches the current compiler's heuristic. The byte budget is an
experimental conservative choice, not a verified Gmail query limit. The prototype
only explores the small-OR optimization proposed in this conversation; it does not
attempt to find the globally smallest equivalent set of filters.

| Concern | Inferred hysteresis | Settings toggle |
|---|---|---|
| Default behavior | Normal below 750; compact above 900 | Normal unless explicitly enabled |
| Between 750 and 900, inclusive | Generate both representations and pick fewer additions + removals; ties use normal | Use the setting |
| Existing small configs on upgrade | Unchanged | Unchanged when omitted/false |
| Ordinary edits after compaction | Usually preserve compact representation inside the band | Preserve compact representation until the setting changes |
| Crossing below 750 | Expand all eligible rules again | No mode change |
| Empty account inside the band | Chooses compact because there are fewer additions | Use the setting |
| Large rewrite inside the band | Can change mode when old matching evidence disappears | Use the setting |
| Mixed state after an interrupted migration | Choose the cheaper completion, which may be either mode | Converge toward the explicit setting |
| Offline export | Compact only above 900, as requested | Use the same setting as online commands |
| Inputs needed for online decision | Config plus installed filters | Config only |
| Production integration | Candidate generation plus selection after fetching upstream | One generation option passed from config |
| User notification | Explain threshold or candidate costs, counts, and regrouping | Explain selected setting and regrouping; suggest enabling near the limit |

There is no persistent state in the hysteresis prototype. Both threshold tests
use the normal candidate count, never the current installed/compacted count.
This prevents a successful apply from triggering its own inverse on the next run.
It does not provide guaranteed historical mode tracking after arbitrary edits.

For a fair behavioral comparison the harness generates both candidates for either
policy. A production toggle needs to generate only its selected representation;
the other candidate can be computed when useful for an optional savings notice.
The size of this experiment includes a config reader, CLI, scenario runner, and
tests and should not be read as an estimate of production patch size.

## Measurements that affect the decision

The sequential trace starts with 850 normally generated filters installed under
each policy. Synthetic rules are distinct two-field ORs with an archive action.
The toggle=true trace opts in once at the start; all later edits keep it enabled.

- Growing from 898 to 902 normal filters makes hysteresis remove 898 filters and
  add 451. An already-enabled toggle adds only two compact filters for that edit.
- Shrinking to 898 retains compact mode under hysteresis, but offline export
  returns 898 split filters instead of the installed 449 compact filters.
- Shrinking from 750 to 748 normal filters makes hysteresis add 748 and remove
  375. The enabled toggle removes one filter.
- With no installed filters, an 850-filter config selects compact mode inside
  the band. Replacing every rule on a normal account also selects compact mode.
  These are minimum-change decisions, not detection of a previous compact mode.
- Of eight current repository fixtures, one changes under full compaction:
  `05-bigdiff.jsonnet` goes from 12 to 8 filters, adding one and removing five.
  None crosses the hysteresis thresholds. This small fixture set is not evidence
  of the proportion of users affected.
- Both policies need the same global savings guard. Six rules pairing two senders
  with three subjects share five normal filters; independent compaction would
  produce six. The prototype detects that and retains normal grouping.

Both approaches have a migration problem with the current create-before-delete
apply order. The 898-to-902 hysteresis transition would temporarily require 1,349
filters. Enabling the toggle on the initial 850-filter account would require
1,275. The report calculates these peaks; its sequential trace assumes the desired
state was installed between rows, so it is not an executable migration plan for
an account with [Gmail's 1,000-filter cap](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.filters/create). Capacity-aware replacement is necessary
for those migrations regardless of the chosen policy.

## Run the comparison

From the repository root:

```sh
go run ./experiments/issue372
go test ./experiments/issue372
```

If the default Go build cache is read-only in the sandbox, prefix either command
with `GOCACHE=/tmp/gmailctl-issue372-go-cache`. The fake Gmail integration test
requires permission to open a localhost listener. It does not need credentials.

Regenerate the measured report:

```sh
go run ./experiments/issue372 > experiments/issue372/results.md
```

## Inspect a real diff and notification

Create a local snapshot in the normal representation, then compare with the
explicit compact setting. The snapshot format is a JSON array of internal filter
objects produced by this executable, not a downloaded Gmail API response.

```sh
go run ./experiments/issue372 \
  -config experiments/issue372/examples/normal.jsonnet \
  -strategy toggle -operation snapshot > /tmp/issue372-normal.json

go run ./experiments/issue372 \
  -config experiments/issue372/examples/compact.jsonnet \
  -strategy toggle -operation diff -upstream /tmp/issue372-normal.json
```

The prototype-only configuration setting is:

```jsonnet
{
  version: 'v1alpha3',
  settings: { compact: true },
  rules: [/* existing rules */],
}
```

Use `-strategy hysteresis` to select from the installed snapshot instead. Use
`-operation plan` for counts without a potentially large diff, or `-operation
export` for real Gmail XML. Export ignores `-upstream`. If no upstream snapshot is
provided, online planning models an empty account; that is deliberately visible
in the inference examples. `snapshot` writes the selected desired state locally;
there is no apply operation in this executable.

Notices go to stderr so XML, snapshots, and piped diffs stay usable. Empty filter
diffs emit no notice. The explanation counts installed filters retained by the
alternate representation and removed by the selected one, so it does not claim
that unrelated edits are merely compaction.

## Validation and limits

Focused tests cover action preservation, nested AND/NOT grouping, query-length
and AST-size boundaries, raw-expression exclusions, duplicate counts, threshold
boundaries, repeat stability, mixed upstream states, export divergence, notices,
and agreement with the production diff on every current fixture. The fake Gmail
test applies an edit to a compact account inside the band, downloads it, and
verifies a subsequent run has no diff.

The fake server checks storage and round trips, not real Gmail search evaluation
or its capacity cap. Query matching and thresholds would need live validation
before shipping. Ignored/unsupported upstream filters still consume real capacity;
this prototype only counts the filters present in its snapshot. It does not
perform rollback, schedule replacements, or persist an explicit historical mode.
