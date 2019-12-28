package sreq

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	// Version of sreq.
	Version = "0.7.13"

	defaultUserAgent = "go-sreq/" + Version
)

var (
	bufPool = &sync.Pool{New: func() interface{} { return &bytes.Buffer{} }}
)

type (
	// KV is the interface that defines a data type used by sreq in many cases.
	// The Keys method should return a slice of keys typed string.
	// The Get method should return a slice of values typed string associated with the given key.
	KV interface {
		Keys() []string
		Get(key string) []string
	}

	// Values maps a string key to an interface{} type value,
	// Its value supports string, bool, float64, float32, int, int64,.int32, uint, uint64, uint32
	// or their combination. sreq will convert to string automatically.
	// Used for query parameters and form values.
	Values map[string]interface{}

	// Params is an alias of Values, used for for request query parameters.
	Params = Values

	// Form is an alias of Values, used for request form values.
	Form = Values

	// Headers maps a string key to an interface{} type value,
	// Its value supports string, bool, float64, float32, int, int64,.int32, uint, uint64, uint32
	// or their combination. sreq will convert to string automatically.
	// Used for request headers.
	Headers map[string]interface{}

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
	H map[string]interface{}
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

// Get gets the value associated with the given key, ignore unsupported data type.
func (v Values) Get(key string) []string {
	if v == nil {
		return nil
	}

	return convert(v[key])
}

// Set sets the key to value. It replaces any existing values.
func (v Values) Set(key string, value interface{}) {
	v[key] = value
}

// Del deletes the values associated with key.
func (v Values) Del(key string) {
	delete(v, key)
}

// Keys returns the keys of v.
func (v Values) Keys() []string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	return keys
}

// Marshal returns the JSON encoding of v.
func (v Values) Marshal() string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// Encode encodes v into URL-unescaped form sorted by key.
func (v Values) Encode() string {
	var sb strings.Builder
	write(&sb, v, writeValues)
	return sb.String()
}

// String returns the text representation of v.
func (v Values) String() string {
	return v.Encode()
}

// Get gets the value associated with the given key, ignore unsupported data type.
func (h Headers) Get(key string) []string {
	if h == nil {
		return nil
	}

	return convert(h[key])
}

// Set sets the key to value. It replaces any existing values.
func (h Headers) Set(key string, value interface{}) {
	h[key] = value
}

// Del deletes the values associated with key.
func (h Headers) Del(key string) {
	delete(h, key)
}

// Keys returns the keys of h.
func (h Headers) Keys() []string {
	keys := make([]string, 0, len(h))
	for k := range h {
		keys = append(keys, k)
	}
	return keys
}

// String returns the text representation of h.
func (h Headers) String() string {
	var sb strings.Builder
	write(&sb, h, writeHeaders)
	return sb.String()
}

// Get gets the interface{} value associated with the given key.
func (h H) Get(key string) interface{} {
	if h == nil {
		return nil
	}

	return h[key]
}

// GetH gets the H value associated with the given key.
func (h H) GetH(key string) H {
	if h == nil {
		return nil
	}

	v, _ := h[key].(map[string]interface{})
	return v
}

// GetString gets the string value associated with the given key.
func (h H) GetString(key string) string {
	if h == nil {
		return ""
	}

	v, _ := h[key].(string)
	return v
}

// GetBool gets the bool value associated with the given key.
func (h H) GetBool(key string) bool {
	if h == nil {
		return false
	}

	v, _ := h[key].(bool)
	return v
}

// GetFloat64 gets the float64 value associated with the given key.
func (h H) GetFloat64(key string) float64 {
	if h == nil {
		return 0
	}

	v, _ := h[key].(float64)
	return v
}

// GetFloat32 gets the float32 value associated with the given key.
func (h H) GetFloat32(key string) float32 {
	return float32(h.GetFloat64(key))
}

// GetInt gets the int value associated with the given key.
func (h H) GetInt(key string) int {
	return int(h.GetFloat64(key))
}

// GetInt64 gets the int64 value associated with the given key.
func (h H) GetInt64(key string) int64 {
	return int64(h.GetFloat64(key))
}

// GetInt32 gets the int32 value associated with the given key.
func (h H) GetInt32(key string) int32 {
	return int32(h.GetFloat64(key))
}

// GetUint gets the uint value associated with the given key.
func (h H) GetUint(key string) uint {
	return uint(h.GetFloat64(key))
}

// GetUint64 gets the uint64 value associated with the given key.
func (h H) GetUint64(key string) uint64 {
	return uint64(h.GetFloat64(key))
}

// GetUint32 gets the uint32 value associated with the given key.
func (h H) GetUint32(key string) uint32 {
	return uint32(h.GetFloat64(key))
}

// GetSlice gets the []interface{} value associated with the given key.
func (h H) GetSlice(key string) []interface{} {
	if h == nil {
		return nil
	}

	v, _ := h[key].([]interface{})
	return v
}

// GetHSlice gets the []H value associated with the given key.
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

// GetStringSlice gets the []string value associated with the given key.
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

// GetBoolSlice gets the []bool value associated with the given key.
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

// GetFloat64Slice gets the []float64 value associated with the given key.
func (h H) GetFloat64Slice(key string) []float64 {
	v := h.GetSlice(key)
	vs := make([]float64, 0, len(v))

	for _, vv := range v {
		if vv, ok := vv.(float64); ok {
			vs = append(vs, vv)
		}
	}
	return vs
}

// GetFloat32Slice gets the []float32 value associated with the given key.
func (h H) GetFloat32Slice(key string) []float32 {
	v := h.GetFloat64Slice(key)
	vs := make([]float32, len(v))
	for i, vv := range v {
		vs[i] = float32(vv)
	}
	return vs
}

// GetIntSlice gets the []int value associated with the given key.
func (h H) GetIntSlice(key string) []int {
	v := h.GetFloat64Slice(key)
	vs := make([]int, len(v))
	for i, vv := range v {
		vs[i] = int(vv)
	}
	return vs
}

// GetInt64Slice gets the []int64 value associated with the given key.
func (h H) GetInt64Slice(key string) []int64 {
	v := h.GetFloat64Slice(key)
	vs := make([]int64, len(v))
	for i, vv := range v {
		vs[i] = int64(vv)
	}
	return vs
}

// GetInt32Slice gets the []int32 value associated with the given key.
func (h H) GetInt32Slice(key string) []int32 {
	v := h.GetFloat64Slice(key)
	vs := make([]int32, len(v))
	for i, vv := range v {
		vs[i] = int32(vv)
	}
	return vs
}

// GetUintSlice gets the []uint value associated with the given key.
func (h H) GetUintSlice(key string) []uint {
	v := h.GetFloat64Slice(key)
	vs := make([]uint, len(v))
	for i, vv := range v {
		vs[i] = uint(vv)
	}
	return vs
}

// GetUint64Slice gets the []uint64 value associated with the given key.
func (h H) GetUint64Slice(key string) []uint64 {
	v := h.GetFloat64Slice(key)
	vs := make([]uint64, len(v))
	for i, vv := range v {
		vs[i] = uint64(vv)
	}
	return vs
}

// GetUint32Slice gets the []uint32 value associated with the given key.
func (h H) GetUint32Slice(key string) []uint32 {
	v := h.GetFloat64Slice(key)
	vs := make([]uint32, len(v))
	for i, vv := range v {
		vs[i] = uint32(vv)
	}
	return vs
}

// String returns the JSON-encoded text representation of h.
func (h H) String() string {
	return toJSON(h)
}

// Get gets the value associated with the given key.
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

// Open opens the named file and returns a *File instance whose Filename is filename.
func Open(filename string) (*File, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewFile(filepath.Base(filename), file), nil
}

// MustOpen opens the named file and returns a *File instance whose Filename is filename.
// If there is an error, it will panic.
func MustOpen(filename string) *File {
	file, err := Open(filename)
	if err != nil {
		panic(err)
	}

	return file
}

func convertBoolSlice(v []bool) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatBool(vv)
	}
	return vs
}

func convertFloat64Slice(v []float64) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatFloat(vv, 'f', -1, 64)
	}
	return vs
}

func convertFloat32Slice(v []float32) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatFloat(float64(vv), 'f', -1, 32)
	}
	return vs
}

func convertIntSlice(v []int) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.Itoa(vv)
	}
	return vs
}

func convertInt64Slice(v []int64) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatInt(vv, 10)
	}
	return vs
}

func convertInt32Slice(v []int32) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatInt(int64(vv), 10)
	}
	return vs
}

func convertUintSlice(v []uint) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatUint(uint64(vv), 10)
	}
	return vs
}

func convertUint64Slice(v []uint64) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatUint(vv, 10)
	}
	return vs
}

func convertUint32Slice(v []uint32) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = strconv.FormatUint(uint64(vv), 10)
	}
	return vs
}

func convertSlice(v []interface{}) []string {
	vs := make([]string, len(v))
	for i, vv := range v {
		vs[i] = toString(vv)
	}
	return vs
}

func toString(v interface{}) string {
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
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	default:
		return ""
	}
}

func convert(v interface{}) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case bool, float64, float32, int, int64, int32, uint, uint64, uint32:
		return []string{toString(v)}
	case []string:
		return v
	case []bool:
		return convertBoolSlice(v)
	case []float64:
		return convertFloat64Slice(v)
	case []float32:
		return convertFloat32Slice(v)
	case []int:
		return convertIntSlice(v)
	case []int64:
		return convertInt64Slice(v)
	case []int32:
		return convertInt32Slice(v)
	case []uint:
		return convertUintSlice(v)
	case []uint64:
		return convertUint64Slice(v)
	case []uint32:
		return convertUint32Slice(v)
	case []interface{}:
		return convertSlice(v)
	default:
		return nil
	}
}

func writeValues(sb *strings.Builder, k string, v []string) {
	for _, vs := range v {
		if sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(vs)
	}
}

func writeHeaders(sb *strings.Builder, k string, v []string) {
	for _, vs := range v {
		if sb.Len() > 0 {
			sb.WriteString("\r\n")
		}
		sb.WriteString(http.CanonicalHeaderKey(k))
		sb.WriteString(": ")
		sb.WriteString(vs)
	}
}

func write(sb *strings.Builder, v KV, callback func(sb *strings.Builder, k string, v []string)) {
	keys := v.Keys()
	sort.Strings(keys)

	for _, k := range keys {
		callback(sb, k, v.Get(k))
	}
}

func toJSON(data interface{}) string {
	b, err := jsonMarshal(data, "", "\t", false)
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
