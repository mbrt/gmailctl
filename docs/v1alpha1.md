# Config v1alpha1

**Note:** This document refer to a deprecated (but still supported) version of
the configuration file. New versions of `gmailctl` will still be able to read
and apply this version of the configuration file, but the support might
disappear in the future. Please refer to the main README for details on the
newer format.

The configuration contains two important sections:

* `consts` that contains global constants that can be referenced later on by
  rules.
* `rules` that specify a set of filters that when match cause a set of actions
  to happen.

## Rule evaluation

With the help of this example, let's explain how rules evaluation works:

```yaml
- filters:
    filterA:
      - valueA
      - valueB
    filterB:
      - valueC
  # omitted actions

- filters:
    filterC:
      - valueD
      - valueE
  # omitted actions
```

1. If the rule matches, then it's applied. This means that every
   rule is in OR with all the others. In the example, given an email, if the
   first filter matches, then its actions are applied; if the second also
   matches, then its actions are also applied.
2. Within a rule, all the filters have to match, in order for the rule to match.
   This means that the filters inside a rule are in AND together. In the
   previous example, if only `filterA` matches, then the first rule is not
   applied. If both `filterA` and `filterB` match, then the rule also matches.
3. Within a filter, the listed values are in OR with each other. In the second
   rule, `filterC` matches if either `valueD` or `valueE` are present.

## Filters

The following simple filters are available:
* from
* to
* subject
* has (contains one of the given values)
* list (matches a mail list)

You can apply the special `not` operator to negate a match in this way:

```yaml
  - filters:
      not:
        to:
          - foo@bar.com
        subject:
          - Baz zorg
```

The rule will match if the email is both not directed to `foo@bar.com` and does
not contain `Baz zorg` in the subject.

## Constants

A filter can refer to global constants specified in the first section by using
the `consts` section inside the filter. All values inside the rule will be
replaced by the constants. Inside `consts` you can put again the same set of
filters of the positive case:
* from
* to
* subject
* has
* list

Example:

```yaml
version: v1alpha1
consts:
  friends:
    values:
      - pippo@gmail.com
      - pippo@hotmail.com
rules:
  - filters:
      consts:
        not:
          from:
            - friends
    actions:
      archive: true
```

is equivalent to:

```yaml
version: v1alpha1
rules:
  - filters:
      consts:
        not:
          from:
            - pippo@gmail.com
            - pippo@hotmail.com
    actions:
      archive: true
```

## Custom query

If the constraints imposed by the provided operators are not enough, it's
possible to use a custom query, by using the
[Gmail search syntax](https://support.google.com/mail/answer/7190?hl=en).

```yaml
  - filters:
      query: "foo {bar baz} list:mylist@mail.com"
    actions:
      archive: true
```

## Actions

When a filter matches, all the actions specified in a rule are applied.

The following boolean actions are available:
* archive
* delete
* markImportant
* markRead

A boolean action should be specified with a `true` value. A `false` value
is equivalent to no action.

A category can be applied to an email, by using the `category` action. Gmail
allows only one category per email and only the following categories are
supported:
* personal
* social
* updates
* forums
* promotions

A list of labels can also be applied, by using the `labels` action.

This example has one action for every type, to illustrate the usage:

```yaml
  - filters:
      from:
        - me@me.com
    actions:
      markImportant: true
      category: updates
      labels:
        - me
        - you
```

## Example

This is a more "real world" example, taken from my configuration with scrambled
values :)

```yaml
version: v1alpha1
author:
  name: Pippo Pluto
  email: pippo@gmail.com

consts:
  me:
    values:
      - pippo@gmail.com
      - pippo@hotmail.com
  spam:
    values:
      - spammer@spam.com

rules:
# important emails
  - filters:
      from:
        - myalarm@myalarm.com
    actions:
      markImportant: true
      labels:
        - alarm
# delete spammers
  - filters:
  # here we need two rules because we want to apply them if the email matches
  # one OR the other condition (blacklisted subject or blacklisted sender)
      subject:
        - foo bar baz
        - I want to spam you
    actions:
      delete: true
  - filters:
      consts:
        from:
          - spam
    actions:
      delete: true
  - filters:
      has:
        - buy this
        - buy that
    actions:
      delete: true
# mail list
  - filters:
  # archive unless directed to me
      from:
        - interesting@maillist.com
      consts:
        not:
          to:
            - me
    actions:
      archive: true
      markRead: true

  - filters:
  # always apply the label (even if directed to me)
      from:
        - interesting@maillist.com
    actions:
      labels:
        - mail-list
```
