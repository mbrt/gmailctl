{
  version: 'v1alpha3',
  labels: [
    { name: 'maillist' },
  ],
  rules: [
    {
      filter: {
        list: 'maillist@google.com',
      },
      actions: {
        labels: ['maillist'],
      }
    },
  ],
}
