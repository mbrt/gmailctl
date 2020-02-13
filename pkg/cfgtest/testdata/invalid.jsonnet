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
  // The two filters contradict each other.
  tests: [
    {
      name: 'both filters',
      messages: [
        {
          lists: ['list2'],
          to: ['myalias@gmail.com'],
        },
      ],
      actions: {},
    },
  ],
}
