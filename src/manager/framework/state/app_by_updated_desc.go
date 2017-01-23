package state

type AppsByUpdated []*App

func (a AppsByUpdated) Len() int           { return len(a) }
func (a AppsByUpdated) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AppsByUpdated) Less(i, j int) bool { return a[i].Updated.After(a[j].Updated) } // NOTE(xychu): Desc order
