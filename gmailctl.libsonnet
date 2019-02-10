// gmailctl standard library
//
// This contains utilities that simplify common filter operations.

// chainFilters is a function that, given a list of rules,
// returns a new list where the rules are chained togheter,
// which means that they can be interpreted as a chain
// of "if elsif elsif".
// The result is basically a list of rules where each filter
// is modified by adding an AND with the negation of all the
// previous filters.
local chainFilters(fs) =
  // utility that given a rule it returns its negated filter.
  local negate(r) = { not: r.filter };
  // recursive that goes trough all elements of arr
  local aux(arr, i, negated, running) =
    if i >= std.length(arr) then
      running
    else
      // the new rule is an AND of:
      // - the negation of all the previous rules
      // - the current rule
      local newr = {
        filter: {
          and: negated + [arr[i].filter],
        },
        actions: arr[i].actions,
      };
      aux(arr, i + 1, negated + [negate(arr[i])], running + [newr]) tailstrict;

  if std.length(fs) == 0 then []
  else aux(fs, 1, [negate(fs[0])], [fs[0]]);

// directlyTo matches only email where the recipient is in the 'TO'
// field, not the 'CC' or 'BCC' ones.
local directlyTo(recipient) = {
  and: [
    { to: recipient },
    { not: { cc: recipient } },
  ],
};

// Exported functions
{
  chainFilters: chainFilters,
  directlyTo: directlyTo,
}
