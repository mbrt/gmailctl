# gmailctl
[![Go Report Card](https://goreportcard.com/badge/github.com/mbrt/gmailctl)](https://goreportcard.com/report/github.com/mbrt/gmailctl)
[![Build Status](https://travis-ci.org/mbrt/gmailctl.svg?branch=master)](https://travis-ci.org/mbrt/gmailctl)

This utility helps you generate and maintain Gmail filters in a declarative way.
It has a Yaml configuration file that aims to be more simple to write and
maintain than using the Gmail web interface, to categorize, label, archive and
manage automatically your inbox.

## Motivation

If you have Gmail and have (like me) have to maintain a lot of filters to apply
labels, get rid of spam or categorize your emails, then you probably have (like
me) a mess of filters that you don't even remember what they are intended to do
and why.

Gmail has both an import export functionality for filters (in XML format) and
powerful APIs that we can access to update our settings. Editing these XML files
is pretty painful because the format is not human friendly. The Gmail query
language is quite powerful but not easy to remember.

This project exists to combine:
1. maintainability
2. powerful but declarative language
3. quick updates to your settings

## Usage

[![asciicast](https://asciinema.org/a/BU00aXIcZV9bYWRko7i7LnugY.png)](https://asciinema.org/a/BU00aXIcZV9bYWRko7i7LnugY)

Make sure to setup your [$GOPATH](https://golang.org/doc/code.html#GOPATH) correctly, including the `bin` subdirectory in your `$PATH`.

```
go get github.com/mbrt/gmailctl/cmd/gmailctl
go install github.com/mbrt/gmailctl/cmd/gmailctl
gmailctl init
# edit the config file
gmailctl apply -f config.yaml
```

where `config.yaml` is the configuration file containing the filtering rules
(see [Configuration](#configuration)). The utility will guide you through
setting up the Gmail APIs and update your settings without leaving your command
line.

**NOTE:** It's recommended to backup your current configuration before to
applying a generated one. Your current filters will be wiped and replaced with
the ones specified in the config file. Since bugs can happen in both code and
configuration, always backup to avoid surprises.

## Configuration

**NOTE:** The configuration format is still in alpha and might change in the
future.

The configuration contains two important sections:

* `consts` that contains global constants that can be referenced later on by
  rules.
* `rules` that specify a set of filters that when match cause a set of actions
  to happen.

### Rule evaluation

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
3. Within filter the listed values are in OR with each other. In the second rule, 
   `filterC` matches if either `valueD` or `valueE` are present.

### Filters
The following simple filters are available:
* from
* to
* subject
* has

You can apply the special `not` operator to negate a match in this way:

```yaml
  - filters:
      not:
        to:
          - foo@bar.com
        subject:
          - Baz zorg
```

The rule will match if the email both is not directed to `foo@bar.com` and does
not contain `Baz zorg` in the subject.

### Constants
A filter can refer to global constants specified in the first section by using
the `consts` section inside the filter. All values inside the rule will be
replaced by the constants. Inside `consts` you can put again the same set of
filters of the positive case:
* from
* to
* subject
* has

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

### Example

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
