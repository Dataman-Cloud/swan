package state

type SlotsById []*Slot

func (a SlotsById) Len() int           { return len(a) }
func (a SlotsById) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SlotsById) Less(i, j int) bool { return a[i].Index < a[j].Index }
