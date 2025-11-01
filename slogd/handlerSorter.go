package slogd

type NameHandlerSorter []*Handler

func (s NameHandlerSorter) Len() int      { return len(s) }
func (s NameHandlerSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s NameHandlerSorter) Less(i, j int) bool {
	return s[i].Name() < s[j].Name()
}

type FailoverHandlerSorter []*Handler

func (s FailoverHandlerSorter) Len() int      { return len(s) }
func (s FailoverHandlerSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s FailoverHandlerSorter) Less(i, j int) bool {
	return s[i].GetFailoverOrder() < s[j].GetFailoverOrder()
}
