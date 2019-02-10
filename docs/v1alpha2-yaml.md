# YAML config v1alpha2

**Note:** This document refer to a deprecated (but still supported) version of
the configuration file. New versions of `gmailctl` will still be able to read
and apply this version of the configuration file, but the support might
disappear in the future. Please refer to the main README for details on the
newer format.

Simple example:

```yaml
version: v1alpha2
filters:
  - name: me
    query:
      or:
        - to: pippo@gmail.com
        - to: pippo@hotmail.com
rules:
  - filter:
      and:
        - list: geeks@newsletter.com
        - not:
            name: me
    actions:
      archive: true
      labels:
        - news
```

The YAML configuration file contains two important sections:

* `filters` that contains named filters that can be called up by subsequent
  filters or rules.
* `rules` that specify a filter expression and a set of actions that will be
  applied if the filter matches.

We will see all the features of the configuration file in the following
sections.

## Search operators

Search operators are the same as the ones you find in the Gmail filter
interface:

* `from`: the mail comes from the given address
* `to`: the mail is delivered to the given address
* `subject`: the subject contains the given words
* `has`: the mail contains the given words

In addition to those visible in the Gmail interface, you can specify natively
the following common operators:

* `list`: the mail is directed to the given mail list
* `cc`: the mail has the given address as CC destination

One more special function is given if you need to use less common operators<sup
id="a1">[1](#f1)</sup>, or want to compose your query manually:

* `query`: passes the given contents verbatim to the Gmail filter, without
  escaping or interpreting the contents in any way.

Example:

```yaml
version: v1alpha2
rules:
  - filter:
      subject: important mail
    actions:
      markImportant: true
  - filter:
      query: "dinner AROUND 5 friday has:spreadsheet"
    actions:
      delete: true
```

## Logic operators

Filters can contain only one expression. If you want to combine multiple of them
in the same rule, you have to use logic operators (and, or, not). These
operators do what you expect:

* `and`: is true only if all the sub-expressions are also true
* `or`: is true if one or more sub-expressions are true
* `not`: is true if the sub-expression is false.

Example:

```yaml
version: v1alpha2
rules:
  - filter:
      or:
        - from: foo
        - and:
            - list: bar
            - not:
                to: baz
    actions:
      markImportant: true
```

This composite filter marks the incoming mail as important if:

* the message comes from "foo", _or_
* it is coming from the mailing list "bar" _and_ _not_ directed to "baz"

## Named filters

Filters can be named and referenced in other filters or rules. This allows
reusing concepts and so avoid repetition.

Example:

```yaml
version: v1alpha2
filters:
  - name: toMe
    query:
      or:
        - to: myself@gmail.com
        - to: myself@yahoo.com
  - name: notToMe
    query:
      not:
        name: toMe

rules:
  - filter:
      and:
        - from: foobar
        - name: notToMe
    actions:
      delete: true
  - filter:
      name: toMe
    actions:
      labels:
        - directed
```

In this example, two named filters are defined. The `toMe` filter gives a name
to emails directed to myself@gmail.com or to myself@yahoo.com. The `notToMe`
filter negates the `toMe` filter, with a `not` operator. Similarly, the two
rules reference the two named filters above. The `name` reference is basically
copying the definition of the filter in place.

The example is effectively equivalent to this one:

```yaml
version: v1alpha2
rules:
  - filter:
      and:
        - from: foobar
        # Was "name: notToMe"
        - not:
            # Inside "notToMe" there was "name: me", so its definition
            # got replaced here
            or:
              - to: myself@gmail.com
              - to: myself@yahoo.com
    actions:
      delete: true
  - filter:
      # Was "name: toMe"
      or:
        - to: myself@gmail.com
        - to: myself@yahoo.com
    actions:
      labels:
        - directed
```

Note that filters can reference only filters previously defined, to avoid cyclic
dependencies.

## Actions

Every rule is a composition of a filter and a set of actions. Those actions will
be applied to all the incoming emails that pass the rule's filter. These actions
are the same as the ones in the Gmail interface:

* `archive: true`: the message will skip the inbox;
* `delete: true`: the message will go directly to the trash can;
* `markRead: true`: the message will be mark as read automatically;
* `star: true`: star the message;
* `markSpam: false`: do never mark these messages as spam. Note that setting this
  field to `true` is _not_ supported by Gmail (I don't know why);
* `markImportant: true`: always mark the message as important, overriding Gmail
  heuristics;
* `markImportant: false`: do never mark the message as important, overriding
  Gmail heuristics;
* `category: <CATEGORY>`: force the message into a specific category (supported
  categories are "personal", "social", "updates", "forums", "promotions");
* `labels: [list, of, labels]`: an array of labels to apply to the message. Note
  that these labels have to be already present in your settings (they won't be
  created automatically), and you can specify multiple labels (normally Gmail
  allows to specify only one label per filter).

Example:

```yaml
version: v1alpha2
rules:
  - filter:
  - filter:
      from: love@gmail.com
    actions:
      markImportant: true
      category: personal
      labels:
        - family
        - P1
```

## Footnotes

<b id="f1">1</b>: See [Search operators you can use with
Gmail](https://support.google.com/mail/answer/7190?hl=en) [↩](#a1).
