local filters = {
  me: {
    or: [
      { to: 'pippo@gmail.com' },
      { to: 'pippo@hotmail.com' },
    ],
  },

  notMe: {
    not: self.me,
  },

  lists: {
    or: [
      { list: 'list1' },
      { list: 'list2' },
      { list: 'list3' },
    ],
  },

  spam: {
    or: [
      { from: 'spammer1' },
      { from: 'spammer2' },
      { subject: 'spam mail' },
      { to: 'pippo+spammy@gmail.com' },
      { has: 'buy this thing' },
      { has: 'very important!!!' },
    ],
  },
};

// The config
{
  version: 'v1alpha3',
  rules: [
    {
      filter: {
        and: [
          filters.lists,
          { not: filters.me },
        ],
      },
      actions: {
        labels: ['maillist', 'onemorelabel'],
        archive: true,
      },
    },
    {
      filter: {
        and: [
          { to: 'myalias@gmail.com' },
          filters.lists,
        ],
      },
      actions: {
        markImportant: true,
      },
    },
    {
      filter: filters.spam,
      actions: {
        delete: true,
      },
    },
    {
      filter: {
        or: [
          { list: 'list%i' % i }
          for i in std.range(0, 50)
        ],
      },
      actions: {
        archive: true,
      },
    },
  ],
}
