package sync

import (
	"sort"

	"github.com/jlaffaye/ftp"
)

func AreEqual(first, second *ftp.Entry) bool {
	return first.Type == second.Type && first.Name == second.Name && first.Size == second.Size
}

type ByName []*ftp.Entry

func (a ByName) Len() int {
	return len(a)
}

func (a ByName) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByName) Less(i, j int) bool {
	return a[i].Name < a[j].Name
}

func (a ByName) Search(name string) *ftp.Entry {
	i := sort.Search(len(a), func(i int) bool { return a[i].Name >= name })
	if i < len(a) && a[i].Name == name {
		return a[i]
	}
	return nil
}
