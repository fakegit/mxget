package sreq

import (
	"encoding/json"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	// Version of sreq.
	Version = "0.9.11"

	defaultUserAgent = "Go-sreq/" + Version
)

var (
	_ FormFile = (*File)(nil) // *File must implement FormFile interface
)

type (
	// Values maps a string key to an interface{} type value,
	// When it's used for request query parameters, form data or headers,
	// besides string and []string, its value also supports bool and number,
	// sreq will convert to string automatically.
	Values map[string]interface{}

	// Params is an alias of Values, used for for request query parameters.
	Params = Values

	// Form is an alias of Values, used for request form data.
	Form = Values

	// Headers is an alias of Values, used for request headers.
	Headers = Values

	// Cookies is a shortcut for map[string]string, used for request cookies.
	Cookies map[string]string

	// Files maps a string key to a FormFile type value, used for form files of multipart payload.
	Files map[string]FormFile

	// FormFile is the interface that specifies a form file of multipart payload.
	FormFile interface {
		io.ReadCloser
		Filename() string
		MIME() string
	}

	// File specifies a file to upload, it implements FormFile interface.
	// Note: To upload a file its filename must be specified, if not, sreq will use "file" as default.
	// If you don't specify the mime, sreq will detect automatically using http.DetectContentType.
	File struct {
		Body     io.Reader
		filename string
		mime     string
	}

	// H is a shortcut for map[string]interface{}, used for JSON unmarshalling.
	// Do not use it for other purposes!
	H map[string]interface{}

	// Number is a shortcut for float64.
	Number float64
)

// Get gets the equivalent request query parameter, form data or header value associated with key.
func (v Values) Get(key string) []string {
	value, ok := v[key]
	if !ok {
		return nil
	}

	switch vv := value.(type) {
	case []string:
		vs := make([]string, len(vv))
		copy(vs, vv)
		return vs
	default:
		return []string{toString(vv)}
	}
}

// Has checks if a key exists.
func (v Values) Has(key string) bool {
	_, ok := v[key]
	return ok
}

// Set sets the key to value. It replaces any existing values.
func (v Values) Set(key string, value interface{}) {
	v[key] = value
}

// SetDefault sets the key to value if the value not exists.
func (v Values) SetDefault(key string, value interface{}) {
	if !v.Has(key) {
		v.Set(key, value)
	}
}

// Del deletes the value associated with key.
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

// Decode translates v and returns the equivalent request query parameters, form data or headers.
func (v Values) Decode(canonicalHeaderKey bool) map[string][]string {
	vv := make(map[string][]string, len(v))
	for k := range v {
		if canonicalHeaderKey {
			vv[http.CanonicalHeaderKey(k)] = v.Get(k)
		} else {
			vv[k] = v.Get(k)
		}
	}
	return vv
}

// URLEncode encodes v into URL form sorted by key when v is considered as request query parameters or form data.
func (v Values) URLEncode(escaped bool) string {
	vv := v.Decode(false)
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

			if escaped {
				k = neturl.QueryEscape(k)
				v = neturl.QueryEscape(v)
			}

			sb.WriteString(k)
			sb.WriteByte('=')
			sb.WriteString(v)
		}
	}
	return sb.String()
}

// Marshal returns the JSON encoding of v.
func (v Values) Marshal() string {
	return toJSON(v, "", "", false)
}

// Get gets the equivalent request cookie associated with key.
func (c Cookies) Get(key string) *http.Cookie {
	value, ok := c[key]
	if !ok {
		return nil
	}

	return &http.Cookie{
		Name:  key,
		Value: value,
	}
}

// Has checks if a key exists.
func (c Cookies) Has(key string) bool {
	_, ok := c[key]
	return ok
}

// Set sets the key to value. It replaces any existing values.
func (c Cookies) Set(key string, value string) {
	c[key] = value
}

// SetDefault sets the key to value if the value not exists.
func (c Cookies) SetDefault(key string, value string) {
	if !c.Has(key) {
		c.Set(key, value)
	}
}

// Del deletes the value associated with key.
func (c Cookies) Del(key string) {
	delete(c, key)
}

// Update merges c2 into c. It replaces any existing values.
func (c Cookies) Update(c2 Cookies) {
	for key, value := range c2 {
		c.Set(key, value)
	}
}

// Merge merges c2 into c. It keeps the existing values.
func (c Cookies) Merge(c2 Cookies) {
	for key, value := range c2 {
		c.SetDefault(key, value)
	}
}

// Clone returns a copy of c or nil if c is nil.
func (c Cookies) Clone() Cookies {
	if c == nil {
		return nil
	}

	c2 := make(Cookies, len(c))
	for key, value := range c {
		c2.Set(key, value)
	}
	return c2
}

// Decode translates c and returns the equivalent request cookies.
func (c Cookies) Decode() []*http.Cookie {
	cookies := make([]*http.Cookie, 0, len(c))
	for k := range c {
		cookies = append(cookies, c.Get(k))
	}
	return cookies
}

// Get gets the value associated with key.
func (f Files) Get(key string) FormFile {
	if f == nil {
		return nil
	}

	return f[key]
}

// Set sets the key to value. It replaces any existing values.
func (f Files) Set(key string, value FormFile) {
	f[key] = value
}

// Del deletes the value associated with key.
func (f Files) Del(key string) {
	delete(f, key)
}

// NewFile returns a *File instance given the body and filename.
func NewFile(body io.Reader, filename string) *File {
	return &File{
		Body:     body,
		filename: filename,
	}
}

// Filename implements FormFile interface, returns the filename of f.
func (f *File) Filename() string {
	return f.filename
}

// MIME implements FormFile interface, returns the mine of f.
func (f *File) MIME() string {
	return f.mime
}

// SetFilename specifies the filename of f.
func (f *File) SetFilename(filename string) *File {
	f.filename = filename
	return f
}

// SetMIME specifies the mine of f.
func (f *File) SetMIME(mime string) *File {
	f.mime = mime
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

	return NewFile(file, filepath.Base(filename)), nil
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

// Has checks if a key exists.
func (h H) Has(key string) bool {
	_, ok := h[key]
	return ok
}

// GetH gets the H value associated with key.
func (h H) GetH(key string) H {
	v, _ := h[key].(map[string]interface{})
	return v
}

// GetStringDefault gets the string value associated with key.
// The defaultValue is returned if the key not exists.
func (h H) GetStringDefault(key string, defaultValue string) string {
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
	v, ok := h[key].(float64)
	if !ok {
		return defaultValue
	}

	return Number(v)
}

// GetNumber gets the Number value associated with key.
// The zero value is returned if the key not exists.
func (h H) GetNumber(key string) Number {
	return h.GetNumberDefault(key, 0)
}

// GetSlice gets the []interface{} value associated with key.
func (h H) GetSlice(key string) []interface{} {
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
// output must be a pointer.
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
