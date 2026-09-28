// Prototype setting only; the production gmailctl binary does not support it.
{
  version: 'v1alpha3',
  settings: { compact: false },
  rules: [
    {
      filter: { or: [{ list: 'list1.example.com' }, { cc: 'list1@example.com' }] },
      actions: { archive: true, labels: ['work', 'lists'] },
    },
  ],
}
