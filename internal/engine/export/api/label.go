package api

import "github.com/mbrt/gmailctl/internal/engine/label"

const (
	labelIDInbox     = "INBOX"
	labelIDTrash     = "TRASH"
	labelIDImportant = "IMPORTANT"
	labelIDUnread    = "UNREAD"
	labelIDSpam      = "SPAM"
	labelIDStar      = "STARRED"

	labelIDCategoryPersonal   = "CATEGORY_PERSONAL"
	labelIDCategorySocial     = "CATEGORY_SOCIAL"
	labelIDCategoryUpdates    = "CATEGORY_UPDATES"
	labelIDCategoryForums     = "CATEGORY_FORUMS"
	labelIDCategoryPromotions = "CATEGORY_PROMOTIONS"
)

// LabelMap maps label names and IDs together.
type LabelMap struct {
	ntid map[string]string
	idtn map[string]string
}

// NewLabelMap creates a new LabelMap given a list of labels.
func NewLabelMap(labels []label.Label) LabelMap {
	nameIDMap := map[string]string{}
	idNameMap := map[string]string{}

	for _, l := range labels {
		nameIDMap[l.Name] = l.ID
		idNameMap[l.ID] = l.Name
	}

	return LabelMap{
		ntid: nameIDMap,
		idtn: idNameMap,
	}
}

// NameToID maps the name of a label to its ID.
func (m LabelMap) NameToID(name string) (string, bool) {
	id, ok := m.ntid[name]
	return id, ok
}

// IDToName maps the id of a string to its name.
func (m LabelMap) IDToName(id string) (string, bool) {
	name, ok := m.idtn[id]
	return name, ok
}

// AddLabel adds a label to the mapping
func (m LabelMap) AddLabel(id, name string) {
	m.ntid[name] = id
	m.idtn[id] = name
}
