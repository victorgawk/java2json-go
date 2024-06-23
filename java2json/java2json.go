package java2json

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// ParseJavaObject parses a serialized java object.
func ParseJavaObject(buf []byte) (interface{}, error) {
	jop := NewJavaObjectParser(bytes.NewReader(buf))
	return jop.ParseJavaObject()
}

// NewJavaObjectParser reads serialized java objects from stream.
func NewJavaObjectParser(rd io.Reader) *JavaObjectParser {
	buf := bufio.NewReaderSize(rd, defaultBufferSize)

	jop := &JavaObjectParser{
		rd:               buf,
		maxDataBlockSize: buf.Size(),
	}

	return jop
}

// SetMaxDataBlockSize set the maximum size of the parsed data block,
// by default it is equal to the value of the buffer size bufio.Reader or size of bytes.Reader.
func (jop *JavaObjectParser) SetMaxDataBlockSize(maxDataBlockSize int) {
	jop.maxDataBlockSize = maxDataBlockSize
}

// ParseSerializedObject parses a serialized java object from stream.
func (jop *JavaObjectParser) ParseJavaObject() (content interface{}, err error) {
	if err = jop.magic(); err != nil {
		return
	}

	if err = jop.version(); err != nil {
		return
	}

	if content, err = jop.content(nil); err != nil {
		if errors.Cause(err).Error() == io.EOF.Error() {
			err = errors.New("premature end of input")
		}

		return
	}

	if !jop.end() {
		err = errors.New("object already parsed but there is more data")
	}

	return
}

const magicNumber uint16 = 0xACED
const protocolVersion uint16 = 5
const objectValueField string = "@@value@@"
const cycleValue = "[CYCLE]"
const defaultBufferSize int = 1024
const typeCodeMask uint8 = 0x70
const endBlock endBlockT = "endBlock"
const objectDataMinLength int = 4
const refIdMask int32 = 0x7E0000
const timestampBlockSize int = 8
const minClassNameLength int = 2
const serialVersionUIDLength int = 8
const classFlagsMask uint8 = 0x0F
const scSerializableWithoutWriteMethod uint8 = 0x02
const scSerializableWithWriteMethod uint8 = 0x03
const scExternalizeWithBlockData uint8 = 0x04
const scExternalizeWithoutBlockData uint8 = 0x0c

// typeNames includes all known type names.
var typeNames = []string{
	"Null",
	"Reference",
	"ClassDesc",
	"Object",
	"String",
	"Array",
	"Class",
	"BlockData",
	"EndBlockData",
	"Reset",
	"BlockDataLong",
	"Exception",
	"LongString",
	"ProxyClassDesc",
	"Enum",
}

// maxTypeCode is used to ensure an encountered type is known.
var maxTypeCode = uint8(len(typeNames) - 1)

// allowedClazzNames includes all allowed names when parsing a class descriptor.
var allowedClazzNames = map[string]bool{
	"ClassDesc":      true,
	"ProxyClassDesc": true,
	"Null":           true,
	"Reference":      true,
}

// postProc handlers are used to format deserialized objects for easier consumption.
type postProc func(map[string]interface{}, []interface{}) (map[string]interface{}, error)

// knownPostProcs maps serialized object signatures to PostProc implementations.
var knownPostProcs = map[string]postProc{
	"java.lang.Byte@9c4e6084ee50f51c":                            primObjectPostProc,
	"java.lang.Character@348b47d96b1a2678":                       primObjectPostProc,
	"java.lang.Double@80b3c24a296bfb04":                          primObjectPostProc,
	"java.lang.Float@daedc9a2db3cf0ec":                           primObjectPostProc,
	"java.lang.Integer@12e2a0a4f7818738":                         primObjectPostProc,
	"java.lang.Long@3b8be490cc8f23df":                            primObjectPostProc,
	"java.lang.Short@684d37133460da52":                           primObjectPostProc,
	"java.lang.Boolean@cd207280d59cfaee":                         primObjectPostProc,
	"java.util.ArrayList@7881d21d99c7619d":                       listPostProc,
	"java.util.ArrayDeque@207cda2e240da08b":                      listPostProc,
	"java.util.Hashtable@13bb0f25214ae4b8":                       mapPostProc,
	"java.util.HashMap@0507dac1c31660d1":                         mapPostProc,
	"java.util.EnumMap@065d7df7be907ca1":                         enumMapPostProc,
	"java.util.HashSet@ba44859596b8b734":                         hashSetPostProc,
	"java.util.Date@686a81014b597419":                            datePostProc,
	"java.util.Calendar@e6ea4d1ec8dc5b8e":                        calendarPostProc,
	"java.util.Arrays$ArrayList@d9a43cbecd8806d2":                arraysArrayListPostProc,
	"java.util.concurrent.CopyOnWriteArrayList@785d9fd546ab90c3": listPostProc,
	"java.util.CollSer@578eabb63a1ba811":                         listPostProc,
}

// primitiveHandler are used to read primitive values.
type primitiveHandler func(jop *JavaObjectParser) (interface{}, error)

// JavaObjectParser reads serialized java objects
// see: https://docs.oracle.com/javase/8/docs/platform/serialization/spec/protocol.html
type JavaObjectParser struct {
	buf              bytes.Buffer
	rd               *bufio.Reader
	handles          []interface{}
	maxDataBlockSize int
}

// clazz contains java class info.
type clazz struct {
	super            *clazz
	annotations      []interface{}
	fields           []*field
	serialVersionUID string
	name             string
	flags            uint8
	isEnum           bool
}

// field contains info about a single class member.
type field struct {
	className string
	typeName  string
	name      string
}

type endBlockT string

// parser is a func capable of reading a single serialized type.
type parser func(jop *JavaObjectParser) (interface{}, error)

// knownParsers maps serialized names to corresponding parser implementations.
var knownParsers map[string]parser

func init() {
	knownParsers = map[string]parser{
		"Enum":          parseEnum,
		"BlockDataLong": parseBlockDataLong,
		"BlockData":     parseBlockData,
		"EndBlockData":  parseEndBlockData,
		"ClassDesc":     parseClassDesc,
		"Class":         parseClass,
		"Array":         parseArray,
		"LongString":    parseLongString,
		"String":        parseString,
		"Null":          parseNull,
		"Object":        parseObject,
		"Reference":     parseReference,
	}
}

// primitiveHandlers maps serialized primitive identifiers to a corresponding primitiveHandler.
var primitiveHandlers = map[string]primitiveHandler{
	"B": func(jop *JavaObjectParser) (b interface{}, err error) {
		if b, err = jop.readInt8(); err != nil {
			err = errors.Wrap(err, "error reading byte primitive")
		}

		return
	},
	"C": func(jop *JavaObjectParser) (char interface{}, err error) {
		var charCode uint16
		if charCode, err = jop.readUInt16(); err != nil {
			err = errors.Wrap(err, "error reading char primitive")
		} else {
			char = string(rune(charCode))
		}

		return
	},
	"D": func(jop *JavaObjectParser) (double interface{}, err error) {
		if double, err = jop.readFloat64(); err != nil {
			err = errors.Wrap(err, "error reading double primitive")
		}

		return
	},
	"F": func(jop *JavaObjectParser) (f32 interface{}, err error) {
		if f32, err = jop.readFloat32(); err != nil {
			err = errors.Wrap(err, "error reading float primitive")
		}

		return
	},
	"I": func(jop *JavaObjectParser) (i32 interface{}, err error) {
		if i32, err = jop.readInt32(); err != nil {
			err = errors.Wrap(err, "error reading int primitive")
		}

		return
	},
	"J": func(jop *JavaObjectParser) (long interface{}, err error) {
		if long, err = jop.readInt64(); err != nil {
			err = errors.Wrap(err, "error reading long primitive")
		}

		return
	},
	"S": func(jop *JavaObjectParser) (short interface{}, err error) {
		if short, err = jop.readInt16(); err != nil {
			err = errors.Wrap(err, "error reading short primitive")
		}

		return
	},
	"Z": func(jop *JavaObjectParser) (b interface{}, err error) {
		var x int8
		if x, err = jop.readInt8(); err != nil {
			err = errors.Wrap(err, "error reading boolean primitive")
		} else {
			b = x != 0
		}

		return
	},
	"L": func(jop *JavaObjectParser) (obj interface{}, err error) {
		if obj, err = jop.content(nil); err != nil {
			err = errors.Wrap(err, "error reading object primitive")
		}

		return
	},
	"[": func(jop *JavaObjectParser) (arr interface{}, err error) {
		if arr, err = jop.content(nil); err != nil {
			err = errors.Wrap(err, "error reading array primitive")
		}

		return
	},
}

// newHandle adds a parsed object to the existing indexed handles which can be used later to lookup references to
// existing objects.
func (jop *JavaObjectParser) newHandle(obj interface{}) interface{} {
	jop.handles = append(jop.handles, obj)
	return obj
}

// content reads the next object in the stream and parses it.
func (jop *JavaObjectParser) content(allowedNames map[string]bool) (content interface{}, err error) {
	var typeCodeRaw uint8
	if typeCodeRaw, err = jop.readUInt8(); err != nil {
		return
	}

	typeCode := typeCodeRaw - typeCodeMask
	if typeCode > maxTypeCode {
		// prevents reading unknown ("foreign") byte from the stream
		jop.rd.UnreadByte() //nolint:errcheck
		err = errors.Errorf("unknown type %#x", typeCodeRaw)
		return
	}

	name := typeNames[typeCode]
	if allowedNames != nil && !allowedNames[name] {
		err = errors.Errorf("%s not allowed here", name)
		return
	}

	parse, exists := knownParsers[name]
	if !exists {
		err = errors.Errorf("parsing %s is currently not supported", name)
		return
	}

	value, err := parse(jop)

	if valueMap, isMap := value.(map[string]interface{}); isMap {
		if valueMap[objectValueField] != nil {
			value = valueMap[objectValueField]
		} else {
			delete(valueMap, "@")
			delete(valueMap, "class")
			delete(valueMap, "extends")
		}
	}

	return value, err
}

// end check has next byte in stream.
func (jop *JavaObjectParser) end() bool {
	if jop.rd.Buffered() == 0 {
		_, eof := jop.rd.Peek(1)
		return eof != nil
	}

	return false
}

// readString reads a string of length cnt bytes.
func (jop *JavaObjectParser) readString(cnt int, asHex bool) (s string, err error) {
	jop.buf.Reset()

	// Prevented to allocate an extremely large block of memory.
	if cnt > jop.maxDataBlockSize {
		err = errors.Errorf("block data exceeds size of reader buffer. " +
			"To increase the size, use the method SetMaxDataBlockSize or use bufio.Reader with a larger buffer size")
		return
	}

	if _, err = io.CopyN(&jop.buf, jop.rd, int64(cnt)); err != nil {
		err = errors.Wrap(err, "error reading string")

		return
	}

	if asHex {
		s = hex.EncodeToString(jop.buf.Bytes())
	} else {
		s = jop.buf.String()
	}

	return
}

func (jop *JavaObjectParser) readUInt8() (x uint8, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading uint8")
	}

	return
}

func (jop *JavaObjectParser) readInt8() (x int8, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading int8")
	}

	return
}

func (jop *JavaObjectParser) readUInt16() (x uint16, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading uint16")
	}

	return
}

func (jop *JavaObjectParser) readInt16() (x int16, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading int16")
	}

	return
}

func (jop *JavaObjectParser) readUInt32() (x uint32, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading uint32")
	}

	return
}

func (jop *JavaObjectParser) readInt32() (x int32, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading int32")
	}

	return
}

func (jop *JavaObjectParser) readFloat32() (x float32, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading float32")
	}

	return
}

func (jop *JavaObjectParser) readInt64() (x int64, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading int64")
	}

	return
}

func (jop *JavaObjectParser) readFloat64() (x float64, err error) {
	if err = binary.Read(jop.rd, binary.BigEndian, &x); err != nil {
		err = errors.Wrap(err, "error reading float64")
	}

	return
}

// utf reads a variable length string.
func (jop *JavaObjectParser) utf() (s string, err error) {
	var offset uint16

	if offset, err = jop.readUInt16(); err != nil {
		err = errors.Wrap(err, "error reading utf: unable to read segment length")
		return
	}

	if s, err = jop.readString(int(offset), false); err != nil {
		err = errors.Wrap(err, "error reading utf: unable to read segment")
	}

	return
}

// utf reads a large (up to 2^32 bytes) variable length string.
func (jop *JavaObjectParser) utfLong() (s string, err error) {
	var offset uint32

	if offset, err = jop.readUInt32(); err != nil {
		err = errors.Wrap(err, "error reading utf: unable to read first segment length")
		return
	}

	if offset != 0 {
		err = errors.New("unable to read string larger than 2^32 bytes")
		return
	}

	if offset, err = jop.readUInt32(); err != nil {
		err = errors.Wrap(err, "error reading utf long: unable to read second segment length")
		return
	}

	if s, err = jop.readString(int(offset), false); err != nil {
		err = errors.Wrap(err, "error reading utf long: unable to read segment")
	}

	return
}

// magic checks for the presence of the magicNumber value.
func (jop *JavaObjectParser) magic() error {
	magicVal, err := jop.readUInt16()

	if err == nil && magicVal != magicNumber {
		return errors.New("magic number not found")
	}

	return err
}

// version checks to be sure the serialized object is using a supported protocol version.
func (jop *JavaObjectParser) version() error {
	ver, err := jop.readUInt16()
	if err != nil {
		return err
	}

	if ver != protocolVersion {
		return errors.Errorf("protocol version not recognized: wanted %d got %d", protocolVersion, ver)
	}

	return nil
}

// fieldDesc reads a single field descriptor.
func (jop *JavaObjectParser) fieldDesc() (f *field, err error) {
	var typeDec uint8

	if typeDec, err = jop.readUInt8(); err != nil {
		err = errors.Wrap(err, "error reading field type")
		return
	}

	var name string

	if name, err = jop.utf(); err != nil {
		err = errors.Wrap(err, "error reading field name")
		return
	}

	typeName := string(typeDec)

	f = &field{
		typeName: typeName,
		name:     name,
	}

	if strings.Contains("[L", typeName) {
		var className interface{}

		if className, err = jop.content(nil); err != nil {
			err = errors.Wrap(err, "error reading field class name")
			return
		}

		var isString bool
		if f.className, isString = className.(string); !isString {
			err = errors.New("unexpected field class name type")
		}
	}

	return
}

// annotations reads all class annotations.
func (jop *JavaObjectParser) annotations(allowedNames map[string]bool) (anns []interface{}, err error) {
	for {
		var ann interface{}
		if ann, err = jop.content(allowedNames); err != nil {
			err = errors.Wrap(err, "error reading class annotation")
			return
		}

		if _, isEndBlock := ann.(endBlockT); isEndBlock {
			break
		}

		anns = append(anns, ann)
	}

	return
}

// classDesc reads a class descriptor.
func (jop *JavaObjectParser) classDesc() (cls *clazz, err error) {
	var x interface{}
	if x, err = jop.content(allowedClazzNames); err != nil {
		err = errors.Wrap(err, "error reading class description")
		return
	}

	if x == nil {
		return
	}

	var isClazz bool
	if cls, isClazz = x.(*clazz); !isClazz {
		err = errors.New("unexpected type returned while reading class description")
	}

	return
}

// parseClassDesc parses a class descriptor.
func parseClassDesc(jop *JavaObjectParser) (x interface{}, err error) {
	cls := &clazz{}
	if cls.name, err = jop.utf(); err != nil {
		err = errors.Wrap(err, "error reading class name")
		return
	}

	if len(cls.name) < minClassNameLength {
		err = errors.Wrapf(err, "invalid class name: '%s'", cls.name)
		return
	}

	if cls.serialVersionUID, err = jop.readString(serialVersionUIDLength, true); err != nil {
		err = errors.Wrap(err, "error reading class serialVersionUID")
		return
	}

	jop.newHandle(cls)
	if cls.flags, err = jop.readUInt8(); err != nil {
		err = errors.Wrap(err, "error reading class flags")
		return
	}

	cls.isEnum = (cls.flags & 0x10) != 0
	var fieldCount uint16
	if fieldCount, err = jop.readUInt16(); err != nil {
		err = errors.Wrap(err, "error reading class field count")
		return
	}

	for i := 0; i < int(fieldCount); i++ {
		var f *field
		if f, err = jop.fieldDesc(); err != nil {
			err = errors.Wrap(err, "error reading class field")
			return
		}

		cls.fields = append(cls.fields, f)
	}

	if cls.annotations, err = jop.annotations(nil); err != nil {
		err = errors.Wrap(err, "error reading class annotations")
		return
	}

	if cls.super, err = jop.classDesc(); err != nil {
		err = errors.Wrap(err, "error reading class super")
		return
	}

	x = cls
	return
}

func parseClass(jop *JavaObjectParser) (cd interface{}, err error) {
	if cd, err = jop.classDesc(); err != nil {
		err = errors.Wrap(err, "error parsing class")
		return
	}

	cd = jop.newHandle(cd)
	return
}

func parseReference(jop *JavaObjectParser) (ref interface{}, err error) {
	var refIdx int32
	if refIdx, err = jop.readInt32(); err != nil {
		err = errors.Wrap(err, "error reading reference index")
		return
	}

	i := int(refIdx - refIdMask)
	if i > -1 && i < len(jop.handles) {
		ref = jop.handles[i]
		if ref == nil {
			ref = cycleValue
		}
	}

	return
}

func parseArray(jop *JavaObjectParser) (arr interface{}, err error) {
	var cls *clazz
	if cls, err = jop.classDesc(); err != nil {
		err = errors.Wrap(err, "error parsing array class")
		return
	}

	res := map[string]interface{}{
		"class": cls,
	}

	jop.newHandle(res)
	var size int32
	if size, err = jop.readInt32(); err != nil {
		err = errors.Wrap(err, "error reading array size")
		return
	}

	res["length"] = size
	if cls == nil {
		return
	}

	primHandler, exists := primitiveHandlers[string(cls.name[1])]
	if !exists {
		err = errors.Errorf("unknown field type '%s'", string(cls.name[1]))
		return
	}

	array := make([]interface{}, int(size))
	for i := 0; i < int(size); i++ {
		var nxt interface{}
		if nxt, err = primHandler(jop); err != nil {
			err = errors.Wrap(err, "error reading primitive array member")
			return
		}

		array[i] = nxt
	}

	arr = array
	return
}

// newDeferredHandle reserves an object handle slot and returns a func which can set the slot value at a later time.
func (jop *JavaObjectParser) newDeferredHandle() func(interface{}) interface{} {
	idx := len(jop.handles)
	jop.handles = append(jop.handles, nil)
	return func(obj interface{}) interface{} {
		jop.handles[idx] = obj
		return obj
	}
}

func parseEnum(jop *JavaObjectParser) (enum interface{}, err error) {
	var cls *clazz
	if cls, err = jop.classDesc(); err != nil {
		err = errors.Wrap(err, "error parsing enum class")
		return
	}

	deferredHandle := jop.newDeferredHandle()
	var enumConstant interface{}
	if enumConstant, err = jop.content(nil); err != nil {
		err = errors.Wrap(err, "error parsing enum constant")
		return
	}

	res := map[string]interface{}{
		objectValueField: enumConstant,
		"class":          cls,
	}

	enum = deferredHandle(res)
	return
}

func parseBlockData(jop *JavaObjectParser) (bd interface{}, err error) {
	var size uint8
	if size, err = jop.readUInt8(); err != nil {
		err = errors.Wrap(err, "error parsing block data size")
		return
	}

	data := make([]byte, size)
	if _, err = io.ReadFull(jop.rd, data); err == nil {
		bd = data
	}

	return
}

func parseBlockDataLong(jop *JavaObjectParser) (bdl interface{}, err error) {
	var size uint32
	if size, err = jop.readUInt32(); err != nil {
		err = errors.Wrap(err, "error parsing block data long size")
		return
	}

	// Prevented to allocate an extremely large block of memory.
	if int(size) > jop.maxDataBlockSize {
		err = errors.Errorf("block data exceeds size of reader buffer. " +
			"To increase the size, use the method SetMaxDataBlockSize or use bufio.Reader with a larger buffer size")
		return
	}

	data := make([]byte, size)
	if _, err = io.ReadFull(jop.rd, data); err == nil {
		bdl = data
	}

	return
}

func parseString(jop *JavaObjectParser) (str interface{}, err error) {
	if str, err = jop.utf(); err != nil {
		err = errors.Wrap(err, "error parsing string")
	} else {
		str = jop.newHandle(str)
	}
	return
}

func parseLongString(jop *JavaObjectParser) (longStr interface{}, err error) {
	if longStr, err = jop.utfLong(); err != nil {
		err = errors.Wrap(err, "error parsing long string")
	} else {
		jop.newHandle(longStr)
	}
	return
}

func parseNull(_ *JavaObjectParser) (interface{}, error) {
	return nil, nil
}

func parseEndBlockData(_ *JavaObjectParser) (interface{}, error) {
	return endBlock, nil
}

// values reads primitive field values.
func (jop *JavaObjectParser) values(cls *clazz) (vals map[string]interface{}, err error) {
	var exists bool
	var handler primitiveHandler
	vals = make(map[string]interface{})

	for _, field := range cls.fields {
		if field == nil {
			continue
		}

		if handler, exists = primitiveHandlers[field.typeName]; !exists {
			err = errors.Errorf("unknown field type '%s'", field.typeName)
			return
		}

		if vals[field.name], err = handler(jop); err != nil {
			err = errors.Wrap(err, "error reading primitive field value")
			return
		}
	}

	return
}

// classData reads a serialized class into a generic data structure.
func (jop *JavaObjectParser) classData(cls *clazz) (data map[string]interface{}, err error) {
	if cls == nil {
		return nil, errors.New("invalid class definition: nil")
	}

	flags := cls.flags & classFlagsMask
	if flags == scExternalizeWithBlockData {
		return nil, errors.New("unable to parse version 1 external content")
	}

	if flags != scSerializableWithoutWriteMethod && flags != scSerializableWithWriteMethod && flags != scExternalizeWithoutBlockData {
		return nil, errors.Errorf("unable to deserialize class with flags %#x", cls.flags)
	}

	var anns []interface{}
	data = make(map[string]interface{})

	if flags == scSerializableWithoutWriteMethod || flags == scSerializableWithWriteMethod {
		if data, err = jop.values(cls); err != nil {
			err = errors.Wrap(err, "error reading class data field values")

			return
		}
	}

	if flags == scSerializableWithWriteMethod || flags == scExternalizeWithoutBlockData {
		if anns, err = jop.annotations(nil); err != nil {
			err = errors.Wrap(err, "error reading annotations")

			return
		}
		data["@"] = anns
	}

	if postproc, exists := knownPostProcs[cls.name+"@"+cls.serialVersionUID]; exists {
		data, err = postproc(data, anns)
	}
	return
}

// recursiveClassData recursively reads inheritance tree until it reaches "java.lang.Object".
func (jop *JavaObjectParser) recursiveClassData(cls *clazz, obj map[string]interface{},
	seen map[*clazz]bool) error {
	if cls == nil {
		return nil
	}

	seen[cls] = true
	if cls.super != nil && !seen[cls.super] {
		seen[cls.super] = true
		if err := jop.recursiveClassData(cls.super, obj, seen); err != nil {
			return err
		}
	}

	extends, isMap := obj["extends"].(map[string]interface{})
	if !isMap {
		return errors.New("unexpected extends value")
	}

	fields, err := jop.classData(cls)
	if err != nil {
		return errors.Wrap(err, "error reading recursive class data")
	}

	extends[cls.name] = fields

	for name, val := range fields {
		obj[name] = val
	}

	return nil
}

func parseObject(jop *JavaObjectParser) (obj interface{}, err error) {
	var cls *clazz
	if cls, err = jop.classDesc(); err != nil {
		err = errors.Wrap(err, "error reading object class")

		return
	}

	objMap := map[string]interface{}{
		"class":   cls,
		"extends": make(map[string]interface{}),
	}

	deferredHandle := jop.newDeferredHandle()
	seen := map[*clazz]bool{}
	if err = jop.recursiveClassData(cls, objMap, seen); err != nil {
		err = errors.Wrap(err, "error reading recursive class data")
		return
	}

	obj = deferredHandle(objMap)
	return
}

// postProcSize reads the object size as an int32 from the first data element.
func postProcSize(data []interface{}, offset int) (size int, err error) {
	if len(data) < 1 {
		err = errors.New("invalid data: at least one element required")
		return
	}

	b, isByteSlice := data[0].([]byte)
	if !isByteSlice {
		err = errors.New("unexpected data at position 0")
		return
	}

	if len(b) < offset+objectDataMinLength {
		err = errors.Errorf("incorrect data at position 0: wanted at least %d bytes, got %d", offset+objectDataMinLength, len(b))
		return
	}

	var size32 int32
	if err = binary.Read(bytes.NewReader(b[offset:]), binary.BigEndian, &size32); err != nil {
		err = errors.Wrap(err, "error reading size")
		return
	}

	size = int(size32)
	return
}

// primObjectPostProc populates the object value with "value" field.
func primObjectPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	fields[objectValueField] = fields["value"]
	return fields, nil
}

// listPostProc populates the object value with a []interface{}.
func listPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	size, err := postProcSize(data, 0)
	if err != nil {
		return nil, err
	}

	if len(data) != size+1 {
		return nil, errors.Errorf("incorrect number of elements: want %d got %d", size, len(data)-1)
	}

	fields[objectValueField] = data[1 : size+1]
	return fields, err
}

// mapPostProc populates the object value with a map of key/value pairs.
func mapPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	size, err := postProcSize(data, 4)
	if err != nil {
		return nil, err
	}

	if size*2+1 > len(data) {
		return nil, errors.Errorf("incorrect number of elements: want %d got %d", size, len(data)-1)
	}

	m := make(map[string]interface{})

	for i := 0; i < size; i++ {
		key := data[2*i+1]
		value := data[2*i+2]
		m[fmt.Sprint(key)] = value
	}

	fields[objectValueField] = m
	return fields, nil
}

// enumMapPostProc populates the object value with a map of key/value pairs where keys are enum constants.
func enumMapPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	size, err := postProcSize(data, 0)
	if err != nil {
		return nil, err
	}

	if size*2+1 > len(data) {
		return nil, errors.Errorf("incorrect number of elements: want %d got %d", size, len(data)-1)
	}

	m := make(map[string]interface{})

	for i := 0; i < size; i++ {
		key := data[2*i+1]
		value := data[2*i+2]
		m[fmt.Sprint(key)] = value
	}

	fields[objectValueField] = m
	return fields, nil
}

// hashSetPostProc populates the object values with a []interface{}.
func hashSetPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	size, err := postProcSize(data, 8)
	if err != nil {
		return nil, err
	}

	if len(data) != size+1 {
		return nil, errors.Errorf("incorrect number of elements: want %d got %d", size, len(data)-1)
	}

	m := make([]interface{}, size)

	for idx := range data[1 : size+1] {
		m[idx] = data[idx+1]
	}

	fields[objectValueField] = m
	return fields, nil
}

// datePostProc populates the object value with a time.Time.
func datePostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	if len(data) < 1 {
		return nil, errors.New("invalid data: at least one element required")
	}

	b, isByteSlice := data[0].([]byte)
	if !isByteSlice {
		return nil, errors.New("unexpected data at position 0")
	}

	if len(b) < timestampBlockSize {
		return nil, errors.Errorf("incorrect data at position 0: wanted 8 bytes, got %d", len(b))
	}

	var timestamp int64
	if err := binary.Read(bytes.NewReader(b[0:timestampBlockSize]), binary.BigEndian, &timestamp); err != nil {
		return nil, errors.Wrap(err, "error reading timestamp")
	}

	fields[objectValueField] = time.Unix(0, timestamp*int64(time.Millisecond))
	return fields, nil
}

// calendarPostProc populates the object value with a time.Time.
func calendarPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	fields[objectValueField] = time.Unix(0, fields["time"].(int64)*int64(time.Millisecond))
	return fields, nil
}

// arraysArrayListPostProc populates the object value with "a" field.
func arraysArrayListPostProc(fields map[string]interface{}, data []interface{}) (map[string]interface{}, error) {
	fields[objectValueField] = fields["a"]
	return fields, nil
}
