package goja_bindings

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
	"strings"
)

type doc struct {
	vm           *goja.Runtime
	doc          *goquery.Document
	docSelection *docSelection
}

type docSelection struct {
	doc       *doc
	selection *goquery.Selection
}

func setSelectionObjectProperties(obj *goja.Object, docS *docSelection) {
	_ = obj.Set("length", docS.Length)
	_ = obj.Set("html", docS.Html)
	_ = obj.Set("text", docS.Text)
	_ = obj.Set("attr", docS.Attr)
	_ = obj.Set("find", docS.Find)
	_ = obj.Set("children", docS.Children)
	_ = obj.Set("each", docS.Each)
	_ = obj.Set("text", docS.Text)
	_ = obj.Set("parent", docS.Parent)
	_ = obj.Set("parentsUntil", docS.ParentsUntil)
	_ = obj.Set("parents", docS.Parents)
	_ = obj.Set("end", docS.End)
	_ = obj.Set("closest", docS.Closest)
	_ = obj.Set("map", docS.Map)
	_ = obj.Set("first", docS.First)
	_ = obj.Set("last", docS.Last)
	_ = obj.Set("eq", docS.Eq)
	_ = obj.Set("contents", docS.Contents)
	_ = obj.Set("contentsFiltered", docS.ContentsFiltered)
	_ = obj.Set("filter", docS.Filter)
	_ = obj.Set("not", docS.Not)
	_ = obj.Set("is", docS.Is)
	_ = obj.Set("has", docS.Has)
	_ = obj.Set("next", docS.Next)
	_ = obj.Set("nextAll", docS.NextAll)
	_ = obj.Set("nextUntil", docS.NextUntil)
	_ = obj.Set("prev", docS.Prev)
	_ = obj.Set("prevAll", docS.PrevAll)
	_ = obj.Set("prevUntil", docS.PrevUntil)
	_ = obj.Set("siblings", docS.Siblings)
	_ = obj.Set("data", docS.Data)
	_ = obj.Set("attrs", docS.Attrs)
}

func BindDocument(vm *goja.Runtime) error {
	// Set Doc "class"
	err := vm.Set("Doc", func(call goja.ConstructorCall) *goja.Object {
		obj := call.This
		if len(call.Arguments) != 1 {
			return goja.Undefined().ToObject(vm)
		}
		html := call.Arguments[0].String()

		goqueryDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return goja.Undefined().ToObject(vm)
		}
		d := &doc{
			vm:  vm,
			doc: goqueryDoc,
			docSelection: &docSelection{
				doc:       nil,
				selection: goqueryDoc.Selection,
			},
		}
		d.docSelection.doc = d

		setSelectionObjectProperties(obj, d.docSelection)
		return obj
	})
	if err != nil {
		return err
	}

	// Set "LoadDoc" function
	err = vm.Set("LoadDoc", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			panic(vm.ToValue("missing argument"))
		}

		html := call.Arguments[0].String()
		goqueryDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			return goja.Null()
		}

		d := &doc{
			vm:  vm,
			doc: goqueryDoc,
			docSelection: &docSelection{
				doc:       nil,
				selection: goqueryDoc.Selection,
			},
		}
		d.docSelection.doc = d

		docSelectionFunction := func(call goja.FunctionCall) goja.Value {
			selectorStr, ok := call.Argument(0).Export().(string)
			if !ok {
				panic(vm.NewTypeError("argument is not a string").ToString())
			}
			return newDocSelectionGojaValue(d, d.doc.Find(selectorStr))
		}

		return vm.ToValue(docSelectionFunction)
	})

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Document
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func newDocSelectionGojaValue(d *doc, selection *goquery.Selection) goja.Value {
	ds := &docSelection{
		doc:       d,
		selection: selection,
	}

	obj := d.vm.NewObject()
	setSelectionObjectProperties(obj, ds)

	return obj
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Selection
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (s *docSelection) getFirstStringArg(call goja.FunctionCall) string {
	selectorStr, ok := call.Argument(0).Export().(string)
	if !ok {
		panic(s.doc.vm.NewTypeError("argument is not a string").ToString())
	}
	return selectorStr
}

func (s *docSelection) Length(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		return s.doc.vm.ToValue(0)
	}
	return s.doc.vm.ToValue(s.selection.Length())
}

// Find gets the descendants of each element in the current set of matched elements, filtered by a selector.
//
//	find(selector: string): DocSelection;
func (s *docSelection) Find(call goja.FunctionCall) (ret goja.Value) {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.Find(selectorStr))
}

func (s *docSelection) Html(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		return goja.Null()
	}
	htmlStr, err := s.selection.Html()
	if err != nil {
		return goja.Null()
	}
	return s.doc.vm.ToValue(htmlStr)
}

func (s *docSelection) Text(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		return s.doc.vm.ToValue("")
	}
	return s.doc.vm.ToValue(s.selection.Text())
}

// Attr gets the specified attribute's value for the first element in the Selection. To get the value for each element individually, use a
// looping construct such as Each or Map method.
//
//	attr(name: string): string | undefined;
func (s *docSelection) Attr(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	attr, found := s.selection.Attr(s.getFirstStringArg(call))
	if !found {
		return goja.Undefined()
	}
	return s.doc.vm.ToValue(attr)
}

// Attrs gets all attributes for the first element in the Selection.
//
//	attrs(): { [key: string]: string };
func (s *docSelection) Attrs(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	attrs := make(map[string]string)
	for _, v := range s.selection.Get(0).Attr {
		attrs[v.Key] = v.Val
	}
	return s.doc.vm.ToValue(attrs)
}

// Data gets data associated with the matched elements or return the value at the named data store for the first element in the set of matched elements.
//
//	data(name?: string): { [key: string]: string } | string | undefined;
func (s *docSelection) Data(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		var data map[string]string
		n := s.selection.Get(0)
		if n == nil {
			return goja.Undefined()
		}
		for _, v := range n.Attr {
			if strings.HasPrefix(v.Key, "data-") {
				if data == nil {
					data = make(map[string]string)
				}
				data[v.Key] = v.Val
			}
		}
		return s.doc.vm.ToValue(data)
	}

	name := call.Argument(0).String()
	n := s.selection.Get(0)
	if n == nil {
		return goja.Undefined()
	}

	data, found := s.selection.Attr(fmt.Sprintf("data-%s", name))
	if !found {
		return goja.Undefined()
	}

	return s.doc.vm.ToValue(data)
}

// Parent gets the parent of each element in the Selection. It returns a new Selection object containing the matched elements.
//
//	parent(selector?: string): DocSelection;
func (s *docSelection) Parent(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Parent())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.ParentFiltered(selectorStr))
}

// Parents gets the ancestors of each element in the current Selection. It returns a new Selection object with the matched elements.
//
//	parents(selector?: string): DocSelection;
func (s *docSelection) Parents(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Parents())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.ParentsFiltered(selectorStr))
}

// ParentsUntil gets the ancestors of each element in the Selection, up to but not including the element matched by the selector. It returns a
// new Selection object containing the matched elements.
//
//	parentsUntil(selector?: string, until?: string): DocSelection;
func (s *docSelection) ParentsUntil(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	selectorStr := s.getFirstStringArg(call)
	if len(call.Arguments) < 2 {
		return newDocSelectionGojaValue(s.doc, s.selection.ParentsUntil(selectorStr))
	}
	untilStr := call.Argument(1).String()
	return newDocSelectionGojaValue(s.doc, s.selection.ParentsFilteredUntil(selectorStr, untilStr))
}

// End ends the most recent filtering operation in the current chain and returns the set of matched elements to its previous state.
//
//	end(): DocSelection;
func (s *docSelection) End(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	return newDocSelectionGojaValue(s.doc, s.selection.End())
}

// Closest gets the first element that matches the selector by testing the element itself and traversing up through its ancestors in the DOM tree.
//
//	closest(selector?: string): DocSelection;
func (s *docSelection) Closest(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Closest(""))
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.Closest(selectorStr))
}

// Next gets the next sibling of each selected element, optionally filtered by a selector.
//
//	next(selector?: string): DocSelection;
func (s *docSelection) Next(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Next())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.NextFiltered(selectorStr))
}

// NextAll gets all following siblings of each element in the Selection, optionally filtered by a selector.
//
//	nextAll(selector?: string): DocSelection;
func (s *docSelection) NextAll(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.NextAll())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.NextAllFiltered(selectorStr))
}

// NextUntil  gets all following siblings of each element up to but not including the element matched by the selector.
//
//	nextUntil(selector: string, until?: string): DocSelection;
func (s *docSelection) NextUntil(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	selectorStr := s.getFirstStringArg(call)
	if len(call.Arguments) < 2 {
		return newDocSelectionGojaValue(s.doc, s.selection.NextUntil(selectorStr))
	}
	untilStr := call.Argument(1).String()
	return newDocSelectionGojaValue(s.doc, s.selection.NextFilteredUntil(selectorStr, untilStr))
}

// Prev gets the previous sibling of each selected element optionally filtered by a selector.
//
//	prev(selector?: string): DocSelection;
func (s *docSelection) Prev(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Prev())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.PrevFiltered(selectorStr))
}

// PrevAll gets all preceding siblings of each element in the Selection, optionally filtered by a selector.
//
//	prevAll(selector?: string): DocSelection;
func (s *docSelection) PrevAll(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.PrevAll())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.PrevAllFiltered(selectorStr))
}

// PrevUntil gets all preceding siblings of each element up to but not including the element matched by the selector.
//
//	prevUntil(selector: string, until?: string): DocSelection;
func (s *docSelection) PrevUntil(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	selectorStr := s.getFirstStringArg(call)
	if len(call.Arguments) < 2 {
		return newDocSelectionGojaValue(s.doc, s.selection.PrevUntil(selectorStr))
	}
	untilStr := call.Argument(1).String()
	return newDocSelectionGojaValue(s.doc, s.selection.PrevFilteredUntil(selectorStr, untilStr))
}

// Siblings gets the siblings of each element (excluding the element) in the set of matched elements, optionally filtered by a selector.
//
//	siblings(selector?: string): DocSelection;
func (s *docSelection) Siblings(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Siblings())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.SiblingsFiltered(selectorStr))
}

// Children gets the element children of each element in the set of matched elements.
//
//	children(selector?: string): DocSelection;
func (s *docSelection) Children(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		return newDocSelectionGojaValue(s.doc, s.selection.Children())
	}

	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.ChildrenFiltered(selectorStr))
}

// Contents gets the children of each element in the Selection, including text and comment nodes. It returns a new Selection object containing
// these elements.
//
//	contents(): DocSelection;
func (s *docSelection) Contents(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	return newDocSelectionGojaValue(s.doc, s.selection.Contents())
}

// ContentsFiltered gets the children of each element in the Selection, filtered by the specified selector. It returns a new Selection object
// containing these elements. Since selectors only act on Element nodes, this function is an alias to ChildrenFiltered unless the selector is
// empty, in which case it is an alias to Contents.
//
//	contentsFiltered(selector: string): DocSelection;
func (s *docSelection) ContentsFiltered(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.ContentsFiltered(selectorStr))
}

// Filter reduces the set of matched elements to those that match the selector string. It returns a new Selection object for this subset of
// matching elements.
//
//	filter(selector: string | (index: number, element: DocSelection) => boolean): DocSelection;
func (s *docSelection) Filter(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		panic(s.doc.vm.ToValue("missing argument"))
	}

	switch call.Argument(0).Export().(type) {
	case string:
		selectorStr := s.getFirstStringArg(call)
		return newDocSelectionGojaValue(s.doc, s.selection.Filter(selectorStr))

	case func(call goja.FunctionCall) goja.Value:
		callback := call.Argument(0).Export().(func(call goja.FunctionCall) goja.Value)
		return newDocSelectionGojaValue(s.doc, s.selection.FilterFunction(func(i int, selection *goquery.Selection) bool {
			ret, ok := callback(goja.FunctionCall{Arguments: []goja.Value{
				s.doc.vm.ToValue(i),
				newDocSelectionGojaValue(s.doc, selection),
			}}).Export().(bool)
			if !ok {
				panic(s.doc.vm.NewTypeError("callback did not return a boolean").ToString())
			}
			return ret
		}))
	default:
		panic(s.doc.vm.NewTypeError("argument is not a string or function").ToString())
	}
}

// Not removes elements from the Selection that match the selector string. It returns a new Selection object with the matching elements removed.
//
//	not(selector: string | (index: number, element: DocSelection) => boolean): DocSelection;
func (s *docSelection) Not(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		panic(s.doc.vm.ToValue("missing argument"))
	}

	switch call.Argument(0).Export().(type) {
	case string:
		selectorStr := s.getFirstStringArg(call)
		return newDocSelectionGojaValue(s.doc, s.selection.Not(selectorStr))
	case func(call goja.FunctionCall) goja.Value:
		callback := call.Argument(0).Export().(func(call goja.FunctionCall) goja.Value)
		return newDocSelectionGojaValue(s.doc, s.selection.NotFunction(func(i int, selection *goquery.Selection) bool {
			ret, ok := callback(goja.FunctionCall{Arguments: []goja.Value{
				s.doc.vm.ToValue(i),
				newDocSelectionGojaValue(s.doc, selection),
			}}).Export().(bool)
			if !ok {
				panic(s.doc.vm.NewTypeError("callback did not return a boolean").ToString())
			}
			return ret
		}))
	default:
		panic(s.doc.vm.NewTypeError("argument is not a string or function").ToString())
	}
}

// Is checks the current matched set of elements against a selector and returns true if at least one of these elements matches.
//
//	is(selector: string | (index: number, element: DocSelection) => boolean): boolean;
func (s *docSelection) Is(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}

	if len(call.Arguments) == 0 || !gojaValueIsDefined(call.Argument(0)) {
		panic(s.doc.vm.ToValue("missing argument"))
	}

	switch call.Argument(0).Export().(type) {
	case string:
		selectorStr := s.getFirstStringArg(call)
		return s.doc.vm.ToValue(s.selection.Is(selectorStr))
	case func(call goja.FunctionCall) goja.Value:
		callback := call.Argument(0).Export().(func(call goja.FunctionCall) goja.Value)
		return s.doc.vm.ToValue(s.selection.IsFunction(func(i int, selection *goquery.Selection) bool {
			ret, ok := callback(goja.FunctionCall{Arguments: []goja.Value{
				s.doc.vm.ToValue(i),
				newDocSelectionGojaValue(s.doc, selection),
			}}).Export().(bool)
			if !ok {
				panic(s.doc.vm.NewTypeError("callback did not return a boolean").ToString())
			}
			return ret
		}))
	default:
		panic(s.doc.vm.NewTypeError("argument is not a string or function").ToString())
	}
}

// Has reduces the set of matched elements to those that have a descendant that matches the selector. It returns a new Selection object with the
// matching elements.
//
//	has(selector: string): DocSelection;
func (s *docSelection) Has(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	selectorStr := s.getFirstStringArg(call)
	return newDocSelectionGojaValue(s.doc, s.selection.Has(selectorStr))
}

// Each iterates over a Selection object, executing a function for each matched element. It returns the current Selection object. The function f
// is called for each element in the selection with the index of the element in that selection starting at 0, and a *Selection that contains only
// that element.
//
//	each(callback: (index: number, element: DocSelection) => void): DocSelection;
func (s *docSelection) Each(call goja.FunctionCall) (ret goja.Value) {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	callback, ok := call.Argument(0).Export().(func(call goja.FunctionCall) goja.Value)
	if !ok {
		panic(s.doc.vm.NewTypeError("argument is not a function").ToString())
	}
	s.selection.Each(func(i int, selection *goquery.Selection) {
		callback(goja.FunctionCall{Arguments: []goja.Value{
			s.doc.vm.ToValue(i),
			newDocSelectionGojaValue(s.doc, selection),
		}})
	})
	return goja.Undefined()
}

// Map passes each element in the current matched set through a function, producing a slice of string holding the returned values. The function f
// is called for each element in the selection with the index of the element in that selection starting at 0, and a *Selection that contains only
// that element.
//
//	map(callback: (index: number, element: DocSelection) => DocSelection): DocSelection[];
func (s *docSelection) Map(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	callback, ok := call.Argument(0).Export().(func(call goja.FunctionCall) goja.Value)
	if !ok {
		panic(s.doc.vm.NewTypeError("argument is not a function").ToString())
	}
	var retStr []interface{}
	var retDocSelection map[string]interface{}
	s.selection.Each(func(i int, selection *goquery.Selection) {
		val := callback(goja.FunctionCall{Arguments: []goja.Value{
			s.doc.vm.ToValue(i),
			newDocSelectionGojaValue(s.doc, selection),
		}})

		if valExport, ok := val.Export().(map[string]interface{}); ok {
			retDocSelection = valExport
		}
		retStr = append(retStr, val.Export())

	})
	if len(retStr) > 0 {
		return s.doc.vm.ToValue(retStr)
	}
	return s.doc.vm.ToValue(retDocSelection)
}

// First reduces the set of matched elements to the first in the set. It returns a new Selection object, and an empty Selection object if the
// selection is empty.
//
//	first(): DocSelection;
func (s *docSelection) First(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	return newDocSelectionGojaValue(s.doc, s.selection.First())
}

// Last reduces the set of matched elements to the last in the set. It returns a new Selection object, and an empty Selection object if the
// selection is empty.
//
//	last(): DocSelection;
func (s *docSelection) Last(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	return newDocSelectionGojaValue(s.doc, s.selection.Last())
}

// Eq reduces the set of matched elements to the one at the specified index. If a negative index is given, it counts backwards starting at the
// end of the set. It returns a new Selection object, and an empty Selection object if the index is invalid.
//
//	eq(index: number): DocSelection;
func (s *docSelection) Eq(call goja.FunctionCall) goja.Value {
	if s.selection == nil {
		panic(s.doc.vm.ToValue("selection is nil"))
	}
	index, ok := call.Argument(0).Export().(int64)
	if !ok {
		panic(s.doc.vm.NewTypeError("argument is not a number").String())
	}
	return newDocSelectionGojaValue(s.doc, s.selection.Eq(int(index)))
}
