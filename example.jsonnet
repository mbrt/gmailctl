// Import the standard library
local lib = import 'gmailctl.libsonnet';

// Some useful variables on top
local me = 'pippo@gmail.com';
local spam = {
  or: [
    { from: 'spammer@spam.com' },
    { subject: 'foo bar baz' },
    { subject: 'I want to spam you' },
    { has: 'buy this' },
    { has: 'buy that' },
  ],
};

// The actual configuration
{
  // Mandatory header
  version: 'v1alpha3',
  author: {
    name: 'Pippo Pluto',
    email: me,
  },

  // The list of Gmail filter rules.
  rules: [
           {
             filter: {
               from: 'myalarm@myalarm.com',
             },
             actions: {
               markImportant: true,
               labels: ['alarm'],
             },
           },
           {
             filter: spam,
             actions: {
               delete: true,
             },
           },
         ]

         // Chained filters. These are applied in order: the first that
         // matches stops the evaluation of the following ones.
         //
         // For example, if an email is directed to me and is coming from
         // the list 'foobar@list.com', following the chain, it will be
         // marked as important, but the 'mylist' label will _not_ be added.
         + lib.chainFilters([
           {
             filter: {
               to: me,
             },
             actions: {
               markImportant: true,
             },
           },
           // else if...
           {
             filter: {
               cc: me,
             },
             actions: {
               archive: true,
             },
           },
           // else if...
           {
             filter: {
               list: 'foobar@list.com',
             },
             actions: {
               labels: ['mylist'],
             },
           },
         ]),
}
