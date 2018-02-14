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

var reflectPurify = regexp.MustCompile(`^[\$\%\@]`)
var reflectPurify2 = regexp.MustCompile(`[\!]$`)

var EReflectDecodeValueError = errors.New("EReflectDecodeValueError")

func reflectEat(v reflect.Value) {}

func reflectDecodePrinizpial(v reflect.Value, val []byte) error {
	i := v.Interface()
	if t,ok := i.(encoding.TextUnmarshaler) ; ok && t!=nil {
		return t.UnmarshalText(val)
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
	return reflectSpawnHandlerInner(reflect.Indirect(reflect.ValueOf(i)),reflectEat,false,nil,nil)
}
func reflectSpawnHandler(v reflect.Value, eat func(v reflect.Value), clazz, word []byte) ContentHandler {
	return reflectSpawnHandlerInner(v,eat,true,clazz,word)
}
func reflectSpawnHandlerInner(v reflect.Value, eat func(v reflect.Value), isprop bool, clazz, word []byte) ContentHandler {
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
			//nn1 = reflectPurify2.FindString(nn)
			//nn = reflectPurify2.ReplaceAllString(nn, "")
			if nn!="" { on = nn }
			rh.fieldsIdx[on] = i
		}
		rh.KeyValuePair(clazz,word)
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
		rh.KeyValuePair(clazz,word)
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
		switch r.fieldsSigil[i] {
		case '%':{
				mv := r.v.Field(i)
				if mv.IsNil() {
					mv.Set(reflect.MakeMap(r.t.Field(i).Type))
				}
				mt := r.t.Field(i).Type
				k := reflect.New(mt.Key()).Elem()
				v := reflect.New(mt.Elem()).Elem()
				reflectDecodeKey  (k,word)
				return reflectSpawnHandler(v,func(v2 reflect.Value){
					r.v.Field(i).SetMapIndex(k,v2)
				},clazz,word)
			}
		case '@':{
				mt := r.t.Field(i).Type
				v := reflect.New(mt.Elem()).Elem()
				return reflectSpawnHandler(v,func(v2 reflect.Value){
					pv := r.v.Field(i)
					pv.Set(reflect.Append(pv,v2))
				},clazz,word)
			}
		case '$':
			return reflectSpawnHandler(r.v.Field(i),reflectEat,clazz,word)
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
		reflectDecodeKey  (k,key  )
		reflectDecodeValue(v,value)
		r.v.SetMapIndex(k,v)
		return
	}
	if i,ok := r.fieldsIdx[string(key)] ; ok {
		switch r.fieldsSigil[i] {
		case '@':{
				v := reflect.New(r.t.Field(i).Type).Elem()
				reflectDecodeValue(v,value)
				pv := r.v.Field(i)
				pv.Set(reflect.Append(pv,v))
			}
		case '$':
			reflectDecodeValue(r.v.Field(i),value)
		}
		return
	}
}


