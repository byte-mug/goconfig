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

var reflectPurify = regexp.MustCompile(`^[\$\%\@]`)
var reflectPurify2 = regexp.MustCompile(`[\!]$`)

const (
)

func reflectEat(v reflect.Value) {}

func decodeKey(v reflect.Value, val []byte) {
	v.SetString(string(val))
}
func decodeValue(v reflect.Value, val []byte) {
	v.SetString(string(val))
}
func spawnHandler(v reflect.Value, eat func(v reflect.Value), clazz, word []byte) ContentHandler {
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
			if on=="" { on = nn }
			rh.fieldsIdx[on] = i
		}
		rh.KeyValuePair(clazz,word)
		return rh
		}
	case reflect.Map:{
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
				mt := r.t.Field(i).Type
				k := reflect.New(mt.Key()).Elem()
				v := reflect.New(mt.Elem()).Elem()
				decodeKey  (k,word)
				return spawnHandler(v,func(v2 reflect.Value){
					r.v.Field(i).SetMapIndex(k,v2)
				},clazz,word)
			}
		case '@':{
				mt := r.t.Field(i).Type
				v := reflect.New(mt.Elem()).Elem()
				return spawnHandler(v,func(v2 reflect.Value){
					pv := r.v.Field(i)
					pv.Set(reflect.Append(pv,v2))
				},clazz,word)
			}
		case '$':
			return spawnHandler(r.v.Field(i),reflectEat,clazz,word)
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
		decodeKey  (k,key  )
		decodeValue(v,value)
		r.v.SetMapIndex(k,v)
		return
	}
	if i,ok := r.fieldsIdx[string(key)] ; ok {
		switch r.fieldsSigil[i] {
		case '@':{
				v := reflect.New(r.t.Field(i).Type).Elem()
				decodeValue(v,value)
				pv := r.v.Field(i)
				pv.Set(reflect.Append(pv,v))
			}
		case '$':
			decodeValue(r.v.Field(i),value)
		}
		return
	}
}


