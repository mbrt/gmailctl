package reporting

import (
	"encoding/json"
	"fmt"
)

// Prettify returns a readable string representing the object.
func Prettify(o interface{}, compact bool) string {
	var (
		b   []byte
		err error
	)
	if compact {
		b, err = json.Marshal(o)
	} else {
		b, err = json.MarshalIndent(o, "", "  ")
	}
	if err != nil {
		return fmt.Sprintf("(invalid) %+v", o)
	}
	return string(b)
}
