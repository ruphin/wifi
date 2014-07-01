package wifi

import (
    "sort"
)

type set []int

type id_set [20]int

func NewSet(s []int) set {
    return s
}

func (s set) Key() id_set {
    var ret id_set
    sort.Ints(s)
    copy(ret[:], s[:])
    return ret
}
