{
  version: 'v1alpha3',
  rules: [
    {
      filter: {
        and: [
          {to: 'some-list'},
          {not: {from: 'blocked'}},
        ],
      },
      actions: {labels: ['mailing-list']},
    },
  ],
  tests: [
    {
      name: 'local part matches full addresses',
      messages: [
        {to: ['some-list@google.com']},
        {to: ['someone@example.com', 'SOME-LIST@google.com']},
        {cc: ['some-list@google.com']},
        {bcc: ['some-list@google.com']},
        {lists: ['some-list.google.com']},
      ],
      actions: {labels: ['mailing-list']},
    },
    {
      name: 'unrelated names do not match',
      messages: [
        {to: ['notsome-list@google.com']},
        {to: ['some-listing@google.com']},
        {},
      ],
      actions: {},
    },
    {
      name: 'negated address criteria match local parts',
      messages: [
        {from: 'blocked@example.com', to: ['some-list@google.com']},
      ],
      actions: {},
    },
  ],
}
