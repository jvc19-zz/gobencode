// decodes bencode using bittorrent specification
// based on the work by Mark Samman <https://github.com/marksamman/bencode>

package gobencode

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

type decoder struct {
	bufio.Reader
}

func (decoder *decoder) readIntUntil(until byte) (interface{}, error) {
	res, err := decoder.ReadSlice(until)
	if err != nil {
		return nil, err
	}

	str := string(res[:len(res)-1]) // TODO try just res

	if value, err := strconv.ParseInt(str, 10, 64); err == nil {
		return value, nil
	} else if value, err := strconv.ParseUint(str, 10, 64); err == nil {
		return value, nil
	}
	return nil, err
}

func (decoder *decoder) readInt() (interface{}, error) {
	return decoder.readIntUntil('e')
}

func (decoder *decoder) readByType(identifyer byte) (item interface{}, err error) {
	switch identifyer {
	case 'i':
		item, err = decoder.readInt()
	case 'l':
		item, err = decoder.readList()
	case 'd':
		item, err = decoder.readDictionary()
	default:
		if err := decoder.UnreadByte(); err != nil {
			return nil, err
		}
		item, err = decoder.readString()
	}
	return item, err
}

func (decoder *decoder) readList() ([]interface{}, error) {
	var list []interface{}
	for {
		ch, err := decoder.ReadByte()
		if err != nil {
			return nil, err
		}

		if ch == 'e' {
			break
		}

		item, err := decoder.readByType(ch)
		if err != nil {
			return nil, err
		}
		list = append(list, item)
	}
	return list, nil
}

func (decoder *decoder) readString() (string, error) {
	len, err := decoder.readIntUntil(':')
	if err != nil {
		return "", err
	}

	var stringLength int64
	var ok bool
	if stringLength, ok = len.(int64); !ok {
		return "", errors.New("string length may not exceed the size of int64")
	}

	if stringLength < 0 {
		return "", errors.New("string length may not be a negative number")
	}

	buffer := make([]byte, stringLength)
	_, err = io.ReadFull(decoder, buffer)
	return string(buffer), err
}

func (decoder *decoder) readDictionary() (map[string]interface{}, error) {
	dict := make(map[string]interface{})
	for {
		key, err := decoder.readString()
		if err != nil {
			return nil, err
		}

		ch, err := decoder.ReadByte()
		if err != nil {
			return nil, err
		}

		item, err := decoder.readByType(ch)
		if err != nil {
			return nil, err
		}

		dict[key] = item

		nextByte, err := decoder.ReadByte()
		if err != nil {
			return nil, err
		}

		if nextByte == 'e' {
			break
		} else if err := decoder.UnreadByte(); err != nil {
			return nil, err
		}
	}
	return dict, nil
}

// Decode takes an io.Reader and parses it as bencode,
// on failure, err will be a non-nil value
func Decode(reader io.Reader) (map[string]interface{}, error) {
	decoder := decoder{*bufio.NewReader(reader)}
	if firstByte, err := decoder.ReadByte(); err != nil {
		return make(map[string]interface{}), nil // TODO why not nil, err? Is it empty?
	} else if firstByte != 'd' {
		return nil, errors.New("bencode data must begin with a dictionary")
	}
	return decoder.readDictionary()
}
