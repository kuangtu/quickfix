package quickfix

import (
	"bytes"
	"sort"
	"time"
)

//包含了所有的字段TagValue中有tag和value
//切片类型，存储多个TagValue??
//field应该只用了一个。
//field stores a slice of TagValues
type field []TagValue

func fieldTag(f field) Tag {
	return f[0].tag
}

//设置索引为0的元素
func initField(f field, tag Tag, value []byte) {
	f[0].init(tag, value)
}

//遍历了切片，对于其中的tv.bytes写入到buffer中
func writeField(f field, buffer *bytes.Buffer) {
	for _, tv := range f {
		buffer.Write(tv.bytes)
	}
}

//判断两个tag的顺序
// tagOrder true if tag i should occur before tag j
type tagOrder func(i, j Tag) bool

//tag类型排序，以为FIX部分字段有顺序设置
type tagSort struct {
	tags    []Tag
	compare tagOrder
}

func (t tagSort) Len() int           { return len(t.tags) }
func (t tagSort) Swap(i, j int)      { t.tags[i], t.tags[j] = t.tags[j], t.tags[i] }
func (t tagSort) Less(i, j int) bool { return t.compare(t.tags[i], t.tags[j]) }

//多个fields，通过map创建了tag到field的映射
//FieldMap is a collection of fix fields that make up a fix message.
type FieldMap struct {
	tagLookup map[Tag]field
	tagSort
}

//按照序号进行字段的排序
// ascending tags
func normalFieldOrder(i, j Tag) bool { return i < j }

func (m *FieldMap) init() {
	m.initWithOrdering(normalFieldOrder)
}

func (m *FieldMap) initWithOrdering(ordering tagOrder) {
	m.tagLookup = make(map[Tag]field)
	m.compare = ordering
}

//获取FieldMap中tag，返回切片形式
//Tags returns all of the Field Tags in this FieldMap
func (m FieldMap) Tags() []Tag {
	tags := make([]Tag, 0, len(m.tagLookup))
	for t := range m.tagLookup {
		tags = append(tags, t)
	}

	return tags
}

//Get parses out a field in this FieldMap. Returned reject may indicate the field is not present, or the field value is invalid.
func (m FieldMap) Get(parser Field) MessageRejectError {
	return m.GetField(parser.Tag(), parser)
}

//Has returns true if the Tag is present in this FieldMap
func (m FieldMap) Has(tag Tag) bool {
	_, ok := m.tagLookup[tag]
	return ok
}

//从FieldMap中根据Tag获取Fieldvalue，然后parser.Read读取value
//GetField parses of a field with Tag tag. Returned reject may indicate the field is not present, or the field value is invalid.
func (m FieldMap) GetField(tag Tag, parser FieldValueReader) MessageRejectError {
	f, ok := m.tagLookup[tag]
	if !ok {
		return ConditionallyRequiredFieldMissing(tag)
	}

	if err := parser.Read(f[0].value); err != nil {
		return IncorrectDataFormatForValue(tag)
	}

	return nil
}

//直接返回Value对应的字节切片
//GetBytes is a zero-copy GetField wrapper for []bytes fields
func (m FieldMap) GetBytes(tag Tag) ([]byte, MessageRejectError) {
	f, ok := m.tagLookup[tag]
	if !ok {
		return nil, ConditionallyRequiredFieldMissing(tag)
	}

	return f[0].value, nil
}

//获取fields的Value
//GetBool is a GetField wrapper for bool fields
func (m FieldMap) GetBool(tag Tag) (bool, MessageRejectError) {
	var val FIXBoolean
	if err := m.GetField(tag, &val); err != nil {
		return false, err
	}
	return bool(val), nil
}

//读取整数
//GetInt is a GetField wrapper for int fields
func (m FieldMap) GetInt(tag Tag) (int, MessageRejectError) {
	bytes, err := m.GetBytes(tag)
	if err != nil {
		return 0, err
	}

	var val FIXInt
	if val.Read(bytes) != nil {
		err = IncorrectDataFormatForValue(tag)
	}

	return int(val), err
}

//GetTime is a GetField wrapper for utc timestamp fields
func (m FieldMap) GetTime(tag Tag) (t time.Time, err MessageRejectError) {
	bytes, err := m.GetBytes(tag)
	if err != nil {
		return
	}

	var val FIXUTCTimestamp
	if val.Read(bytes) != nil {
		err = IncorrectDataFormatForValue(tag)
	}

	return val.Time, err
}

//GetString is a GetField wrapper for string fields
func (m FieldMap) GetString(tag Tag) (string, MessageRejectError) {
	var val FIXString
	if err := m.GetField(tag, &val); err != nil {
		return "", err
	}
	return string(val), nil
}

//GetGroup is a Get function specific to Group Fields.
func (m FieldMap) GetGroup(parser FieldGroupReader) MessageRejectError {
	f, ok := m.tagLookup[parser.Tag()]
	if !ok {
		return ConditionallyRequiredFieldMissing(parser.Tag())
	}

	if _, err := parser.Read(f); err != nil {
		if msgRejErr, ok := err.(MessageRejectError); ok {
			return msgRejErr
		}
		return IncorrectDataFormatForValue(parser.Tag())
	}

	return nil
}

//设置字段中的值
//SetField sets the field with Tag tag
func (m *FieldMap) SetField(tag Tag, field FieldValueWriter) *FieldMap {
	return m.SetBytes(tag, field.Write())
}

//SetBytes sets bytes
func (m *FieldMap) SetBytes(tag Tag, value []byte) *FieldMap {
	//如果没有tag，创建tag
	f := m.getOrCreate(tag)
	initField(f, tag, value)
	return m
}

//SetBool is a SetField wrapper for bool fields
func (m *FieldMap) SetBool(tag Tag, value bool) *FieldMap {
	return m.SetField(tag, FIXBoolean(value))
}

//SetInt is a SetField wrapper for int fields
func (m *FieldMap) SetInt(tag Tag, value int) *FieldMap {
	v := FIXInt(value)
	return m.SetBytes(tag, v.Write())
}

//SetString is a SetField wrapper for string fields
func (m *FieldMap) SetString(tag Tag, value string) *FieldMap {
	return m.SetBytes(tag, []byte(value))
}

//Clear purges all fields from field map
func (m *FieldMap) Clear() {
	//re-slice
	m.tags = m.tags[0:0]
	//删除map的value
	for k := range m.tagLookup {
		delete(m.tagLookup, k)
	}
}

//CopyInto overwrites the given FieldMap with this one
func (m *FieldMap) CopyInto(to *FieldMap) {
	to.tagLookup = make(map[Tag]field)
	for tag, f := range m.tagLookup {
		clone := make(field, 1)
		clone[0] = f[0]
		to.tagLookup[tag] = clone
	}
	to.tags = make([]Tag, len(m.tags))
	copy(to.tags, m.tags)
	to.compare = m.compare
}

func (m *FieldMap) add(f field) {
	t := fieldTag(f)
	if _, ok := m.tagLookup[t]; !ok {
		m.tags = append(m.tags, t)
	}

	m.tagLookup[t] = f
}

func (m *FieldMap) getOrCreate(tag Tag) field {
	if f, ok := m.tagLookup[tag]; ok {
		f = f[:1]
		return f
	}

	f := make(field, 1)
	m.tagLookup[tag] = f
	m.tags = append(m.tags, tag)
	return f
}

//Set is a setter for fields
func (m *FieldMap) Set(field FieldWriter) *FieldMap {
	f := m.getOrCreate(field.Tag())
	initField(f, field.Tag(), field.Write())
	return m
}

//SetGroup is a setter specific to group fields
func (m *FieldMap) SetGroup(field FieldGroupWriter) *FieldMap {
	_, ok := m.tagLookup[field.Tag()]
	if !ok {
		m.tags = append(m.tags, field.Tag())
	}
	m.tagLookup[field.Tag()] = field.Write()
	return m
}

func (m *FieldMap) sortedTags() []Tag {
	sort.Sort(m)
	return m.tags
}

//按照tag序号，将vlaue写入到Buffer中
func (m FieldMap) write(buffer *bytes.Buffer) {
	for _, tag := range m.sortedTags() {
		if f, ok := m.tagLookup[tag]; ok {
			writeField(f, buffer)
		}
	}
}

func (m FieldMap) total() int {
	total := 0
	for _, fields := range m.tagLookup {
		for _, tv := range fields {
			switch tv.tag {
			case tagCheckSum: //tag does not contribute to total
			default:
				total += tv.total()
			}
		}
	}

	return total
}

func (m FieldMap) length() int {
	length := 0
	for _, fields := range m.tagLookup {
		for _, tv := range fields {
			switch tv.tag {
			case tagBeginString, tagBodyLength, tagCheckSum: //tags do not contribute to length
			default:
				length += tv.length()
			}
		}
	}

	return length
}
