/*
bep_0012: announce-list
bep_0030: Merkle hash torrent extension.
*/
package bencode

import (
	"errors"
)

const (
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

type KvBody struct {
	key   []byte
	value *Body
}

type Body struct {
	Type    int
	Value   int64
	ByteStr []byte
	Dict    []KvBody
	List    []*Body
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
		Type: ListType,
		List: make([]*Body, 0, 1),
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
				Type: DictType,
				Dict: make([]KvBody, 0, 8),
			}
		case ListType:
			tmp = &Body{
				Type: ListType,
				List: make([]*Body, 0, 8),
			}
		case IntValue:
			tmp = &Body{
				Type:  IntValue,
				Value: value,
			}
		case ByteString:
			offset += shift
			shift = int(value)
			tmp = &Body{
				Type:    ByteString,
				ByteStr: data[offset+1 : offset+int(value)+1],
			}
		}

		if mark != DorLEnd {
			switch stack[pos].Type {
			case DictType:
				if lastString == nil {
					if tmp.Type == ByteString {
						lastString = tmp.ByteStr
					} else {
						e = ErrInvalidDictKey
					}

				} else {
					stack[pos].Dict = append(stack[pos].Dict, KvBody{
						key:   lastString,
						value: tmp,
					})
					lastString = nil
				}
			case ListType:
				stack[pos].List = append(stack[pos].List, tmp)
			default:
				e = ErrTypeFrom
			}

			if tmp.Type < IntValue {
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
		//stack[0].List[0].print(0)
		//fmt.Println()
	}

	result = stack[0].List
	return
}

func (body *Body) Get(key string) *Body {
	if body.Type != DictType {
		return nil
	}
	dict := body.Dict
	dlen := len(dict) - 1
	l := 0
	r := dlen
	var m int
	var ckey string
	for {
		m = (l + r) / 2
		if m < 0 || m > dlen || l > r {
			return nil
		}
		ckey = string(dict[m].key)
		if ckey == key {
			return dict[m].value
		}
		if ckey > key {
			r = m - 1
		} else {
			l = m + 1
		}
	}
}
