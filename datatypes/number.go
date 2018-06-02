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


package datatypes

import "strings"
import "strconv"

type Number string
func (n Number) me() (string,string) {
	o := string(n)
	if i := strings.Index(o,"<<"); 0<=i { return o[:i],o[i+2:] }
	return o,""
}
func (n Number) Int64() int64 {
	m,e := n.me()
	i,_ := strconv.ParseInt(m,0,64)
	u,_ := strconv.ParseUint(e,0,64)
	return i<<u
}
func (n Number) Uint64() uint64 {
	m,e := n.me()
	i,_ := strconv.ParseUint(m,0,64)
	u,_ := strconv.ParseUint(e,0,64)
	return i<<u
}


