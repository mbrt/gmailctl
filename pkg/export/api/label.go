package api

const (
	labelIDInbox     = "INBOX"
	labelIDTrash     = "TRASH"
	labelIDImportant = "IMPORTANT"
	labelIDUnread    = "UNREAD"

	labelIDCategoryPersonal   = "CATEGORY_PERSONAL"
	labelIDCategorySocial     = "CATEGORY_SOCIAL"
	labelIDCategoryUpdates    = "CATEGORY_UPDATES"
	labelIDCategoryForums     = "CATEGORY_FORUMS"
	labelIDCategoryPromotions = "CATEGORY_PROMOTIONS"
)

// LabelMap maps label names and IDs together.
type LabelMap interface {
	// NameToID maps the name of a label to its ID.
	NameToID(name string) (string, bool)
	// IDToName maps the id of a string to its name.
	IDToName(id string) (string, bool)
}

// DefaultLabelMap implements the LabelMap interface with a regular map of strings
type DefaultLabelMap struct {
	ntid map[string]string
	idtn map[string]string
}

// NewDefaultLabelMap creates a new DefaultLabelMap, given mapping between IDs to label names.
func NewDefaultLabelMap(idNameMap map[string]string) DefaultLabelMap {
	nameIDMap := map[string]string{}
	for id, name := range idNameMap {
		nameIDMap[name] = id
	}
	return DefaultLabelMap{
		ntid: nameIDMap,
		idtn: idNameMap,
	}
}

// NameToID maps the name of a label to its ID.
func (m DefaultLabelMap) NameToID(name string) (string, bool) {
	id, ok := m.ntid[name]
	return id, ok
}

// IDToName maps the id of a string to its name.
func (m DefaultLabelMap) IDToName(id string) (string, bool) {
	name, ok := m.idtn[id]
	return name, ok
}

// AddLabel adds a label to the mapping
func (m DefaultLabelMap) AddLabel(id, name string) {
	m.ntid[name] = id
	m.idtn[id] = name
}
