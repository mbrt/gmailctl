package main

import "fmt"

// False is also the value when settings.compact is omitted. Neither upstream
// state nor filter count can change the user's configured choice.
func chooseToggle(compact bool) decision {
	return decision{compact, fmt.Sprintf("settings.compact is %t", compact)}
}
