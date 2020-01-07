package sreq

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	// Version of sreq.
	Version = "0.8.7"

	defaultUserAgent = "go-sreq/" + Version
)

var (
	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type (
	// Values maps a string key to an interface{} type value,
	// It's typically used for request query parameters, form data and headers.
	Values map[string]interface{}

	// Params is an alias of Values, used for for request query parameters.
	Params = Values

	// Form is an alias of Values, used for request form values.
	Form = Values

	// Headers is an alias of Values, used for request headers.
	Headers = Values

	// Files maps a string key to a *File type value, used for files of multipart payload.
	Files map[string]*File

	// File specifies a file.
	// To upload a file you must specify its Filename field, otherwise sreq will raise a *RequestError.
	// If you don't specify the MIME field, sreq will detect automatically using http.DetectContentType.
	File struct {
		Filename string
		Body     io.Reader
		MIME     string
	}

	// H is a shortcut for map[string]interface{}, used for JSON unmarshalling.
	// Do not use it for other purposes!
	H map[string]interface{}

	// Number is a shortcut for float64.
	Number float64
)

func acquireBuffer() *bytes.Buffer {
	return bufPool.Get().(*bytes.Buffer)
}

func releaseBuffer(buf *bytes.Buffer) {
	if buf != nil {
		buf.Reset()
		bufPool.Put(buf)
	}
}

// Get gets the value associated with key.
func (v Values) Get(key string) interface{} {
	if v == nil {
		return nil
	}

	return v[key]
}

// Set sets the key to value. It replaces any existing values.
func (v Values) Set(key string, value interface{}) {
	v[key] = value
}

// SetDefault sets the key to value if the value not exists.
func (v Values) SetDefault(key string, value interface{}) {
	_, ok := v[key]
	if !ok {
		v.Set(key, value)
	}
}

// Del deletes the values associated with key.
func (v Values) Del(key string) {
	delete(v, key)
}

// Update merges v2 into v. It replaces any existing values.
func (v Values) Update(v2 Values) {
	for key, value := range v2 {
		v.Set(key, value)
	}
}

// Merge merges v2 into v. It keeps the existing values.
func (v Values) Merge(v2 Values) {
	for key, value := range v2 {
		v.SetDefault(key, value)
	}
}

// Clone returns a copy of the actual data to be used for request query parameters, form data or headers from v.
func (v Values) Clone() map[string][]string {
	if v == nil {
		return nil
	}

	res := make(map[string][]string, len(v))
	for k := range v {
		switch v := v[k].(type) {
		case []string:
			vs := make([]string, len(v))
			copy(vs, v)
			res[k] = vs
		default:
			res[k] = []string{toString(v, "")}
		}
	}
	return res
}

// Marshal returns the JSON encoding of v.
func (v Values) Marshal() string {
	s := toJSON(v, "", "", false)
	return strings.TrimSuffix(s, "\n")
}

// Encode encodes v into URL-unescaped form sorted by key.
// Only use when you consider Values as request query parameters or form data.
func (v Values) Encode() string {
	vv := v.Clone()
	keys := make([]string, 0, len(vv))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		vs := vv[k]
		for _, v := range vs {
			if sb.Len() > 0 {
				sb.WriteByte('&')
			}
			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	return sb.String()
}

// Float64 converts n to a float64.
func (n Number) Float64() float64 {
	return float64(n)
}

// Float32 converts n to a float32.
func (n Number) Float32() float32 {
	return float32(n)
}

// Int converts n to an int.
func (n Number) Int() int {
	return int(n)
}

// Int64 converts n to an int64.
func (n Number) Int64() int64 {
	return int64(n)
}

// Int32 converts n to an int32.
func (n Number) Int32() int32 {
	return int32(n)
}

// Int16 converts n to an int16.
func (n Number) Int16() int16 {
	return int16(n)
}

// Int8 converts n to an int8.
func (n Number) Int8() int8 {
	return int8(n)
}

// Uint converts n to a uint.
func (n Number) Uint() uint {
	return uint(n)
}

// Uint64 converts n to a uint64.
func (n Number) Uint64() uint64 {
	return uint64(n)
}

// Uint32 converts n to a uint32.
func (n Number) Uint32() uint32 {
	return uint32(n)
}

// Uint16 converts n to a uint16.
func (n Number) Uint16() uint16 {
	return uint16(n)
}

// Uint8 converts n to a uint8.
func (n Number) Uint8() uint8 {
	return uint8(n)
}

// String converts n to a string.
func (n Number) String() string {
	return strconv.FormatFloat(n.Float64(), 'f', -1, 64)
}

// Get gets the interface{} value associated with key.
func (h H) Get(key string) interface{} {
	if h == nil {
		return nil
	}

	return h[key]
}

// GetH gets the H value associated with key.
func (h H) GetH(key string) H {
	if h == nil {
		return nil
	}

	v, _ := h[key].(map[string]interface{})
	return v
}

// GetStringDefault gets the string value associated with key.
// The defaultValue is returned if the key not exists.
func (h H) GetStringDefault(key string, defaultValue string) string {
	if h == nil {
		return defaultValue
	}

	v, ok := h[key].(string)
	if !ok {
		return defaultValue
	}

	return v
}

// GetString gets the string value associated with key.
// The zero value is returned if the key not exists.
func (h H) GetString(key string) string {
	return h.GetStringDefault(key, "")
}

// GetBoolDefault gets the bool value associated with key.
// The defaultValue is returned if the key not exists.
func (h H) GetBoolDefault(key string, defaultValue bool) bool {
	if h == nil {
		return defaultValue
	}

	v, ok := h[key].(bool)
	if !ok {
		return defaultValue
	}

	return v
}

// GetBool gets the bool value associated with key.
// The zero value is returned if the key not exists.
func (h H) GetBool(key string) bool {
	return h.GetBoolDefault(key, false)
}

// GetNumberDefault gets the Number value associated with key.
// The defaultValue is returned if the key not exists.
func (h H) GetNumberDefault(key string, defaultValue Number) Number {
	if h == nil {
		return defaultValue
	}

	v, ok := h[key].(float64)
	if !ok {
		return defaultValue
	}

	return Number(v)
}

// GetNumber gets the Number value associated with key.
// The zero value is returned if the key not exists.
func (h H) GetNumber(key string) Number {
	return h.GetNumberDefault(key, Number(0))
}

// GetSlice gets the []interface{} value associated with key.
func (h H) GetSlice(key string) []interface{} {
	if h == nil {
		return nil
	}

	v, _ := h[key].([]interface{})
	return v
}

// GetHSlice gets the []H value associated with key.
func (h H) GetHSlice(key string) []H {
	v := h.GetSlice(key)
	vs := make([]H, 0, len(v))

	for _, vv := range v {
		if vv, ok := vv.(map[string]interface{}); ok {
			vs = append(vs, vv)
		}
	}
	return vs
}

// GetStringSlice gets the []string value associated with key.
func (h H) GetStringSlice(key string) []string {
	v := h.GetSlice(key)
	vs := make([]string, 0, len(v))

	for _, vv := range v {
		if vv, ok := vv.(string); ok {
			vs = append(vs, vv)
		}
	}
	return vs
}

// GetBoolSlice gets the []bool value associated with key.
func (h H) GetBoolSlice(key string) []bool {
	v := h.GetSlice(key)
	vs := make([]bool, 0, len(v))

	for _, vv := range v {
		if vv, ok := vv.(bool); ok {
			vs = append(vs, vv)
		}
	}
	return vs
}

// GetNumberSlice gets the []Number value associated with key.
func (h H) GetNumberSlice(key string) []Number {
	v := h.GetSlice(key)
	vs := make([]Number, 0, len(v))
	for _, vv := range v {
		if vv, ok := vv.(float64); ok {
			vs = append(vs, Number(vv))
		}
	}
	return vs
}

// Decode encodes h to JSON and then decodes to the output structure.
func (h H) Decode(output interface{}) error {
	b, err := jsonMarshal(h, "", "", false)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, output)
}

// String returns the JSON-encoded text representation of h.
func (h H) String() string {
	return toJSON(h, "", "\t", false)
}

// Get gets the value associated with key.
func (f Files) Get(key string) *File {
	if f == nil {
		return nil
	}

	return f[key]
}

// Set sets the key to value. It replaces any existing values.
func (f Files) Set(key string, value *File) {
	f[key] = value
}

// Del deletes the values associated with key.
func (f Files) Del(key string) {
	delete(f, key)
}

// NewFile returns a *File instance given a filename and its body.
func NewFile(filename string, body io.Reader) *File {
	return &File{
		Filename: filename,
		Body:     body,
	}
}

// SetFilename sets Filename field value of f.
func (f *File) SetFilename(filename string) *File {
	f.Filename = filename
	return f
}

// SetMIME sets MIME field value of f.
func (f *File) SetMIME(mime string) *File {
	f.MIME = mime
	return f
}

// Read implements Reader interface.
func (f *File) Read(p []byte) (int, error) {
	if f.Body == nil {
		return 0, io.EOF
	}
	return f.Body.Read(p)
}

// Close implements Closer interface.
func (f *File) Close() error {
	rc, ok := f.Body.(io.Closer)
	if !ok || rc == nil {
		return nil
	}

	return rc.Close()
}

// Open opens the named file and returns a *File instance.
func Open(filename string) (*File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewFile(filepath.Base(filename), file), nil
}

// MustOpen opens the named file and returns a *File instance.
// If there is an error, it will panic.
func MustOpen(filename string) *File {
	file, err := Open(filename)
	if err != nil {
		panic(err)
	}

	return file
}

func toString(v interface{}, defaultValue string) string {
	switch v := v.(type) {
	case string:
		return v
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case error:
		return v.Error()
	case interface {
		String() string
	}:
		return v.String()
	default:
		return defaultValue
	}
}

func toJSON(v interface{}, prefix string, indent string, escapeHTML bool) string {
	b, err := jsonMarshal(v, prefix, indent, escapeHTML)
	if err != nil {
		return "{}"
	}

	return string(b)
}

func jsonMarshal(v interface{}, prefix string, indent string, escapeHTML bool) ([]byte, error) {
	buf := acquireBuffer()
	defer releaseBuffer(buf)

	encoder := json.NewEncoder(buf)
	encoder.SetIndent(prefix, indent)
	encoder.SetEscapeHTML(escapeHTML)
	err := encoder.Encode(v)
	return buf.Bytes(), err
}
