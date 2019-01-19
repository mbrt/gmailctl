# gmailctl
[![Go Report Card](https://goreportcard.com/badge/github.com/mbrt/gmailctl)](https://goreportcard.com/report/github.com/mbrt/gmailctl)
[![Build Status](https://travis-ci.org/mbrt/gmailctl.svg?branch=master)](https://travis-ci.org/mbrt/gmailctl)

This utility helps you generate and maintain Gmail filters in a declarative way.
It has a Yaml configuration file that aims to be more simple to write and
maintain than using the Gmail web interface, to categorize, label, archive and
manage automatically your inbox.

## Motivation

If you have Gmail and have (like me) to maintain a lot of filters, because you
want to apply labels, get rid of spam or categorize your emails, then you
probably have (like me) a very long list of messy filters. Then the day that you
actually want to understand why a certain message got labeled in a certain way
comes. You scroll through that horrible mess and you wish you could
find-and-replace stuff, check the change before applying it, refactor some
filters together... in a way treat them like you do with your code!

Gmail allows to import and export filters in XML format. This can be used to
maintain them in some better way... but dear Lord, no! Not by hand! That's what
most other tools do: providing some kind of DSL that generate XML filters that
can be imported in your settings... by hand [this is the approach of the popular
[antifuchs/gmail-britta](https://github.com/antifuchs/gmail-britta) for
example].

Gmail happens to have also a neat API that we can use to automate the import
step as well, so to eliminate all manual, slow tasks to be done with the Gmail
settings.

This project then exists to provide to your Gmail filters:

1. Maintainability;
2. An easy to understand, declarative, composable language;
3. A builtin query simplifier, to keep the size of your filters down (Gmail has
   a limit of 1500 chars per filter);
4. Ability to review your changes before applying them;
5. Automatic update of the settings (no manual import) in seconds.

## Usage

[![asciicast](https://asciinema.org/a/BU00aXIcZV9bYWRko7i7LnugY.png)](https://asciinema.org/a/BU00aXIcZV9bYWRko7i7LnugY)

TODO

Make sure to setup your [$GOPATH](https://golang.org/doc/code.html#GOPATH) correctly, including the `bin` subdirectory in your `$PATH`.

```
go get github.com/mbrt/gmailctl/cmd/gmailctl
go install github.com/mbrt/gmailctl/cmd/gmailctl
gmailctl init
# edit the config file in ~/.gmailctl/config.yaml
gmailctl apply
```

where `config.yaml` is the configuration file containing the filtering rules
(see [Configuration](#configuration)). The utility will guide you through
setting up the Gmail APIs and update your settings without leaving your command
line.

**NOTE:** It's recommended to backup your current configuration before to apply
the generated one for the first time. Your current filters will be wiped and
replaced with the ones specified in the config file. The diff you'll get during
the first run will probably be pretty big, but from that point on, all changes
should generate a small and simple to review diff.

## Configuration

**NOTE:** The configuration format is still in alpha and might change in the
future. If you are looking for the deprecated version `v1alpha1`, please refer
to [docs/v1alpha1.md](docs/v1alpha1.md).

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

### Search operators

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

### Logic operators

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

### Named filters

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

### Actions

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

## Tips and tricks

TODO

## Comparison with existing projects

[gmail-britta](https://github.com/antifuchs/gmail-britta) has similar
motivations and is quite popular. The difference between that project and
this one are:

* `gmail-britta` uses a custom DSL (versus YAML in `gmailctl`)
* `gmail-britta` is imperative because it allows you to write arbitrary Ruby
  code in your filters (versus pure declarative for `gmailctl`)
* `gmail-britta` allows to write complex chains of filters, but fails to provide
  easy ways to write reasonably easy filters <sup id="a2">[2](#f2)</sup>.
* `gmail-britta` exports only to the Gmail XML format. You have to import the
  filters yourself by using the Gmail web interface, manually delete the filters
  you updated and import only the new ones. This process becomes tedious very
  quickly and you will resort to quickly avoid using the tool when in a hurry.
  `gmailctl` provides you this possibility, but also allows you to review your
  changes and update the filters by using the Gmail APIs, without you having to
  do anything manually.
* `gmailctl` tries to workaround certain limitations in Gmail (like applying
  multiple labels with the same filter) `gmail-britta` tries to workaround
  others (chain filtering).
* chain filtering is not supported in `gmailctl` by design. The declarative
  nature of the configuration makes it that every rule that matches is applied,
  just like Gmail does.

In short `gmailctl` takes the declarative approach to Gmail filters
configuration, hoping it stays simpler to read and maintain, sacrificing complex
scenarios handled instead by `gmail-britta` (like chaining), and provides the
automatic update that will save you time while you are iterating through new
versions of your filters.

## Footnotes

<b id="f1">1</b>: See [Search operators you can use with
Gmail](https://support.google.com/mail/answer/7190?hl=en) [↩](#a1).

<b id="f2">2</b>:

Try to write the equivalent of this filter with `gmail-britta`:

```yaml
version: v1alpha1
consts:
  spammers:
    values:
      - pippo@gmail.com
      - pippo@hotmail.com
  spamSubjects:
    values:
      - buy this
      - buy my awesome product
rules:
  - filters:
      consts:
        from:
          - spammers
    actions:
      delete: true
  - filters:
      consts:
        subject:
          - spamSubjects
    actions:
      delete: true
```

It becomes something like this:

```ruby
#!/usr/bin/env ruby

# NOTE: This file requires the latest master (30/07/2018) of gmail-britta.
# The Ruby repos are not up to date

require 'rubygems'
require 'gmail-britta'

SPAM_EMAILS = %w{pippo@gmail.com pippo@hotmail.com}
SPAM_SUBJECTS = ['"buy this"', '"buy my awesome product"']

puts(GmailBritta.filterset(:me => MY_EMAILS) do
       # Spam
       filter {
         has [{:or => "from:(#{SPAM_EMAILS.join("|")})"}]
         delete_it
       }
       filter {
         has [{:or => "subject:(#{SPAM_SUBJECTS.join("|")})"}]
         delete_it
       }
     end.generate)
```

Not the most readable configuration I would say. Note: You have also to make
sure to quote correctly terms when they contain spaces.

So what about this one?

```yaml
version: v1alpha1
consts:
  friends:
    values:
      - pippo@gmail.com
      - pippo@hotmail.com
rules:
  - filters:
      from:
        - interesting@maillist.com
      consts:
        not:
          from:
            - friends
    actions:
      archive: true
      markRead: true
```

[↩](#a2)
