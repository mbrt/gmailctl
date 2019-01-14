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
2. An easy to understand, declarative language;
3. Ability to review your changes before applying them;
4. Quickly update your settings (no minutes waiting for an import for every
   little change).

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

The YAML configuration file contains two important sections:

* `filters` that contains named filters that can be called up by subsequent
  filters or rules.
* `rules` that specify a filter expression and a set of actions that will be
  applied if the filter matches.

TODO

## Comparison with existing projects

[gmail-britta](https://github.com/antifuchs/gmail-britta) has similar
motivations and is quite popular. The difference between that project and
this one are:

* `gmail-britta` uses a custom DSL (versus YAML in `gmailctl`)
* `gmail-britta` is imperative because it allows you to write arbitrary Ruby
  code in your filters (versus pure declarative for `gmailctl`)
* `gmail-britta` allows to write complex chains of filters, but fails to provide
  easy ways to write reasonably easy filters [1](#footnote-1).
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

### Footnote 1

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
