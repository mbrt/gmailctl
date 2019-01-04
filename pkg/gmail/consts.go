package gmail

// Categories supported by Gmail.
const (
	CategoryPersonal   Category = "personal"
	CategorySocial              = "social"
	CategoryUpdates             = "updates"
	CategoryForums              = "forums"
	CategoryPromotions          = "promotions"
)

// Category is one of the smart categories in Gmail.
type Category string

// PossibleCategoryValues returns the list of possible values Category can assume.
//
// Keep in sync with the categories.
func PossibleCategoryValues() []string {
	return []string{
		string(CategoryPersonal),
		string(CategorySocial),
		string(CategoryUpdates),
		string(CategoryForums),
		string(CategoryPromotions),
	}
}
