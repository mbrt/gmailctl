// NOTE: This is a simple example.
// Please refer to https://github.com/mbrt/gmailctl#configuration for docs about
// the config format. Don't forget to change the configuration before to apply it
// to your own inbox!

// Import the standard library
local lib = import 'gmailctl.libsonnet';

// Some useful variables on top
// TODO: Put your email here
local me = 'YOUR.EMAIL@gmail.com';
local toMe = { to: me };


// The actual configuration
{
  // Mandatory header
  version: 'v1alpha3',
  author: {
    name: 'YOUR NAME HERE',
    email: me,
  },

  // TODO: Use your own rules here
  rules: [
    {
      filter: toMe,
      actions: {
        markImportant: true,
      },
    },
    {
      filter: {
        from: 'bar@yahoo.com',
      },
      actions: {
        archive: true,
        labels: ['foo'],
      },
    },
  ],
}
