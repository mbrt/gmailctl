# gmailctl
[![Go Report Card](https://goreportcard.com/badge/github.com/mbrt/gmailctl)](https://goreportcard.com/report/github.com/mbrt/gmailctl)
[![Build Status](https://travis-ci.org/mbrt/gmailctl.svg?branch=master)](https://travis-ci.org/mbrt/gmailctl)

This utility helps you generate and maintain Gmail filters in a declarative way.
It has a [Jsonnet](https://jsonnet.org/) configuration file that aims to be more
simple to write and maintain than using the Gmail web interface, to categorize,
label, archive and manage automatically your inbox.

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

## Install

Make sure to setup your [$GOPATH](https://golang.org/doc/code.html#GOPATH) correctly, including the `bin` subdirectory in your `$PATH`.

```
go get github.com/mbrt/gmailctl/cmd/gmailctl
go install github.com/mbrt/gmailctl/cmd/gmailctl
gmailctl init
```

The init will guide you through setting up the Gmail APIs and update your
settings without leaving your command line.

## Usage

[![asciicast](https://asciinema.org/a/1NIWhzeJNcrN7cCe7mGjWQQnx.svg)](https://asciinema.org/a/1NIWhzeJNcrN7cCe7mGjWQQnx)

The easiest way to use gmailctl is to run `gmailctl edit`. This will open the local `.gmailctl/config.jsonnet` file in your editor. After you exit the editor the configuration is applied to Gmail.  See [Configuration](#configuration) for the configuration file format.

**NOTE:** It's recommended to backup your current configuration before you apply
the generated one for the first time. Your current filters will be wiped and
replaced with the ones specified in the config file. The diff you'll get during
the first run will probably be pretty big, but from that point on, all changes
should generate a small and simple to review diff.

Other available commands:

```
  apply       Apply a configuration file to Gmail settings
  debug       Shows an annotated version of the configuration
  diff        Shows a diff between the local configuaration and Gmail settings
  edit        Edit the configuration and apply it to Gmail
  export      Export filters into the Gmail XML format
  help        Help about any command
  init        Initialize the Gmail configuration
```


## Configuration

**NOTE:** The configuration format is still in alpha and might change in the
future. If you are looking for the deprecated version `v1alpha1`, please refer
to [docs/v1alpha1.md](docs/v1alpha1.md).

For the configuration file, both YAML and Jsonnet are supported. The YAML format
is kept for retro-compatibility, it can be more readable but also much less
flexible. The Jsonnet version is very powerful and also comes with a utility
library that helps you write some more complex filters.

For the documentation on the YAML version, please refer to
[docs/v1alpha2-yaml.md](docs/v1alpha2-yaml.md).

Jsonnet is a very powerful configuration language, derived from JSON, adding
functionality such as comments, variables, references, arithmetic and logic
operations, functions, conditionals, importing other files, parametrizations and
so on. For more details on the language, please refer to [the official
tutorial](https://jsonnet.org/learning/tutorial.html).

Simple example:

```jsonnet
// Local variables help reuse config fragments
local me = {
  or: [
    { to: 'pippo@gmail.com' },
    { to: 'pippo@hotmail.com' },
  ],
};

// The exported configuration starts here
{
  version: 'v1alpha2',
  // Optional author information (used in exports).
  author: {
    name: 'Pippo Pluto',
    email: 'pippo@gmail.com'
  },
  rules: [
    {
      filter: {
        and: [
          { list: 'geeks@newsletter.com' },
          { not: me },  // Reference to the local variable 'me'
        ],
      },
      actions: {
        archive: true,
        labels: ['news'],
      },
    },
  ],
}
```

The Jsonnet configuration file contains mandatory version information, optional
author metadata and a list of rules. Rules specify a filter expression and a set
of actions that will be applied if the filter matches.

Filter operators are prefix of the operands they apply to. In the example above,
the filter applies to emails that come from the mail list 'geeks@newsletter.com'
AND the recipient is not 'me' (which can be 'pippo@gmail.com' OR
'pippo@hotmail.com').

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

```jsonnet
{
  version: 'v1alpha2',
  rules: [
    {
      filter: { subject: 'important mail' },
      actions: {
        markImportant: true,
      },
    },
    {
      filter: {
        query: 'dinner AROUND 5 friday has:spreadsheet',
      },
      actions: {
        delete: true,
      },
    },
  ],
}
```

### Logic operators

Filters can contain only one expression. If you want to combine multiple of them
in the same rule, you have to use logic operators (and, or, not). These
operators do what you expect:

* `and`: is true only if all the sub-expressions are also true
* `or`: is true if one or more sub-expressions are true
* `not`: is true if the sub-expression is false.

Example:

```jsonnet
{
  version: 'v1alpha2',
  rules: [
    {
      filter: {
        or: [
          { from: 'foo' },
          {
            and: [
              { list: 'bar' },
              { not: { to: 'baz' } },
            ],
          },
        ],
      },
      actions: {
        markImportant: true,
      },
    },
  ],
}
```

This composite filter marks the incoming mail as important if:

* the message comes from "foo", _or_
* it is coming from the mailing list "bar" _and_ _not_ directed to "baz"

### Reusing filters

Filters can be named and referenced in other filters. This allows reusing
concepts and so avoid repetition. Note that this is not a gmailctl functionality
but comes directly from the fact that we rely on Jsonnet.

Example:

```jsonnet
local toMe = {
  or: [
    { to: 'myself@gmail.com' },
    { to: 'myself@yahoo.com' },
  ],
};
local notToMe = { not: toMe };

{
  version: 'v1alpha2',
  rules: [
    {
      filter: {
        and: [
          { from: 'foobar' },
          notToMe,
        ],
      },
      actions: {
        delete: true,
      },
    },
    {
      filter: toMe,
      actions: {
        labels: ['directed'],
      },
    },
  ],
}
```

In this example, two named filters are defined. The `toMe` filter gives a name
to emails directed to 'myself@gmail.com' or to 'myself@yahoo.com'. The `notToMe`
filter negates the `toMe` filter, with a `not` operator. Similarly, the two
rules reference the two named filters above. The `name` reference is basically
copying the definition of the filter in place.

The example is effectively equivalent to this one:

```jsonnet
{
  version: 'v1alpha2',
  rules: [
    {
      filter: {
        and: [
          { from: 'foobar' },
          {
            not: {
              or: [
                { to: 'myself@gmail.com' },
                { to: 'myself@yahoo.com' },
              ],
            },
          },
        ],
      },
      actions: {
        delete: true,
      },
    },
    {
      filter: {
        or: [
          { to: 'myself@gmail.com' },
          { to: 'myself@yahoo.com' },
        ],
      },
      actions: {
        labels: ['directed'],
      },
    },
  ],
}
```

Or in YAML:

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
* `forward: 'forward@to.com'`: forward the message to another email address.
  The forwarding address must be already in your settings
  (Forwarding and POP/IMAP > Add a forwarding address). Gmail allows no more
  than 20 filters with a forward.

Example:

```jsonnet
{
  version: 'v1alpha2',
  rules: [
    {
      filter: { from: 'love@gmail.com' },
      actions: {
        markImportant: true,
        category: 'personal',
        labels: ['family', 'P1'],
      },
    },
  ],
}
```

## Tips and tricks

### Chain filtering

Gmail filters are _all_ applied to a mail, if they match, in a non-specified
order. So having some if-else alternative is pretty hard to encode by hand. For
example sometimes you get interesting stuff from a mail list, but also a lot of
garbage too. So, to put some emails with certain contents in one label and the
rest somewhere else, you'd have to make multiple filters. Gmail filters however
lack if-else constructs, so a way to simulate that is to declare a sequence of
filters, where each one negates the previous alternatives.

For example you want to:

* mark the email as important if directed to you;
* or if it's coming from a list of favourite addresses, label as interesting;
* of if it's directed to a certain alias, archive it.

Luckily you don't have to do that by hand, thanks to the utility library coming
with `gmailctl`. There's a `chainFilters` function that does exactly that: takes
a list of rules and chains them together, so if the first matches, the others
are not applied, otherwise the second is checked, and so on...

```jsonnet
// Import the standard library
local lib = import 'gmailctl.libsonnet';

local favourite = {
  or: [
    { from: 'foo@bar.com' },
    { from: 'baz@bar.com' },
    { list: 'wow@list.com' },
  ],
};

{
  version: 'v1alpha2',
  rules: [
           // ... Other filters applied in any order
         ]

         // And a chain of filters
         + lib.chainFilters([
           // All directed emails will be marked as important
           {
             filter: { to: 'myself@gmail.com' },
             actions: { markImportant: true },
           },
           // Otherwise, if they come from interesting senders, apply a label
           {
             filter: favourite,
             actions: { labels: ['interesting'] },
           },
           // Otherwise, if they are directed to my spam alias, archive
           {
             filter: { to: 'myself+spam@gmail.com' },
             actions: { archive: true },
           },
         ]),
}
```

This is equivalent to this YAML configuration:

```yaml
version: v1alpha2
rules:
- filter:
    to: myself@gmail.com
  actions:
    markImportant: true

- filter:
    and:
    - not:
        to: myself@gmail.com
    - or:
      - from: foo@bar.com
      - from: baz@bar.com
      - list: wow@list.com
  actions:
    labels:
    - interesting
  
- filter:
    and:
    - not:
        to: myself@gmail.com
    - not:
        or:
        - from: foo@bar.com
        - from: baz@bar.com
        - list: wow@list.com
    - to: myself+spam@gmail.com
  actions:
    archive: true
```

### To me

Gmail gives you the possibility to write literally `to:me` in a filter, to match
incoming emails where you are the recipient. This is going to mostly work as
intended, except that it will also match emails directed to `me@example.com`.
The risk you are getting an email where you are not one of the recipients, but a
`me@example.com` is, is pretty low, but if you are paranoid you might consider
using your full email instead. The config is also easier to read in my opinion.
You can also save some typing by introducing a local variable like this:

```jsonnet
// Local variable, referenced in all your config.
local me = 'myemail@gmail.com';

{
  version: 'v1alpha2',
  rules: [
    {
      // Save typing here.
      filter: { to: me },
      actions: {
        markImportant: true,
      },
    },
  ],
}
```

### Directly to me

If you need to match emails that are to you directly, (i.e. you are not in CC,
or BCC, but only in the TO field), then the default Gmail filter `to:
mymail@gmail.com` is not what you are looking for. This filter in fact
(surprisingly) matches all the recipient fields (TO, CC, BCC). To make this work
the intended way we have to pull out this trick:

```jsonnet
local directlyTo(recipient) = {
  and: [
    { to: recipient },
    { not: { cc: recipient } },
  ],
};
```

So, from all emails where your mail is a recipient, we remove the ones where
your mail is in the CC field. Note that we don't need to remove BCC emails,
because no mail matches that filter.

This trick is conveniently provided by the `gmailctl` library, so you can use it
for example in this way:

```jsonnet
// Import the standard library
local lib = import 'gmailctl.libsonnet';
local me = 'pippo@gmail.com';
{
  version: 'v1alpha2',
  rules: [
    {
      filter: lib.directlyTo(me),
      actions: { markImportant: true },
    },
  ],
}
```

### Multiple Gmail accounts

If you need to manage two or more accounts, it's useful to setup bash aliases
this way:

```bash
alias gmailctlu1='gmailctl --config=$HOME/.gmailctlu1'
alias gmailctlu2='gmailctl --config=$HOME/.gmailctlu2'
```

You will then be able to configure both accounts separately by using one or
the other alias.

## Comparison with existing projects

[gmail-britta](https://github.com/antifuchs/gmail-britta) has similar
motivations and is quite popular. The difference between that project and
this one are:

* `gmail-britta` uses a custom DSL (versus Jsonnet in `gmailctl`)
* `gmail-britta` is imperative because it allows you to write arbitrary Ruby
  code in your filters (versus pure declarative for `gmailctl`)
* `gmail-britta` allows to write complex chains of filters, but they feel very
  hardcoded and fails to provide easy ways to write reasonably easy filters <sup
  id="a2">[2](#f2)</sup>.
* `gmail-britta` exports only to the Gmail XML format. You have to import the
  filters yourself by using the Gmail web interface, manually delete the filters
  you updated and import only the new ones. This process becomes tedious very
  quickly and you will resort to quickly avoid using the tool when in a hurry.
  `gmailctl` provides you this possibility, but also allows you to review your
  changes and update the filters by using the Gmail APIs, without you having to
  do anything manually.
* `gmailctl` tries to workaround certain limitations in Gmail (like applying
  multiple labels with the same filter) and provide a generic query language to
  Gmail, `gmail-britta` focuses on writing chain filtering and archiving in very
  few lines.

In short `gmailctl` takes the declarative approach to Gmail filters
configuration, hoping it stays simpler to read and maintain, doesn't attempt to
simplify complex scenarios with shortcuts (again, hoping the configuration
becomes more readable) and provides automatic and fast updates to the filters
that will save you time while you are iterating through new versions of your
filters.

## Footnotes

<b id="f1">1</b>: See [Search operators you can use with
Gmail](https://support.google.com/mail/answer/7190?hl=en) [↩](#a1).

<b id="f2">2</b>:

Try to write the equivalent of this filter with `gmail-britta`:

```jsonnet
local spam = {
  or: [
    { from: 'pippo@gmail.com' },
    { from: 'pippo@hotmail.com' },
    { subject: 'buy this' },
    { subject: 'buy that' },
  ],
};
{
  version: 'v1alpha2',
  rules: [
    {
      filter: spam,
      actions: { delete: true },
    },
  ],
}
```

It becomes something like this:

```ruby
#!/usr/bin/env ruby

# NOTE: This file requires the latest master (30/07/2018) of gmail-britta.
# The Ruby repos are not up to date

require 'rubygems'
require 'gmail-britta'

SPAM_EMAILS = %w{foo@gmail.com bar@hotmail.com}
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

Not the most readable configuration I would say. Note: You also have to make
sure to quote the terms correctly when they contain spaces.

So what about nesting expressions?

```jsonnet
local me = 'pippo@gmail.com';
local spam = {
  or: [
    { from: 'foo@gmail.com' },
    { from: 'bar@hotmail.com' },
    { subject: 'buy this' },
    { subject: 'buy that' },
  ],
};
{
  version: 'v1alpha2',
  rules: [
    {
      filter: {
        and: [
          { to: me },
          { from: 'friend@mail.com' },
          { not: spam },
        ],
      },
      actions: { delete: true },
    },
  ],
}
```

The reality is that you have to manually build the Gmail expressions yourself.

[↩](#a2)
