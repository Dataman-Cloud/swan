package utils

import (
	"strings"
)

// example node path  => /swan/leader-election/_c_c7b2927d40ec05db4d199a804437995c-node0000000023
// sortable value is node0000000023
type SortableNodePath []string

func (a SortableNodePath) Len() int      { return len(a) }
func (a SortableNodePath) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SortableNodePath) Less(i, j int) bool {
	if strings.Contains(a[i], "-") {
		return strings.SplitN(a[i], "-", -1)[1] < strings.SplitN(a[j], "-", -1)[1]
	} else {
		return strings.Compare(a[i], a[j]) == -1
	}
}
