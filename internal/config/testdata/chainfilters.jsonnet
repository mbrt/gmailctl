local lib = import 'gmailctl.libsonnet';

local me = 'pippo@gmail.com';
local ruleA = {
  filter: {
    to: me,
  },
  actions: {
    markImportant: true,
  },
};
local ruleB = {
  filter: {
    cc: me,
  },
  actions: {
    archive: true,
  },
};
local ruleC = {
  filter: {
    list: 'foobar@list.com',
  },
  actions: {
    labels: ['mylist'],
  },
};

{
  version: 'v1alpha3',
  rules: [
    {
      filter: lib.directlyTo(me),
      actions: {
        markImportant: true,
      },
    },
  ]

  // Empty list should not err
  + lib.chainFilters([])
  // Single rule should not err
  + lib.chainFilters([ruleB])
  // Chain 3 rules
  + lib.chainFilters([ruleA, ruleB, ruleC])
}
