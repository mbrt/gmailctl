// This tests that filters with long lists of operands are automatically split.
{
  version: 'v1alpha3',
  rules: [
    {
      filter: {
        or: [
          { list: 'list%i' % i }
          for i in std.range(0, 50)
        ],
      },
      actions: {
        archive: true,
      },
    },
  ],
}
