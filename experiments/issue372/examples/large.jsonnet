// 451 rules produce 902 normal filters or 451 compact filters.
{
  version: 'v1alpha3',
  rules: [
    {
      filter: { or: [
        { from: 'sender-%04d@example.com' % i },
        { subject: 'topic-%04d' % i },
      ] },
      actions: { archive: true },
    }
    for i in std.range(0, 450)
  ],
}
