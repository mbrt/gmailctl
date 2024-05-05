# gmailctl
[![Go Report Card](https://goreportcard.com/badge/github.com/mbrt/gmailctl)](https://goreportcard.com/report/github.com/mbrt/gmailctl)
![Go](https://github.com/mbrt/gmailctl/workflows/Go/badge.svg)

This utility helps you generate and maintain Gmail filters in a declarative way.
It has a [Jsonnet](https://jsonnet.org/) configuration file that aims to be
simpler to write and maintain than using the Gmail web interface, to categorize,
label, archive and manage your inbox automatically.

## Table of contents
- [gmailctl](#gmailctl)
  - [Table of contents](#table-of-contents)
  - [Motivation](#motivation)
  - [Install](#install)
  - [Usage](#usage)
    - [Migrate from another solution](#migrate-from-another-solution)
    - [Other commands](#other-commands)
  - [Configuration](#configuration)
    - [Search operators](#search-operators)
    - [Logic operators](#logic-operators)
    - [Reusing filters](#reusing-filters)
    - [Actions](#actions)
    - [Labels](#labels)
    - [Tests](#tests)
  - [Tips and tricks](#tips-and-tricks)
    - [Chain filtering](#chain-filtering)
    - [To me](#to-me)
    - [Directly to me](#directly-to-me)
    - [Automatic labels](#automatic-labels)
    - [Multiple Gmail accounts](#multiple-gmail-accounts)
  - [Known issues](#known-issues)
    - [Apply filters to existing emails](#apply-filters-to-existing-emails)
    - [OAuth2 authentication errors](#oauth2-authentication-errors)
    - [YAML config is unsupported](#yaml-config-is-unsupported)
  - [Comparison with existing projects](#comparison-with-existing-projects)
  - [Footnotes](#footnotes)

## Motivation

If you use Gmail and have to maintain (like me) a lot of filters (to apply
labels, get rid of spam or categorize your emails), then you probably have (like
me) a very long list of messy filters. At a certain point one of your messages
got mislabled and you try to understand why. You scroll through that horrible
mess of filters, you wish you could find-and-replace stuff, test the changes on
your filters before applying them, refactor some filters together... in a way
treat them like you treat your code!

Gmail allows one to import and export filters in XML format. This can be used to
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

gmailctl is written in Go and requires a recent version (see [go.mod](go.mod)).
Make sure to setup your [`$GOPATH`](https://golang.org/doc/code.html#GOPATH)
correctly and include its `bin` subdirectory in your `$PATH`.

```
go install github.com/mbrt/gmailctl/cmd/gmailctl@latest
```

Alternatively, if you're on macOS, you can install easily via Homebrew or Macports:

```
# Install with Homebrew
brew install gmailctl
```
```
# Install with Macports
sudo port install gmailctl
```

On Fedora Linux, you can install from the official repositories:

```
sudo dnf install gmailctl
```

You can also choose to install the snap:

```
sudo snap install gmailctl
```

If so, make sure to configure xdg-mime to open the config file with your favorite
editor. For example, if you'd like to use `vim`:

```
xdg-mime default vim.desktop text/x-csrc
```

Once installed, run the init process:

```
gmailctl init
```

This will guide you through setting up the Gmail APIs and update your
settings without leaving your command line.

## Usage

[![asciicast](https://asciinema.org/a/1NIWhzeJNcrN7cCe7mGjWQQnx.svg)](https://asciinema.org/a/1NIWhzeJNcrN7cCe7mGjWQQnx)

The easiest way to use gmailctl is to run `gmailctl edit`. This will open the
local `.gmailctl/config.jsonnet` file in your editor. After you exit the editor
the configuration is applied to Gmail. See [Configuration](#configuration) for
the configuration file format. This is the preferred way if you want to start
your filters from scratch.

**NOTE:** It's recommended to backup your current configuration before you apply
the generated one for the first time. Your current filters will be wiped and
replaced with the ones specified in the config file. The diff you'll get during
the first run will probably be pretty big, but from that point on, all changes
should generate a small and simple to review diff.

### Migrate from another solution

If you want to preserve your current filters and migrate to a more sane
configuration gradually, you can try to use the `download` command. This will
look up at your currently configured filters in Gmail and try to create a
configuration file matching the current state.

**NOTE:** This functionality is experimental. It's recommended to download the
filters and check that they correspond to the remote ones before making any
changes, to avoid surprises. Also note that the configuration file will be quite
ugly, as expressions won't be reconstructed properly, but it should serve as a
starting point if you are migrating from other systems.

Example of usage:

```bash
# download the filters to the default configuration file
gmailctl download > ~/.gmailctl/config.jsonnet
# check that the diff is empty and no errors are present
gmailctl diff
# happy editing!
gmailctl edit
```

Often you'll see imported filters with the `isEscaped: true` marker. This tells
gmailctl to not escape or quote the expression, as it might contain operators
that have to be interpreted as-is by Gmail. This happens when the `download`
command was unable to map the filter to native gmailctl expressions. It's
recommended to manually port the filter to regular gmailctl operators before
doing any changes, to avoid unexpected results. Example of such conversion:

```jsonnet
{
  from: "{foo bar baz}",
  isEscaped: true,
}
```

Can be translated into:

```jsonnet
{
  or: [
    {from: "foo"},
    {from: "bar"},
    {from: "baz"},
  ],
}
```

### Other commands

All the available commands (you can also check with `gmailctl help`):

```
  apply       Apply a configuration file to Gmail settings
  debug       Shows an annotated version of the configuration
  diff        Shows a diff between the local configuration and Gmail settings
  download    Download filters from Gmail to a local config file
  edit        Edit the configuration and apply it to Gmail
  export      Export filters into the Gmail XML format
  help        Help about any command
  init        Initialize the Gmail configuration
  test        Execute config tests
```

## Configuration

**NOTE:** Despite the name, the configuration format is stable at `v1alpha3`.
If you are looking for the deprecated versions `v1alpha1`, or `v1alpha2`,
please refer to [docs/v1alpha1.md](docs/v1alpha1.md) and
[docs/v1alpha2.md](docs/v1alpha2.md).

The configuration file is written in Jsonnet, that is a very powerful
configuration language, derived from JSON. It adds functionality such as
comments, variables, references, arithmetic and logic operations, functions,
conditionals, importing other files, parameterizations and so on. For more
details on the language, please refer to [the official
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
  version: 'v1alpha3',
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
* `bcc`: the mail has the given address as BCC destination
* `replyto`: the mail has the given address as Reply-To destination

One more special function is given if you need to use less common operators<sup
id="a1">[1](#f1)</sup>, or want to compose your query manually:

* `query`: passes the given contents verbatim to the Gmail filter, without
  escaping or interpreting the contents in any way.

Example:

```jsonnet
{
  version: 'v1alpha3',
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
  version: 'v1alpha3',
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
  version: 'v1alpha3',
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
  version: 'v1alpha3',
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
Relying on Jsonnet also allows [importing code and raw data](https://jsonnet.org/learning/tutorial.html#imports) from other files<sup id="a3">[3](#f3)</sup>.

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
  allows only one label per filter).
* `forward: 'forward@to.com'`: forward the message to another email address. The
  forwarding address must be already in your settings (Forwarding and POP/IMAP >
  Add a forwarding address). Gmail allows no more than 20 forwarding filters.
  Only one address can be specified for one filter.

Example:

```jsonnet
{
  version: 'v1alpha3',
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

### Labels

You can optionally manage your labels with gmailctl. The config contains a
`labels` section. Adding labels in there will opt you in to full label
management as well. If you prefer to manage your labels through the GMail web
interface, you can by all means still do so by simply omitting the `labels`
section from the config.

Example:

```jsonnet
{
  version: 'v1alpha3',
  // optional
  labels: [
    { name: 'family' },
    { name: 'friends' },
  ],
  rules: [
    {
      filter: { from: 'love@gmail.com' },
      actions: {
        labels: ['family'],
      },
    },
  ],
}
```

To make this work, your credentials need to contain permissions for labels
management as well. If you configured gmailctl before this functionality was
available, you probably need to update your 'Scopes for Google API' in the
'OAuth content screen' by adding `https://www.googleapis.com/auth/gmail.labels`.
If you don't know how to do this, just reset and re-create your credentials
following the steps in:

```
$ gmailctl init --reset
$ gmailctl init
```

If you want to update your existing config to include your existing labels, the
best way to get started is to use the `download` command and copy paste the
`labels` field into your config:

```
$ gmailctl download > /tmp/cfg.jsonnet
$ gmailctl edit
```

After the import, verify that your current config does not contain unwanted
changes with `gmailctl diff`.

Managing the color of a label is optional. If you specify it, it will be
enforced; if you don't, the existing color will be left intact. This is useful
to people who want to keep setting the colors with the Gmail UI. You can find
the list of supported colors
[here](https://developers.google.com/gmail/api/v1/reference/users/labels).

Example:

```jsonnet
{
  version: 'v1alpha3',
  labels: [
    {
      name: 'family',
      color: {
        background: "#fad165",
        text: "#000000",
      },
    },
  ],
  rules: [ // ...
  ],
}
```

Note that renaming labels is not supported because there's no way to tell the
difference between a rename and a deletion. This distinction is important
because deleting a label and creating it with a new name would remove it from
all the messages. This is a surprising behavior for some users, so it's
currently gated by a confirmation prompt (for the `edit` command), or by the
`--remove-labels` flag (for the `apply` command). If you want to rename a label,
please do so through the GMail interface and then change your gmailctl config.

### Tests

You can optionally add unit tests to your configuration. The tests will be
executed before applying any changes to the upstream Gmail filters or by running
the dedicated `test` subcommand. Tests results can be ignored by passing the
`--yolo` command line option.

Tests can be added by using the `tests` field of the main configuration object:

```jsonnet
{
  version: 'v1alpha3',
  rules: [ /* ... */ ],
  tests: [
    // you tests here.
  ],
}
```

A test object looks like this:

```jsonnet
{
  // Reported when the test fails.
  name: "the name of the test",
  // A list of messages to test against.
  messages: [
    { /* message object */ },
    // ... more messages
  ],
  // The actions that should be applied to the messages, according the config.
  actions: {
    // Same as the Actions object in the filters.
  },
}
```

A message object is similar to a filter, but it doesn't allow arbitrary
expressions, uses arrays of strings for certain fields (e.g. the `to` field),
and has some additional fields (like `body`) to represent an email as faithfully
as possible. This is the list of fields:

* `from: <string>`: the sender of the email.
* `to: [<list>]`: a list of recipients of the email.
* `cc: [<list>]`: a list of emails in cc.
* `bcc: [<list>]`: a list of emails in bcc.
* `replyto: <string>`: the email listed in the Reply-To field.
* `lists: [<list>]`: a list of mailing lists.
* `subject: <string>`: the subject of the email.
* `body: <string>`: the body of the email.

All the fields are optional. Remember that each message object represent one
email and that the `messages` field of a test is an array of messages. A common
mistake is to provide an array of messages thinking that they are only one.
Example:

```jsonnet
{
  // ...
  tests: [
    messages: [
      { from: "foobar" },
      { to: ["me"] },
    ],
    actions: {
      // ...
    },
  ],
}
```

This doesn't represent one message from "foobar" to "me", but two messages, one
from "foobar" and the other to "me". The correct representation for that would
be instead:

```jsonnet
{
  // ...
  tests: [
    messages: [
      {
        from: "foobar",
        to: "me",
      },
    ],
    actions: {
      // ...
    },
  ],
}
```

**NOTE:** Not all filters are supported in tests. Arbitrary `query` expressions
and filters with `isEscaped: true` are ignored by the tests. Warnings are
generated when this happens. Keep in mind that in that case your tests might
yield incorrect results.

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
* or if it's directed to a certain alias, archive it.

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
  version: 'v1alpha3',
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
  version: 'v1alpha3',
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
    { not: { bcc: recipient } },
  ],
};
```

So, from all emails where your mail is a recipient, we remove the ones where
your mail is in the CC field.

This trick is conveniently provided by the `gmailctl` library, so you can use it
for example in this way:

```jsonnet
// Import the standard library
local lib = import 'gmailctl.libsonnet';
local me = 'pippo@gmail.com';
{
  version: 'v1alpha3',
  rules: [
    {
      filter: lib.directlyTo(me),
      actions: { markImportant: true },
    },
  ],
}
```

### Automatic labels

If you opted in for labels management, you will find yourself often having to
both add a filter and a label to your config. To alleviate this problem, you can
use the utility function `lib.rulesLabels` provided with the gmailctl standard
library. With that you can avoid providing the labels referenced by filters.
They will be automatically added to the list of labels.

Example:

```jsonnet
local lib = import 'gmailctl.libsonnet';
local rules = [
  {
    filter: { to: 'myself@gmail.com' },
    actions: { labels: ['directed'] },
  },
  {
    filter: { from: 'foobar' },
    actions: { labels: ['lists/foobar'] },
  },
  {
    filter: { list: 'baz' },
    actions: { labels: ['lists/baz', 'wow'] },
  },
];

// the config
{
  version: 'v1alpha3',
  rules: rules,
  labels: lib.rulesLabels(rules) + [{ name: l } for l in [
    'manual-label1',
    'priority',
    'priority/p1',
  ]],
}
```

The resulting list of labels will be:

```jsonnet
labels: [
  // Automatically added
  { name: 'directed' },
  { name: 'lists' },            // Implied parent label
  { name: 'lists/baz' },
  { name: 'lists/foobar' },
  { name: 'wow' },
  // Manually added
  { name: 'manual-label1' },
  { name: 'priority' },
  { name: 'priority/p1' },
]
```

Note that there's no need to specify the label `lists`, because even if it's not
used in any filter, it's the parent of a label that is used.

Things to keep in mind / gotchas:

* Removing the last filter referencing a label will delete the label.
* The only thing managed by the function is the list of labels names. You need
  to apply some transformations yourself if you want other properties (e.g. the
  color).
* If you have labels that are not referenced by any filters (maybe archive
  labels, or labels applied manually). You have to remember to specify them
  manually in the list.

Thanks to [legeana](https://github.com/legeana) for the idea!

### Multiple Gmail accounts

If you need to manage two or more accounts, it's useful to setup bash aliases
this way:

```bash
alias gmailctlu1='gmailctl --config=$HOME/.gmailctlu1'
alias gmailctlu2='gmailctl --config=$HOME/.gmailctlu2'
```

You will then be able to configure both accounts separately by using one or
the other alias.

## Known issues

### Apply filters to existing emails

gmailctl doesn't support this functionality for security reasons. The project
currently needs only very basic permissisons, and applying filters to existing
emails requires full Gmail access. Bugs in gmailctl or in your configuration
won't screw up your old emails in any way, so this is an important safety
feature. If you really want to do this, you can manually export your rules with
`gmailctl export > filters.xml`, upload them by using the Gmail Settings UI and
select the "apply new filters to existing email" checkbox.

### OAuth2 authentication errors

Gmail APIs require strict controls, even if you are only accessing your own
data. If you're getting errors similar to:

```
oauth2: cannot fetch token: 400 Bad Request
Response: {
  "error": "invalid_grant",
  "error_description": "Bad Request"
}
```

it's likely your auth token expired. Try refreshing it with:

```bash
$ gmailctl init --refresh-expired
```

and follow the instructions on screen.

If this doesn't help, retry the authorization workflow from the start:

```bash
$ gmailctl init --reset
$ gmailctl init
```

### YAML config is unsupported

gmailctl recently deprecated older config versions (`v1alpha1`, `v1alpha2`).
There's however a migration tool to port those into the latest Jsonnet format.
To convert your config:

```bash
$ go run github.com/mbrt/gmailctl/cmd/gmailctl-config-migrate \
    ~/.gmailct/config.yaml > /tmp/gmailctl-config.jsonnet
```

**Note:** Adjust your paths if you're not keeping your config file in the
default directory.

Confirm that the new config file doesn't have errors, nor shows diffs with your
remote filters.

```bash
$ gmailctl diff -f /tmp/gmailctl-config.jsonnet
```

If everything looks good, replace the old with the new config:

```bash
$ mv /tmp/gmailctl-config.jsonnet ~/.gmailctl/config.jsonnet
$ rm ~/.gmailctl/config.yaml
```

## Comparison with existing projects

[gmail-britta](https://github.com/antifuchs/gmail-britta) has similar
motivations and is quite popular. The difference between that project and
this one are:

* `gmail-britta` uses a custom DSL (versus Jsonnet in `gmailctl`)
* `gmail-britta` is imperative because it allows you to write arbitrary Ruby
  code in your filters (versus pure declarative for `gmailctl`)
* `gmail-britta` allows one to write complex chains of filters, but they feel
  very hardcoded and fails to provide easy ways to write reasonably easy filters
  <sup id="a2">[2](#f2)</sup>.
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

<b id="f2">2</b>: Try to write the equivalent of this filter with `gmail-britta` [↩](#a2)

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
  version: 'v1alpha3',
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
  version: 'v1alpha3',
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

<b id="f3">3</b>: Import variables from  a `.libjsonnet` file [↩](#a3)

File: `spam.libjsonnet`
```jsonnet
{
  or: [
    { from: 'foo@gmail.com' },
    { from: 'bar@hotmail.com' },
    { subject: 'buy this' },
    { subject: 'buy that' },
  ],
}
```

File `config.jsonnet`
```jsonnet
local spam_filter = import 'spam.libjsonnet';
{
  version: 'v1alpha3',
  rules: [
    {
      filter: spam_filter,
      actions: { delete: true },
    },
  ],
}
```
