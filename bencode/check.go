package bencode

// Dictionary:
//   1. All keys must be byte strings.
//   2. All keys must appear in lexicographical order.

func (body *Body) Check() (f bool) {
	f = true

	if body.btype == DictType {
		if len(body.dict) == 0 {
			return false
		}
		f = f && body.dict[0].value.Check()
		for i := 1; i < len(body.dict); i++ {
			if string(body.dict[i].key) < string(body.dict[i-1].key) {
				return false
			}
			if body.dict[i].value.btype == DictType {
				f = f && body.dict[i].value.Check()
			}
		}
	}
	return
}
