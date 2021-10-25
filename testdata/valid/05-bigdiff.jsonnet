// This changes all filters and labels.
local filters = {
  me: {
    or: [
      { to: 'none@gmail.com' },
    ],
  },

  notMe: {
    not: filters.me,
  },

  lists: {
    or: [
      { list: 'list3' },
      { list: 'list1' },
      { list: 'list4' },
      { list: 'list6' },
    ],
  },

  spam: {
    or: [
      {
        and: [
          { from: 'spammer1' },
          { subject: 'spam mail' },
          { cc: 'foo@baz.com' },
          { bcc: 'bar@baz.com' },
        ],
      },
      { from: 'spammer2' },
      {
        and: [
          { list: 'foobaz.mail.com' },
          { not: { has: 'action needed' } },
        ],
      },
      { to: 'pippo+spammy@gmail.com' },
      { has: 'buy this thing' },
    ],
  },
};

// The config
{
  version: 'v1alpha3',
  labels: [
    { name: 'differentlabel' },
    { name: 'maillist' },
    { name: 'thirdlabel' },
  ],
  rules: [
    {
      filter: {
        and: [
          filters.lists,
          { not: filters.me },
        ],
      },
      actions: {
        labels: ['maillist', 'differentlabel', 'thirdlabel'],
        archive: true,
      },
    },
    {
      filter: filters.spam,
      actions: {
        delete: true,
      },
    },
    {
      filter: { from: 'baz+zuz@mail.com' },
      actions: {
        markImportant: true,
        forward: 'other@mail.com',
      },
    },
    {
      filter: {
        or: [
          {
            and: [
              { subject: 'hey there' },
              { from: 'notfriend@gmail.com' },
              filters.notMe,
            ],
          },
        ],
      },
      actions: {
        archive: true,
        star: true,
      },
    },
  ],
}
