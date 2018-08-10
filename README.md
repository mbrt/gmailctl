# gmailfilter
[![Go Report Card](https://goreportcard.com/badge/github.com/mbrt/gmailfilter)](https://goreportcard.com/report/github.com/mbrt/gmailfilter)
[![Build Status](https://travis-ci.org/mbrt/gmailfilter.svg?branch=master)](https://travis-ci.org/mbrt/gmailfilter)

This utility helps you generate Gmail filters. It has a Yaml configuration file
that aims to be more simple to write and maintain than using the Gmail web
interface.

## Motivation

If you have Gmail and have (like me) have to maintain a lot of filters to apply
labels, get rid of spam or categorize your emails, then you probably have (like
me) a mess of filters that you don't even remember what they are intended to do
and why.

Gmail has an import export functionality for filters. The format used is XML.
Every entry of this file maps maps 1:1 to a rule in Gmail. Editing these files
is also quite painful because what you see in it doesn't map exactly with what
you see in the web interface. What's more is that Gmail provides an undocumented
and powerful expression language that goes beyond what you see in the Search
form.

This project exists to combine:
1. maintainability
2. powerful but declarative language

## Usage

```
go install github.com/mbrt/gmailfilter/cmd/gmailfilter
gmailfilter config.yaml > filters.xml
```

where `config.yaml` is the configuration file containing the filtering rules
(see [Configuration](#configuration)). The utility will print out an XML that you
can import as Gmail filters.

**NOTE:** It's recommended to backup your current configuration before to
applying a generated one. Bugs can happen in both code and configuration. Always
backup to avoid surprises.

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

1. If the filter of a rule matches, then it's applied. This means that every
   rule is in OR with all the others. In the example, given an email, if the
   first filter matches, then its actions are applied; if the second also
   matches, then its actions are also applied.
2. Within a rule, all the filters have to match, in order for the rule to match.
   This means that the filters inside a rule are in AND together. In the
   previous example, if only `filterA` matches, then the first rule is not
   applied. If both `filterA` and `filterB` match, then the rule also matches.

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
replaced by the constants.

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

## Comparison with exising projects

[gmail-britta](https://github.com/antifuchs/gmail-britta) has similar
motivations and is quite popular. The difference between that project and
this one are:

* `gmail-britta` uses a custom DSL (versus YAML in `gmailfilter`)
* `gmail-britta` is imperative because it allows you to write arbitrary Ruby
  code in your filters (versus pure declarative for `gmailfilter`)
* `gmail-britta` allows to write complex chains of filters, but fails to provide
  easy ways to write reasonably easy filters [1](#footnote-1).
* `gmailfilter` tries to workaround certain limitations in Gmail (like applying
  multiple labels with the same filter) `gmail-britta` tries to workaround
  others (chain filtering).
* chain filtering is not supported in `gmailfilter` by design. The declarative
  nature of the configuration makes it that every rule that matches is applyied,
  just like Gmail does.

In short `gmailfilter` takes the declarative approach to Gmail filters
configuration, hoping it stays simpler to read and maintain, sacrificing complex
scenarios handled instead by `gmail-britta` (like chaining).
  
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
