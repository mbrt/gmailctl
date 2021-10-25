// This test doesn't specify labels and doesn't change the filters.
// The diff should be empty!
{
  version: 'v1alpha3',
  rules: [
    {
      filter: {
        list: 'maillist@google.com',
      },
      actions: {
        markImportant: false,
      },
    },
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
        forward: 'forward-address@gmail.com',
      },
    },
  ],
}
