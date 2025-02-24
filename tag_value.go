package quickfix

import (
	"bytes"
	"fmt"
	"strconv"
)

//Tag=Value模式
//TagValue is a low-level FIX field abstraction
type TagValue struct {
	tag   Tag
	value []byte
	//bytes表示了整个tag=value字节序列
	bytes []byte
}

//初始化Tag=Value
func (tv *TagValue) init(tag Tag, value []byte) {
	tv.bytes = strconv.AppendInt(nil, int64(tag), 10)
	//append中的...表示将slice拆开为单个的元素，再添加
	tv.bytes = append(tv.bytes, []byte("=")...)
	tv.bytes = append(tv.bytes, value...)
	tv.bytes = append(tv.bytes, []byte("")...)

	tv.tag = tag
	tv.value = value
}

//解析Tag=Value序列
func (tv *TagValue) parse(rawFieldBytes []byte) error {
	//按照"="进行切分
	sepIndex := bytes.IndexByte(rawFieldBytes, '=')

	//没有"="以及"="是首个字符，都是错误
	switch sepIndex {
	case -1:
		return fmt.Errorf("tagValue.Parse: No '=' in '%s'", rawFieldBytes)
	case 0:
		return fmt.Errorf("tagValue.Parse: No tag in '%s'", rawFieldBytes)
	}

	//获取tag至
	parsedTag, err := atoi(rawFieldBytes[:sepIndex])
	if err != nil {
		return fmt.Errorf("tagValue.Parse: %s", err.Error())
	}

	//从int转为Tag类型
	tv.tag = Tag(parsedTag)
	//bytes切片的长度
	n := len(rawFieldBytes)
	//value去除了最后的SOH
	tv.value = rawFieldBytes[(sepIndex + 1):(n - 1):(n - 1)]
	tv.bytes = rawFieldBytes[:n:n]

	return nil
}

func (tv TagValue) String() string {
	return string(tv.bytes)
}

func (tv TagValue) total() int {
	total := 0

	for _, b := range []byte(tv.bytes) {
		total += int(b)
	}

	return total
}

func (tv TagValue) length() int {
	return len(tv.bytes)
}
