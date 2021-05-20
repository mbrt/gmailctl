#!/bin/bash

# Wrapper arround xdg-open for snap package.
# xdg-open is the only way to escape the containerized
# snap environment to use the user's favorite editor.

xdg-open "$@"

# Because xdg-open returns as soon as the editor is open
# we need to wait here for the user to close the editor.
read  -n 1 -p "Once finished editing, press any key to continue..."
