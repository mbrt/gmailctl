# Issue #372: measured prototype comparison

These counts use the real Jsonnet/parser/filter pipeline plus a shared experimental compactor. Synthetic cases are controlled examples, not a sample of user configurations. No Gmail requests are made.

Policies: hysteresis enters compact mode above 900 normal filters, exits below 750, and chooses fewer additions + removals inside the band (normal on ties). Export uses only the >900 threshold. The toggle obeys settings.compact, default false, for both online planning and export.

## Sequential edits

All strategies start with the same 850 normal filters installed. Toggle=true opts in on the first row. Subsequent rows use each strategy's previous output as installed state. Each cell reports selected count and additions/removals. The trace assumes successful installation between steps; peaks exceeding 1000 are explicitly infeasible under today's apply order.

| Edit | Normal candidate | Hysteresis | Toggle=false | Toggle=true | Hysteresis apply peak | Toggle=true apply peak |
|---|---:|---|---|---|---:|---:|
| Upgrade / enable toggle | 850 | normal: 850; +0 / -0 | 850; +0 / -0 | 425; +425 / -850 | 850 | 1275 |
| Grow within band | 898 | normal: 898; +48 / -0 | 898; +48 / -0 | 449; +24 / -0 | 898 | 449 |
| Cross upper threshold | 902 | compact: 451; +451 / -898 | 902; +4 / -0 | 451; +2 / -0 | 1349 | 451 |
| Drop back into band | 898 | compact: 449; +0 / -2 | 898; +0 / -4 | 449; +0 / -2 | 451 | 451 |
| Reach lower boundary | 750 | compact: 375; +0 / -74 | 750; +0 / -148 | 375; +0 / -74 | 449 | 449 |
| Cross lower threshold | 748 | normal: 748; +748 / -375 | 748; +0 / -2 | 374; +0 / -1 | 1123 | 375 |
| Grow back into band | 752 | normal: 752; +4 / -0 | 752; +4 / -0 | 376; +2 / -0 | 752 | 376 |
| Cross upper threshold again | 902 | compact: 451; +451 / -752 | 902; +150 / -0 | 451; +75 / -0 | 1203 | 451 |

## Offline export after the same edits

Differences compare export with the successfully installed result of online planning. Both toggle settings always export their selected representation.

| Edit | Hysteresis export count | Difference from installed filters |
|---|---:|---|
| Upgrade / enable toggle | 850 | +0 / -0 |
| Grow within band | 898 | +0 / -0 |
| Cross upper threshold | 451 | +0 / -0 |
| Drop back into band | 898 | +898 / -449 |
| Reach lower boundary | 750 | +750 / -375 |
| Cross lower threshold | 748 | +0 / -0 |
| Grow back into band | 752 | +0 / -0 |
| Cross upper threshold again | 451 | +0 / -0 |

## Mode inference under ambiguous history

| Scenario | Normal operations | Compact operations | Hysteresis choice |
|---|---:|---:|---|
| Unchanged normal account | 0 | 1275 | normal |
| Unchanged compact account | 1275 | 0 | compact |
| Empty account, 850-filter config | 850 | 425 | compact |
| Replace every rule on normal account | 1700 | 1275 | compact |
| Partial migration: 212 compact / 213 split rules | 636 | 639 | normal |
| Partial migration: 213 compact / 212 split rules | 639 | 636 | compact |

## Repository fixtures

Each fixture is evaluated independently, with its current normal output as the installed state. Toggle=true changes and full compaction savings are shown. These small fixtures do not establish how common affected configs are among users.

| Fixture | Normal | Compact | Rules combined | Hysteresis changes | Toggle=true changes |
|---|---:|---:|---:|---|---|
| 00-simple.jsonnet | 1 | 1 | 0 | +0 / -0 | +0 / -0 |
| 01-nodiff.jsonnet | 1 | 1 | 0 | +0 / -0 | +0 / -0 |
| 02-add-one.jsonnet | 15 | 15 | 0 | +0 / -0 | +0 / -0 |
| 03-labels.jsonnet | 8 | 8 | 0 | +0 / -0 | +0 / -0 |
| 04-labels-unspecified.jsonnet | 8 | 8 | 0 | +0 / -0 | +0 / -0 |
| 05-bigdiff.jsonnet | 12 | 8 | 1 | +0 / -0 | +1 / -5 |
| 06-bigsplit.jsonnet | 3 | 3 | 0 | +0 / -0 | +0 / -0 |
| 07-deleteall.jsonnet | 0 | 0 | 0 | +0 / -0 | +0 / -0 |

## Shared compactor guard: overlapping rules

Six rules combining each of two senders with each of three subjects produce 5 distinct normal filters. Independently compacting all six would produce 6 distinct filters. The shared guard rejects that expansion and keeps 5 filters in either policy. Equal-size candidates are also kept in the normal representation to avoid churn without savings.

## Example explanation before a diff

```text
Note: using compact filters because normal count 902 exceeds 900.
Normal generation: 902 filters; compact generation: 451 filters (451 rules combined).
This grouping choice replaces 898 installed filters that the other mode would retain.
Grouping preserves rule meaning; the diff can include replacements beyond your edits.
Capacity: the current create-before-delete apply order would peak at 1349 filters (limit 1000).
```

With the toggle disabled, the same edit keeps normal generation and suggests opting in:

```text
Note: using normal filters because settings.compact is false.
Normal generation: 902 filters; compact generation: 451 filters (451 rules combined).
This config is close to the filter limit. Setting settings.compact to true would reduce it to 451 filters.
```

Notices are suppressed when there is no filter diff. The grouping explanation is emitted only when the alternate representation would retain installed filters that the selected representation removes.
