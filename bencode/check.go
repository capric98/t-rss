package bencode

// Dictionary:
//   1. All keys must be byte strings.
//   2. All keys must appear in lexicographical order.

func (body *Body) Check() (f bool) {
	f = true
	if body.btype == DictType {
		for i := 0; i < len(body.dict); i++ {
			if body.dict[i].value.btype == DictType {
				f = f && body.dict[i].value.Check()
			}
		}
	}
	return
}
