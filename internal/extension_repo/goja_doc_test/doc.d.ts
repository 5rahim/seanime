declare class DocSelection {
    // Length returns the number of elements in the Selection object.
    length(): number

    // Attributes

    // Attr gets the specified attribute's value for the first element in the Selection. To get the value for each element individually, use a
    // looping construct such as Each or Map method.
    attr(name: string): string | undefined;

    // Get the descendants of each element in the current set of matched elements, filtered by a selector, or element.
    find(selector: string): DocSelection;

    // Parent gets the parent of each element in the Selection. It returns a new Selection object containing the matched elements.
    parent(selector?: string): DocSelection;

    // Parents gets the ancestors of each element in the current Selection. It returns a new Selection object with the matched elements.
    parents(selector?: string): DocSelection;

    // ParentsUntil gets the ancestors of each element in the Selection, up to but not including the element matched by the selector. It returns a
    // new Selection object containing the matched elements.
    parentsUntil(selector?: string, until?: string): DocSelection;

    // End ends the most recent filtering operation in the current chain and returns the set of matched elements to its previous state.
    end(): DocSelection;

    // For each element in the set, get the first element that matches the selector by testing the element itself and traversing up through its
    // ancestors in the DOM tree.
    closest(selector?: string): DocSelection;

    // Gets the next sibling of each selected element, optionally filtered by a selector.
    next(selector?: string): DocSelection;

    // Gets the previous sibling of each selected element optionally filtered by a selector.
    prev(selector?: string): DocSelection;

    // Get the ancestors of each element in the current set of matched elements, up to but not including the element matched by the selector, DOM
    // node, or cheerio object.
    prevUntil(selector?: string, until?: string): DocSelection;

    // Get the siblings of each element (excluding the element) in the set of matched elements, optionally filtered by a selector.
    siblings(selector?: string): DocSelection;

    // Gets the element children of each element in the set of matched elements.
    children(selector?: string): DocSelection;

    // Contents gets the children of each element in the Selection, including text and comment nodes. It returns a new Selection object containing
    // these elements.
    contents(): DocSelection;

    // ContentsFiltered gets the children of each element in the Selection, filtered by the specified selector. It returns a new Selection object
    // containing these elements. Since selectors only act on Element nodes, this function is an alias to ChildrenFiltered unless the selector is
    // empty, in which case it is an alias to Contents.
    contentsFiltered(selector: string): DocSelection;

    // Filter reduces the set of matched elements to those that match the selector string. It returns a new Selection object for this subset of
    // matching elements.
    filter(selector: string): DocSelection;

    // Not removes elements from the Selection that match the selector string. It returns a new Selection object with the matching elements removed.
    not(selector: string): DocSelection;

    // Is checks the current matched set of elements against a selector and returns true if at least one of these elements matches.
    is(selector: string): boolean;

    // Has reduces the set of matched elements to those that have a descendant that matches the selector. It returns a new Selection object with the
    // matching elements.
    has(selector: string): DocSelection;

    // Text gets the combined text contents of each element in the set of matched elements, including their descendants.
    text(): string;

    html(): string | null;

    // Each iterates over a Selection object, executing a function for each matched element. It returns the current Selection object. The function f
    // is called for each element in the selection with the index of the element in that selection starting at 0, and a *Selection that contains only
    // that element.
    each(callback: (index: number, element: DocSelection) => void): DocSelection;

    // Map passes each element in the current matched set through a function, producing a slice of string holding the returned values. The function f
    // is called for each element in the selection with the index of the element in that selection starting at 0, and a *Selection that contains only
    // that element.
    map(callback: (index: number, element: DocSelection) => DocSelection): DocSelection[];

    // First reduces the set of matched elements to the first in the set. It returns a new Selection object, and an empty Selection object if the
    // selection is empty.
    first(): DocSelection;

    // Last reduces the set of matched elements to the last in the set. It returns a new Selection object, and an empty Selection object if the
    // selection is empty.
    last(): DocSelection;

    // Eq reduces the set of matched elements to the one at the specified index. If a negative index is given, it counts backwards starting at the
    // end of the set. It returns a new Selection object, and an empty Selection object if the index is invalid.
    eq(index: number): DocSelection;
}

declare class Doc extends DocSelection {
    constructor(html: string);
}
