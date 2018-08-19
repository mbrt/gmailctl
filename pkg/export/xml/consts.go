package xml

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mbrt/gmailfilter/pkg/config"
)

// Property values
const (
	PropertyFrom          = "from"
	PropertyTo            = "to"
	PropertySubject       = "subject"
	PropertyHas           = "hasTheWord"
	PropertyMarkImportant = "shouldAlwaysMarkAsImportant"
	PropertyApplyLabel    = "label"
	PropertyApplyCategory = "smartLabelToApply"
	PropertyDelete        = "shouldTrash"
	PropertyArchive       = "shouldArchive"
	PropertyMarkRead      = "shouldMarkAsRead"
)

// SmartLabel values
const (
	SmartLabelPersonal     = "personal"
	SmartLabelGroup        = "group"
	SmartLabelNotification = "notification"
	SmartLabelPromo        = "promo"
	SmartLabelSocial       = "social"
)

func categoryToSmartLabel(cat config.Category) (string, error) {
	var smartl string
	switch cat {
	case config.CategoryPersonal:
		smartl = SmartLabelPersonal
	case config.CategorySocial:
		smartl = SmartLabelSocial
	case config.CategoryUpdates:
		smartl = SmartLabelNotification
	case config.CategoryForums:
		smartl = SmartLabelGroup
	case config.CategoryPromotions:
		smartl = SmartLabelPromo
	default:
		possib := config.PossibleCategoryValues()
		return "", errors.Errorf("unrecognized category '%s' (possible values: %s)",
			cat, strings.Join(possib, ", "))
	}
	return fmt.Sprintf("^smartlabel_%s", smartl), nil
}
