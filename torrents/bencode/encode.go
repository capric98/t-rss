package bencode

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	ErrInvalidBody    = errors.New("bencode: Body did not pass Check().")
	ErrDictWithoutKey = errors.New("bencode: Dict need a key to add value.")
	ErrUnknownType    = errors.New("bencode: Unknown type to add.")
	t                 = []string{"Dict", "List", "Int", "ByteString"}
)

type BEncoder struct {
	stack []*Body
	pos   int
}

func NewEncoder() *BEncoder {
	encoder := &BEncoder{
		stack: make([]*Body, MaxDepth),
		pos:   0,
	}
	encoder.stack[0] = &Body{
		btype: ListType,
		dict:  make([]kvBody, 0, 2),
	}
	return encoder
}

func (e *BEncoder) Add(k string, val interface{}) error {
	var v *Body
	switch val := val.(type) {
	case byte, int, int8, int16, int32, int64:
		v = &Body{
			btype: IntValue,
			value: vtoint64(val),
		}
	case string:
		v = &Body{
			btype:   ByteString,
			byteStr: []byte(val),
		}
	case []byte:
		v = &Body{
			btype:   ByteString,
			byteStr: val,
		}
	default:
		return ErrUnknownType
	}

	if k == "" && e.stack[e.pos].btype == DictType {
		return ErrDictWithoutKey
	}
	switch e.stack[e.pos].btype {
	case DictType:
		newkv := kvBody{
			key:   []byte(k),
			value: v,
		}
		if len(e.stack[e.pos].dict) == 0 {
			e.stack[e.pos].dict = append(e.stack[e.pos].dict, newkv)
		} else {
			i := e.stack[e.pos].inspos(k)
			e.stack[e.pos].dict = append(e.stack[e.pos].dict[:i], append([]kvBody{newkv}, e.stack[e.pos].dict[i:]...)...)
		}
	case ListType:
		e.stack[e.pos].dict = append(e.stack[e.pos].dict, kvBody{value: v})
	default:
		return errors.New("bencode: Cannot add k-v struct to " + t[e.stack[e.pos].btype])
	}
	return nil
}

func (e *BEncoder) NewDict(k string) error {
	return e.newpart(DictType, k)
}

func (e *BEncoder) NewList(k string) error {
	return e.newpart(ListType, k)
}

func (e *BEncoder) EndPart() error {
	if e.pos == 0 {
		return ErrTooManyEnd
	}
	switch e.stack[e.pos].btype {
	case ListType, DictType:
		e.pos--
	default:
		return errors.New("bencode: Cannot end at " + t[e.stack[e.pos].btype])
	}
	return nil
}

func (e *BEncoder) End() []*Body {
	result := make([]*Body, 0, 1)
	for i := 0; i < len(e.stack[0].dict); i++ {
		result = append(result, e.stack[0].dict[i].value)
	}
	return result
}

func (e *BEncoder) newpart(Type int, k string) error {
	if e.pos+1 == MaxDepth {
		return ErrEncodeDepthTooGreat
	}
	switch e.stack[e.pos].btype {
	case ListType:
		e.pos++
		e.stack[e.pos] = &Body{
			btype: Type,
			dict:  make([]kvBody, 0, 2),
		}
		e.stack[e.pos-1].dict = append(e.stack[e.pos-1].dict, kvBody{value: e.stack[e.pos]})
	case DictType:
		if k == "" {
			return ErrDictWithoutKey
		}
		e.pos++
		e.stack[e.pos] = &Body{
			btype: Type,
			dict:  make([]kvBody, 0, 2),
		}
		i := e.stack[e.pos-1].inspos(k)
		e.stack[e.pos-1].dict = append(e.stack[e.pos-1].dict[:i], append([]kvBody{kvBody{
			key:   []byte(k),
			value: e.stack[e.pos],
		}}, e.stack[e.pos-1].dict[i:]...)...)
	default:
		return errors.New("bencode: Cannot add Dict to " + t[e.stack[e.pos].btype])
	}
	return nil
}

func (body *Body) inspos(k string) int {
	l := -1
	r := len(body.dict)
	for {
		if l+1 >= r {
			return l + 1
		}
		m := (l + r) / 2
		if k < string(body.dict[m].key) {
			r = m
		}
		if k > string(body.dict[m].key) {
			l = m
		}
	}
}

func (body *Body) Encode() ([]byte, error) {
	if !body.Check() {
		return nil, ErrInvalidBody
	}
	return encode(body), nil
}

func encode(b *Body) []byte {
	var buf bytes.Buffer

	switch b.btype {
	case IntValue:
		_ = (&buf).WriteByte('i')
		i := b.value
		if i < 0 {
			i = -i
			_ = (&buf).WriteByte('-')
		}
		_, _ = (&buf).WriteString(fmt.Sprintf("%d", i))
		_ = (&buf).WriteByte('e')
	case ByteString:
		_, _ = (&buf).WriteString(fmt.Sprintf("%d", len(b.byteStr)))
		_ = (&buf).WriteByte(':')
		_, _ = (&buf).Write(b.byteStr)
	default:
		if b.btype == ListType {
			_ = (&buf).WriteByte('l')
		} else {
			_ = (&buf).WriteByte('d')
		}
		for _, v := range b.dict {
			if v.key != nil {
				_, _ = (&buf).WriteString(fmt.Sprintf("%d", len(v.key)))
				_ = (&buf).WriteByte(':')
				_, _ = (&buf).Write(v.key)
			}
			_, _ = (&buf).Write(encode(v.value))
		}
		_ = (&buf).WriteByte('e')
	}
	return buf.Bytes()
}

func (body *Body) Delete(k string) {
	var pos int
	if pos = body.findpos(k); pos == -1 {
		return
	}
	body.dict = append(body.dict[:pos], body.dict[pos+1:]...)
	//body.dict[len(body.dict)-1] = kvBody{}
	//body.dict = body.dict[:len(body.dict)-1]
}
func (body *Body) DeleteN(n int) {
	if len(body.dict) <= n {
		return
	}
	body.dict = append(body.dict[:n], body.dict[n+1:]...)
	//body.dict[len(body.dict)-1] = kvBody{}
	//body.dict = body.dict[:len(body.dict)-1]
}
func (body *Body) Edit(v interface{}) {
	if body.btype != ByteString && body.btype != IntValue {
		return
	}
	switch v := v.(type) {
	case byte, int, int8, int16, int32, int64:
		body.btype = IntValue
		body.value = vtoint64(v)
	case string:
		body.btype = ByteString
		body.byteStr = []byte(v)
	case []byte:
		body.btype = ByteString
		body.byteStr = v
	}
}
func (body *Body) AddPart(k string, v *Body) error {
	if body.btype != ListType && body.btype != DictType {
		return ErrTypeFrom
	}
	if body.btype == DictType && k == "" {
		return ErrDictWithoutKey
	}
	i := body.inspos(k)
	if body.btype == DictType {
		body.dict = append(body.dict[:i], append([]kvBody{kvBody{
			key:   []byte(k),
			value: v,
		}}, body.dict[i:]...)...)
	} else {
		body.dict = append(body.dict[:i], append([]kvBody{kvBody{
			value: v,
		}}, body.dict[i:]...)...)
	}
	return nil
}

func NewBStr(s string) *Body {
	return &Body{
		btype:   ByteString,
		byteStr: []byte(s),
	}
}

func NewEmptyList() *Body {
	return &Body{
		btype: ListType,
		dict:  make([]kvBody, 0, 2),
	}
}

func (b *Body) AnnounceList(s []string) {
	if b.btype != ListType {
		return
	}
	for _, v := range s {
		tmp := kvBody{
			value: &Body{
				btype: ListType,
				dict: []kvBody{kvBody{value: &Body{
					btype:   ByteString,
					byteStr: []byte(v),
				}}},
			},
		}
		b.dict = append(b.dict, tmp)
	}
}

func vtoint64(v interface{}) int64 {
	switch v := v.(type) {
	case byte:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}
