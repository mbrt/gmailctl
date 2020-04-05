package xml

import (
	"fmt"
	"strings"

	"github.com/mbrt/gmailctl/pkg/gmail"
)

// Property values
const (
	PropertyFrom             = "from"
	PropertyTo               = "to"
	PropertySubject          = "subject"
	PropertyHas              = "hasTheWord"
	PropertyMarkImportant    = "shouldAlwaysMarkAsImportant"
	PropertyMarkNotImportant = "shouldNeverMarkAsImportant"
	PropertyApplyLabel       = "label"
	PropertyApplyCategory    = "smartLabelToApply"
	PropertyDelete           = "shouldTrash"
	PropertyArchive          = "shouldArchive"
	PropertyMarkRead         = "shouldMarkAsRead"
	PropertyMarkNotSpam      = "shouldNeverSpam"
	PropertyStar             = "shouldStar"
	PropertyForward          = "forwardTo"
)

// SmartLabel values
const (
	SmartLabelPersonal     = "personal"
	SmartLabelGroup        = "group"
	SmartLabelNotification = "notification"
	SmartLabelPromo        = "promo"
	SmartLabelSocial       = "social"
)

func categoryToSmartLabel(cat gmail.Category) (string, error) {
	var smartl string
	switch cat {
	case gmail.CategoryPersonal:
		smartl = SmartLabelPersonal
	case gmail.CategorySocial:
		smartl = SmartLabelSocial
	case gmail.CategoryUpdates:
		smartl = SmartLabelNotification
	case gmail.CategoryForums:
		smartl = SmartLabelGroup
	case gmail.CategoryPromotions:
		smartl = SmartLabelPromo
	default:
		possib := gmail.PossibleCategoryValues()
		return "", fmt.Errorf("unrecognized category %q (possible values: %s)",
			cat, strings.Join(possib, ", "))
	}
	return fmt.Sprintf("^smartlabel_%s", smartl), nil
}
