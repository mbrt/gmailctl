local filters = {
  me: {
    or: [
      {to: 'pippo@gmail.com'},
      {to: 'pippo@hotmail.com'},
    ]
  },

  notMe: {
    not: self.me,
  },

  lists: {
    or: [
      {list: 'list1'},
      {list: 'list2'},
      {list: 'list3'},
    ]
  },

  spam: {
    or: [
      {from: 'spammer1'},
      {from: 'spammer2'},
      {subject: 'spam mail'},
      {to: 'pippo+spammy@gmail.com'},
      {has: 'buy this thing'},
      {has: 'very important!!!'},
    ]
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
          {not: filters.me},
        ]
      },
      actions: {
        labels: ['maillist', 'onemorelabel'],
        archive: true,
      }
    },
    {
      filter: {
        and: [
          {to: 'myalias@gmail.com'},
          filters.lists,
        ]
      },
      actions: {
        markImportant: true,
      }
    },
    {
      filter: filters.spam,
      actions: {
        delete: true,
      }
    },
  ],
  tests: [
    {
      name: 'spam',
      messages: [
        {from: 'spammer1'},
        {subject: 'spam mail'},
        {to: ['pippo+spammy@gmail.com', 'someone@else.com]']},
      ],
      actions: {
        delete: true,
      }
    },
    {
      name: 'undirected lists',
      messages: [
        {
          lists: ['list1'],
          to: ['someone@else.com'],
        },
        {lists: ['list2']},
      ],
      actions: {
        labels: ['maillist', 'onemorelabel'],
        archive: true,
      }
    },
    {
      name: 'directed lists',
      messages: [
        {
          lists: ['list3'],
          to: ['pippo@gmail.com'],
        },
      ],
      actions: {},
    },
    {
      name: 'important and mail lists',
      messages: [
        {
          lists: ['list3'],
          to: ['myalias@gmail.com'],
        },
      ],
      actions: {
        archive: true,
        markImportant: true,
        labels: ['maillist', 'onemorelabel'],
      },
    },
  ],
}
