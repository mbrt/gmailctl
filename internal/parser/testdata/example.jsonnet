local me = {
  or: [
    { to: 'pippo@gmail.com' },
    { to: 'pippo@hotmail.com' },
  ],
};
local lists = {
  or: [
    { list: 'list1' },
    { list: 'list2' },
    { list: 'list3' },
  ],
};
local spam = {
  or: [
    { from: 'spammer1' },
    { from: 'spammer2' },
    { subject: 'spam mail' },
    { to: 'pippo+spammy@gmail.com' },
    { has: 'buy this thing' },
    { has: 'very important!!!' },
  ],
};

{
  version: 'v1alpha3',
  rules: [
    {
      filter: {
        and: [
          lists,
          { not: me },
        ],
      },
      actions: {
        labels: ['maillist'],
        archive: true,
      },
    },
    {
      filter: {
        and: [
          { to: 'myalias@gmail.com' },
          lists,
        ],
      },
      actions: {
        markImportant: true,
      },
    },
    {
      filter: spam,
      actions: {
        delete: true,
      },
    },
  ],
}
