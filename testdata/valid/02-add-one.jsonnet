// This adds one more filter and tests several different fields of 'filter'
// and 'action'.
{
  version: 'v1alpha3',
  labels: [
    { name: 'maillist' },
    {
      name: 'label2',
      color: {
        text: 'blue',
        background: 'red',
      },
    },
  ],
  rules: [
    {
      filter: {
        list: 'maillist@google.com',
      },
      actions: {
        labels: ['maillist'],
      },
    },
    // The new filter below.
    {
      filter: {
        or: [
          { from: 'someone@gmail.com' },
          { to: 'someone-else@gmail.com' },
          {
            and: [
              { not: { subject: 'a subject' } },
              { cc: 'peeker@yahoo.com' },
            ],
          },
          { bcc: 'bccer@gmail.com' },
          { replyto: 'replyer@gmail.com' },
          { has: 'something in the body' },
          { query: 'is:muted' },
        ],
      },
      actions: {
        archive: true,
        markRead: true,
        star: true,
        markSpam: false,
        markImportant: true,
        category: 'social',
        labels: [
          'maillist',
          'label2',
        ],
        forward: 'forward-address@gmail.com',
      },
    },
  ],
}
