/*
bep_0012: announce-list
bep_0030: Merkle hash torrent extension.
*/
package bencode

import (
	"errors"
)

const (
	Unknown    = -100
	DorLEnd    = -1 // Dictionary or List end mark.
	DictType   = 0
	ListType   = 1
	IntValue   = 2
	ByteString = 3 // Actually byte string.
	MaxDepth   = 10
)

var (
	ErrEncodeDepthTooGreat = errors.New("bencode: Bencode depth over than 10, it is abnormal!")
	ErrInvalidDictKey      = errors.New("bencode: Dictionary's key must be byte string!")
	ErrTypeFrom            = errors.New("bencode: Cannot be body after int or string.")
	ErrTooManyEnd          = errors.New("bencode: Too many end.")
)

type kvBody struct {
	value *Body
	key   []byte
}

type Body struct {
	btype   int
	value   int64
	byteStr []byte
	dict    []kvBody // nil key List -> Dict
}

func decodepart(data []byte) (typemark int, offset int, value int64, e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	offset = 0
	value = 0

	switch data[offset] {
	case 'd':
		// Dictionary start.
		typemark = DictType
	case 'l':
		// List start.
		typemark = ListType
	case 'e':
		// Dictionary or List end.
		typemark = DorLEnd
	case 'i':
		// Interger.
		offset++
		sgn := int64(1)
		if data[offset] == '-' {
			offset++
			sgn = -1
		}
		for data[offset] != 'e' {
			value = value*10 + int64(data[offset]-'0')
			offset++
		}
		value = sgn * value
		typemark = IntValue
	default:
		// Byte string.
		for data[offset] != ':' {
			value = value*10 + int64(data[offset]-'0')
			offset++
		}
		typemark = ByteString
	}

	return
}

func Decode(data []byte) (result []*Body, e error) {
	defer func() {
		if p := recover(); p != nil {
			e = p.(error)
		}
	}()

	var tmp *Body
	var offset, pos int
	stack := make([]*Body, MaxDepth+1)
	length := len(data)

	stack[0] = &Body{
		btype: ListType,
		dict:  make([]kvBody, 0, 1),
	}
	var lastString []byte

	for offset < length {
		mark, shift, value, err := decodepart(data[offset:])
		if err != nil {
			return nil, err
		}

		switch mark {
		case DorLEnd:
			pos--
			lastString = nil
		case DictType:
			tmp = &Body{
				btype: DictType,
				dict:  make([]kvBody, 0, 2),
			}
		case ListType:
			tmp = &Body{
				btype: ListType,
				dict:  make([]kvBody, 0, 4),
			}
		case IntValue:
			tmp = &Body{
				btype: IntValue,
				value: value,
			}
		case ByteString:
			offset += shift
			shift = int(value)
			tmp = &Body{
				btype:   ByteString,
				byteStr: data[offset+1 : offset+int(value)+1],
			}
		}

		if mark != DorLEnd {
			switch stack[pos].btype {
			case DictType:
				if lastString == nil {
					if tmp.btype == ByteString {
						lastString = tmp.byteStr
					} else {
						e = ErrInvalidDictKey
					}

				} else {
					stack[pos].dict = append(stack[pos].dict, kvBody{
						key:   lastString,
						value: tmp,
					})
					lastString = nil
				}
			case ListType:
				stack[pos].dict = append(stack[pos].dict, kvBody{value: tmp})
			default:
				e = ErrTypeFrom
			}

			if tmp.btype < IntValue {
				(pos)++
				stack[pos] = tmp
			}
		}

		offset += shift + 1
		if pos > MaxDepth {
			e = ErrEncodeDepthTooGreat
			return
		}
		if pos < 0 {
			e = ErrTooManyEnd
			return nil, ErrTooManyEnd
		}
		if e != nil {
			return
		}
	}

	result = make([]*Body, len(stack[0].dict))
	for i := 0; i < len(stack[0].dict); i++ {
		result[i] = stack[0].dict[i].value
	}
	return
}

func (body *Body) Type() int {
	return body.btype
}

func (body *Body) Value() int64 {
	if body.btype == IntValue {
		return body.value
	}
	return Unknown
}

func (body *Body) BStr() []byte {
	if body.btype == ByteString {
		return body.byteStr
	}
	return nil
}

func (body *Body) Dict(key string) *Body {
	if body.btype != DictType {
		return nil
	}
	pos := body.findpos(key)
	if pos == -1 {
		return nil
	}
	return body.dict[pos].value
}

func (body *Body) findpos(k string) int {
	dict := body.dict
	dlen := len(dict) - 1
	l := 0
	r := dlen
	var m int
	var key string
	for {
		m = (l + r) / 2
		if m < 0 || m > dlen || l > r {
			return -1
		}
		key = string(dict[m].key)
		if key == k {
			return m
		}
		if key > k {
			r = m - 1
		} else {
			l = m + 1
		}
	}
}

func (body *Body) DictN(n int) (k string, b *Body) {
	if body.btype != DictType || len(body.dict) <= n {
		return
	}
	return string(body.dict[n].key), body.dict[n].value
}

func (body *Body) List(n int) *Body {
	if body.btype != ListType || len(body.dict) <= n {
		return nil
	}
	return body.dict[n].value
}

func (body *Body) Len() int {
	if body == nil {
		return 0
	}
	if body.btype == ListType || body.btype == DictType {
		return len(body.dict)
	}
	return 0
}
