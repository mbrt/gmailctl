local filters = {
  lists: {
    or: [
      {list: 'list1'},
      {list: 'list2'},
      {list: 'list3'},
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
          {to: 'myalias@gmail.com'},
          filters.lists,
        ]
      },
      actions: {
        markImportant: true,
      }
    },
    {
      filter: filters.lists,
      actions: {
        markImportant: false,
      }
    },
  ],
  tests: [
    {
      name: 'wrong test',
      messages: [
        {
          lists: ['list2'],
          to: ['myalias@gmail.com'],
        },
      ],
      actions: {
        archive: true,
      },
    },
    {
      name: 'another wrong test',
      messages: [
        {
          lists: ['list1'],
        },
        {
          lists: ['list2'],
        },
      ],
      actions: {
        markImportant: true,
      },
    },
  ],
}
