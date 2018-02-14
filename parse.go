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

import "regexp"
import "fmt"

var comment = regexp.MustCompile(`^[\t ]*\#`)
var preprocess = regexp.MustCompile(`(?:\".*?\"|\'.*?\'|[\t ]*\#[^\n]*)`)

func replComment(s string) string {
	if comment.MatchString(s) { return "" }
	return s
}
func replCommentB(s []byte) []byte {
	if comment.Match(s) { return nil }
	return s
}

type DeComment struct{}

func (d DeComment) OfString(s string) string {
	return preprocess.ReplaceAllStringFunc(s,replComment)
}
func (d DeComment) OfBytes(s []byte) []byte {
	return preprocess.ReplaceAllFunc(s,replCommentB)
}

type ContentHandler interface{
	StartElement(clazz, word []byte) ContentHandler
	EndElement()
	KeyValuePair(key, value []byte)
}

type contentHandler struct{}
var vContentHandler ContentHandler = contentHandler{}
func DefaultContentHandler() ContentHandler { return vContentHandler }
func chDef(ch ContentHandler) ContentHandler {
	if ch==nil { return vContentHandler }
	return ch
}

func (c contentHandler) StartElement(clazz, word []byte) ContentHandler { return nil }
func (c contentHandler) EndElement() {}
func (c contentHandler) KeyValuePair(key, value []byte) {}

var elemenEx = regexp.MustCompile(`^\s*(\S+)\s+(\S+)\s*{`/*}*/)
var element  = regexp.MustCompile(`^\s*(\S+)\s*{`/*}*/)
var elemEnd  = regexp.MustCompile(/*{*/`^\s*}`)

var kvpair1 = regexp.MustCompile(`^\s*(\S+)\:\s+\"(\S+)\"`)
var kvpair2 = regexp.MustCompile(`^\s*(\S+)\:\s+\'(\S+)\'`)
var kvpair = regexp.MustCompile(`^\s*(\S+)\:\s+(\S+)`)

func parse(b []byte,ch ContentHandler) error {
	stack := make([]ContentHandler,0,16)
	
	for {
		idx := elemenEx.FindSubmatchIndex(b)
		if len(idx)!=0 {
			clazz := b[idx[2]:idx[3]]
			word  := b[idx[4]:idx[5]]
			nch := chDef(ch.StartElement(clazz,word))
			stack = append(stack,ch)
			ch = nch
			b = b[idx[1]:]
			continue
		}
		idx = element.FindSubmatchIndex(b)
		if len(idx)!=0 {
			clazz := b[idx[2]:idx[3]]
			nch := ch.StartElement(clazz,nil)
			stack = append(stack,ch)
			ch = nch
			b = b[idx[1]:]
			continue
		}
		idx = elemEnd.FindSubmatchIndex(b)
		if len(idx)!=0 {
			if len(stack)>0 {
				ch.EndElement()
				ch = stack[len(stack)-1]
				stack = stack[:len(stack)-1]
			} else {
				return fmt.Errorf(/*{*/"Unexpected '}'")
			}
			b = b[idx[1]:]
			continue
		}
		idx = kvpair1.FindSubmatchIndex(b)
		if len(idx)==0 { idx = kvpair2.FindSubmatchIndex(b) }
		if len(idx)==0 { idx = kvpair.FindSubmatchIndex(b) }
		if len(idx)!=0 {
			k := b[idx[2]:idx[3]]
			v := b[idx[4]:idx[5]]
			ch.KeyValuePair(k,v)
			b = b[idx[1]:]
			continue
		}
		if len(stack)>0 {
			return fmt.Errorf("unexpected EOF")
		}
		ch.EndElement()
		return nil
	}
	panic("unreachable")
}

func Parse(b []byte,ch ContentHandler) error {
	c := preprocess.ReplaceAllFunc(b,replCommentB)
	return parse(c,ch)
}

