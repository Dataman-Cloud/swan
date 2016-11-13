package backend

import (
	"github.com/Dataman-Cloud/swan/types"
	"strconv"
	"strings"
)

type TaskSorter []*types.Task

func (s TaskSorter) Len() int      { return len(s) }
func (s TaskSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s TaskSorter) Less(i, j int) bool {
	a, _ := strconv.Atoi(strings.Split(s[i].Name, ".")[0])
	b, _ := strconv.Atoi(strings.Split(s[j].Name, ".")[0])

	return a < b
}
