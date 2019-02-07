
local me = 'pippo@gmail.com';
local spam = {
  or: [
    {from: 'spammer@spam.com'},
    {subject: 'foo bar baz'},
    {subject: 'I want to spam you'},
    {has: 'buy this'},
    {has: 'buy that'},
  ]
};

// chainFilters is a utility function that given a list of rules
// it returns a list of rules that can be interpreted as a chain
// of "if elsif elsif".
// The result is basically a list of filters where each element
// is the original one, plus the negation of all the previous
// conditions.
local chainFilters(fs) =
  // utility that given a rule it returns its negated filter.
  local negate(r) = {not: r.filter};
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
          and: negated + [arr[i].filter]
        },
        actions: arr[i].actions,
      };
      aux(arr, i + 1, negated + [negate(arr[i])], running + [newr]) tailstrict;

  aux(fs, 1, [negate(fs[0])], [fs[0]]);

{
  version: 'v1alpha3',
  author: {
    name: 'Pippo Pluto',
    email: me,
  },

  rules: [
    {
      filter: {
        from: 'myalarm@myalarm.com',
      },
      actions: {
        markImportant: true,
        labels: ['alarm'],
      },
    },
    {
      filter: spam,
      actions: {
        delete: true,
      }
    },
  ]
  + chainFilters([
    {
      filter: {
        to: me,
      },
      actions: {
        markImportant: true,
      },
    },
    {
      filter: {
        cc: me,
      },
      actions: {
        archive: true,
      },
    },
    {
      filter: {
        list: 'foobar@list.com',
      },
      actions: {
        labels: ['mylist'],
      }
    },
  ])
}
