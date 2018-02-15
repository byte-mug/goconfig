/*
Copyright (c) 2018 Simon Schmidt

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/


package goconfig

/*
type ContentHandler interface{
	StartElement(clazz, word []byte) ContentHandler
	EndElement()
	KeyValuePair(key, value []byte)
}
*/
import "reflect"
import "regexp"
import "encoding"
import "errors"
import "strconv"
import "strings"

var reflectPurify = regexp.MustCompile(`^[\$\%\@]`)
var reflectPurify2 = regexp.MustCompile(`[\!]$`)

/* Internal use only. */
var EReflectDecodeValueError = errors.New("EReflectDecodeValueError")

func reflectContains(s string, b byte) bool {
	for _,c := range []byte(s) {
		if b==c { return true }
	}
	return false
}

func reflectEat(v reflect.Value) {}

var reflectTypeTextUnmarshaler = reflect.ValueOf(new(encoding.TextUnmarshaler)).Type().Elem()

func reflectDecodePrinizpial(v reflect.Value, val []byte) error {
	if v.Type().Implements(reflectTypeTextUnmarshaler) {
		return v.Interface().(encoding.TextUnmarshaler).UnmarshalText(val)
	}
	if reflect.PtrTo(v.Type()).Implements(reflectTypeTextUnmarshaler) {
		if v.CanAddr() {
			return v.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText(val)
		}
		v2 := reflect.New(v.Type())
		e := v2.Interface().(encoding.TextUnmarshaler).UnmarshalText(val)
		if e==nil { v.Set(v2.Elem()) }
		return e
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			tni,e := strconv.ParseInt(string(val),0,64)
			if e!=nil { return e }
			v.SetInt(tni)
			return nil
		}
	case reflect.Uint,reflect.Uint8,reflect.Uint16,reflect.Uint32,reflect.Uint64:
		{
			tni,e := strconv.ParseUint(string(val),0,64)
			if e!=nil { return e }
			v.SetUint(tni)
			return nil
		}
	case reflect.Float32,reflect.Float64:
		{
			f,e := strconv.ParseFloat(string(val),64)
			if e!=nil { return e }
			v.SetFloat(f)
			return nil
		}
	case reflect.Bool:
		{
			b := false
			switch strings.ToLower(string(val)) {
			case "true","t","yes","y": b = true
			case "false","f","no","n": b = false
			default: return EReflectDecodeValueError
			}
			v.SetBool(b)
		}
	case reflect.String:
		v.SetString(string(val))
		return nil
	}
	return EReflectDecodeValueError
}

func reflectDecodeKey(v reflect.Value, val []byte) error {
	return reflectDecodePrinizpial(v,val)
}
func reflectDecodeValue(v reflect.Value, val []byte) error {
	return reflectDecodePrinizpial(v,val)
}

func CreateReflectHandler(i interface{}) ContentHandler {
	return reflectSpawnHandlerInner(reflect.Indirect(reflect.ValueOf(i)),reflectEat,false,false,nil,nil)
}
/*
func reflectSpawnHandler(v reflect.Value, eat func(v reflect.Value), clazz, word []byte) ContentHandler {
	return reflectSpawnHandlerInner(v,eat,true,true,clazz,word)
}
*/
func reflectSpawnHandler2(v reflect.Value, eat func(v reflect.Value), clazz, word []byte, propagate bool) ContentHandler {
	return reflectSpawnHandlerInner(v,eat,true,propagate,clazz,word)
}
func reflectSpawnHandlerInner(v reflect.Value, eat func(v reflect.Value), isprop bool, propagateWord bool, clazz, word []byte) ContentHandler {
	start:
	switch v.Type().Kind() {
	case reflect.Ptr:
		v2 := reflect.New(v.Type().Elem())
		v.Set(v2)
		eat(v)
		v = v.Elem()
		eat = reflectEat
		goto start
	case reflect.Struct:{
		rh := new(reflectHandler)
		rh.v = v
		rh.t = v.Type()
		rh.eat = eat
		n := rh.t.NumField()
		rh.fieldsIdx    = make(map[string]int)
		rh.fieldsSigil  = make([]int,n)
		rh.fieldsSuffix = make([]string,n)
		for i:=0;i<n;i++ {
			sf := rh.t.Field(i)
			on := sf.Name
			nn := sf.Tag.Get("inn")
			nn1 := reflectPurify.FindString(nn)
			nn = reflectPurify.ReplaceAllString(nn, "")
			if nn1=="" { rh.fieldsSigil[i]='$' } else { rh.fieldsSigil[i]=int(nn1[0]&0xff) }
			rh.fieldsSuffix[i] = reflectPurify2.FindString(nn)
			nn = reflectPurify2.ReplaceAllString(nn, "")
			if nn!="" { on = nn }
			rh.fieldsIdx[on] = i
		}
		if isprop && propagateWord {
			rh.KeyValuePair(clazz,word)
		}
		return rh
		}
	case reflect.Map:{
		if v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
		rh := new(reflectHandler)
		rh.v = v
		rh.t = v.Type()
		rh.eat = eat
		rh.fieldsIdx    = make(map[string]int)
		rh.isMap = true
		if isprop && propagateWord {
			rh.KeyValuePair(clazz,word)
		}
		return rh
		}
	}
	return nil
}

type reflectHandler struct{
	v reflect.Value
	t reflect.Type
	eat func(v reflect.Value)
	fieldsIdx map[string]int
	fieldsSigil []int
	fieldsSuffix []string
	isMap bool
}

func (r *reflectHandler) StartElement(clazz, word []byte) ContentHandler {
	if i,ok := r.fieldsIdx[string(clazz)] ; ok {
		propagate := !reflectContains(r.fieldsSuffix[i],'!')
		if len(word)==0 { propagate = false }
		switch r.fieldsSigil[i] {
		case '%':{
				mv := r.v.Field(i)
				if mv.IsNil() {
					mv.Set(reflect.MakeMap(r.t.Field(i).Type))
				}
				mt := r.t.Field(i).Type
				k := reflect.New(mt.Key()).Elem()
				v := reflect.New(mt.Elem()).Elem()
				e := reflectDecodeKey  (k,word)
				if e!=nil { return nil }
				return reflectSpawnHandler2(v,func(v2 reflect.Value){
					r.v.Field(i).SetMapIndex(k,v2)
				},clazz,word,propagate)
			}
		case '@':{
				mt := r.t.Field(i).Type
				v := reflect.New(mt.Elem()).Elem()
				return reflectSpawnHandler2(v,func(v2 reflect.Value){
					pv := r.v.Field(i)
					pv.Set(reflect.Append(pv,v2))
				},clazz,word,propagate)
			}
		case '$':
			return reflectSpawnHandler2(r.v.Field(i),reflectEat,clazz,word,propagate)
		}
		return nil
	}
	return nil
}

func (r *reflectHandler) EndElement() {
	r.eat(r.v)
}

func (r *reflectHandler) KeyValuePair(key, value []byte) {
	if r.isMap {
		k := reflect.New(r.t.Key()).Elem()
		v := reflect.New(r.t.Elem()).Elem()
		e := reflectDecodeKey  (k,key  )
		if e!=nil { return }
		e = reflectDecodeValue(v,value)
		if e!=nil { return }
		r.v.SetMapIndex(k,v)
		return
	}
	if i,ok := r.fieldsIdx[string(key)] ; ok {
		switch r.fieldsSigil[i] {
		case '@':{
				v := reflect.New(r.t.Field(i).Type).Elem()
				e := reflectDecodeValue(v,value)
				if e!=nil { return }
				pv := r.v.Field(i)
				pv.Set(reflect.Append(pv,v))
			}
		case '$':
			reflectDecodeValue(r.v.Field(i),value)
		}
		return
	}
}


