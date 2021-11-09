// This test changes all the labels.
{
  version: 'v1alpha3',
  labels: [
    // The label 'maillist' is gone.
    // The label 'label2' doesn't change. We just don't specify the color.
    { name: 'label2' },
    // One more label.
    { name: 'label3' },
    // And one more label with a color.
    {
      name: 'label4',
      color: {
        text: 'black',
        background: 'white',
      },
    },
  ],
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
