// encodes bencode using bittorrent specification
// based on the work by Mark Samman <https://github.com/marksamman/bencode>

package gobencode

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"
)

// encoder is just a Buffer with functions to read and write
type encoder struct {
	bytes.Buffer
}

// writes string to a buffer following bencode specification
func (encoder *encoder) writeString(str string) {
	encoder.WriteString(strconv.Itoa(len(str)))
	encoder.WriteByte(':')
	encoder.WriteString(str)
}

// writes integer to a buffer following bencode specification
func (encoder *encoder) writeInt(value int64) {
	encoder.WriteByte('i')
	encoder.WriteString(strconv.FormatInt(value, 10))
	encoder.WriteByte('e')
}

// writes unsigned integer to a buffer following bencode specification
func (encoder *encoder) writeUInt(value uint64) {
	encoder.WriteByte('i')
	encoder.WriteString(strconv.FormatUint(value, 10))
	encoder.WriteByte('e')
}

// calls write functions according to input type
func (encoder *encoder) writeByType(value interface{}) {
	switch v := value.(type) {
	case string:
		encoder.writeString(v)
	case int, int8, int16, int32, int64:
		encoder.writeInt(reflect.ValueOf(v).Int())
	case uint, uint8, uint16, uint32, uint64:
		encoder.writeUInt(reflect.ValueOf(v).Uint())
	case []interface{}:
		encoder.writeList(v)
	case map[string]interface{}:
		encoder.writeDict(v)
	}
}

// writes list to a buffer following bencode specification
func (encoder *encoder) writeList(list []interface{}) {
	encoder.WriteByte('l')
	for _, v := range list {
		encoder.writeByType(v)
	}
	encoder.WriteByte('e')
}

// writes dictionary to a buffer following bencode specification
func (encoder *encoder) writeDict(d map[string]interface{}) {
	// get list of sorted keysencoder.WriteByte('d')
	keyList := make(sort.StringSlice, len(d))
	i := 0
	for key := range d {
		keyList[i] = key
		i++
	}
	keyList.Sort()

	encoder.WriteByte('d')
	for _, key := range keyList {
		encoder.writeString(key)    // write key
		encoder.writeByType(d[key]) // write value
	}
	encoder.WriteByte('e')
}

// Encode takes a bencode supported input. Returns a bencode encoded byte array
func Encode(value interface{}) []byte {
	encoder := encoder{}
	encoder.writeByType(value)
	return encoder.Bytes()
}
