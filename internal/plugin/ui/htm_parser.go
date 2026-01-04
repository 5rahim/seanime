package plugin_ui

import (
	"fmt"
	"strings"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// JSX-like syntax for rendering
// Example:
//	<stack gap="2">
//      <switch fieldRef="notifications" label="Enable Notifications" />
//      <checkbox fieldRef="auto-update" label="Auto Update" />
//      <select fieldRef="theme" placeholder="Select Theme" options=${[{ value: "test", label: "Test" }]} />
//  </stack>
//
//  <button
//      label="${showAdvanced.get() ? "Hide Advanced" : "Show Advanced"}"
//      onClick="${ctx.eventHandler("toggle-advanced", () => {
//            showAdvanced.set(!showAdvanced.get())
//        })}"
//      variant="outline"
//  />

// INCOMPLETE
// Handle passing objects are prop
// loops using tray.inlineHtm e.g., <stack>${items.map(i => tray.inlineHtm`<span>${i}</span>`)}</stack>

// parseHTM converts an HTM string into component structures.
// It first evaluates template literals (${...}) and then parses the resulting markup.
func (c *ComponentManager) parseHTM(htm string) (interface{}, error) {
	htm = strings.TrimSpace(htm)
	if htm == "" {
		return nil, nil
	}

	// 1. Evaluate ${} expressions handling nested braces
	processedHtm, err := c.preprocessTemplateLiterals(htm)
	if err != nil {
		return nil, err
	}

	// 2. Convert string to component tree
	parser := newHtmParser(processedHtm, c.ctx.vm, c.ctx)
	return parser.parse()
}

// preprocessTemplateLiterals evaluates ${} expressions in the HTM string.
// It uses a depth counter to correctly handle nested braces within expressions.
func (c *ComponentManager) preprocessTemplateLiterals(input string) (string, error) {
	var sb strings.Builder
	length := len(input)
	cursor := 0

	for i := 0; i < length; i++ {
		// Detect start of literal "${"
		if i+1 < length && input[i] == '$' && input[i+1] == '{' {
			// Write everything before the literal
			sb.WriteString(input[cursor:i])

			startExpr := i + 2
			depth := 1
			endExpr := -1

			// Scan forward to find matching closing brace
			for j := startExpr; j < length; j++ {
				switch input[j] {
				case '{':
					depth++
				case '}':
					depth--
				}

				if depth == 0 {
					endExpr = j
					break
				}
			}

			if endExpr == -1 {
				return "", fmt.Errorf("unclosed template literal starting at position %d", i)
			}

			// Extract and evaluate the JS expression
			expr := input[startExpr:endExpr]
			val, err := c.ctx.vm.RunString(expr)
			if err != nil {
				// Log error but continue
				// we could fail completely
				c.ctx.logger.Warn().Err(err).Str("expr", expr).Msg("plugin: Failed to evaluate template expression")
			} else {
				exported := val.Export()
				if exported != nil {
					sb.WriteString(fmt.Sprintf("%v", exported))
				}
			}

			// Advance cursor
			i = endExpr
			cursor = i + 1
		}
	}

	// Append remaining content
	if cursor < length {
		sb.WriteString(input[cursor:])
	}

	return sb.String(), nil
}

// htmParser implements a recursive descent parser.
// It preserves case sensitivity for attributes.
type htmParser struct {
	input string
	len   int
	pos   int
	vm    *goja.Runtime
	ctx   *Context
}

func newHtmParser(input string, vm *goja.Runtime, ctx *Context) *htmParser {
	return &htmParser{
		input: input,
		len:   len(input),
		pos:   0,
		vm:    vm,
		ctx:   ctx,
	}
}

func (p *htmParser) peek() byte {
	if p.pos >= p.len {
		return 0
	}
	return p.input[p.pos]
}

func (p *htmParser) eof() bool {
	return p.pos >= p.len
}

func (p *htmParser) consume() byte {
	b := p.peek()
	if !p.eof() {
		p.pos++
	}
	return b
}

func (p *htmParser) match(s string) bool {
	if p.pos+len(s) > p.len {
		return false
	}
	return p.input[p.pos:p.pos+len(s)] == s
}

func (p *htmParser) skipWhitespace() {
	for !p.eof() && isWhitespace(p.peek()) {
		p.pos++
	}
}

// readIdentifier reads [a-zA-Z0-9-:]+
func (p *htmParser) readIdentifier() string {
	start := p.pos
	for !p.eof() {
		b := p.peek()
		if isAlphaNumeric(b) || b == '-' || b == ':' || b == '_' || b == '.' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *htmParser) parse() (interface{}, error) {
	// Root level can be a list of nodes
	var nodes []interface{}
	for !p.eof() {
		p.skipWhitespace()
		if p.eof() {
			break
		}
		node, err := p.parseNode()
		if err != nil {
			return nil, err
		}
		if node != nil {
			nodes = append(nodes, node)
		}
	}

	// If single root, return it directly to match component_utils expectation
	if len(nodes) == 1 {
		return nodes[0], nil
	}
	return nodes, nil
}

func (p *htmParser) parseNode() (interface{}, error) {
	// Check for element
	if p.peek() == '<' {
		// Check for closing tag (should not happen at parseNode entry unless mismatched)
		if p.match("</") {
			return nil, fmt.Errorf("unexpected closing tag at position %d", p.pos)
		}
		return p.parseElement()
	}

	// text name
	return p.parseText()
}

func (p *htmParser) parseText() (interface{}, error) {
	start := p.pos
	for !p.eof() && p.peek() != '<' {
		p.pos++
	}
	text := p.input[start:p.pos]

	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	// For text nodes, we also return a Component struct
	return p.createComponent("span", map[string]interface{}{"text": text}, nil)
}

func (p *htmParser) parseElement() (interface{}, error) {
	p.consume() // eat '<'

	tagName := p.readIdentifier()
	if tagName == "" {
		return nil, fmt.Errorf("expected tag name at position %d", p.pos)
	}

	attrs := p.parseAttributes()

	p.skipWhitespace()
	selfClosing := false
	if p.peek() == '/' {
		selfClosing = true
		p.consume() // eat '/'
	}

	if p.consume() != '>' {
		return nil, fmt.Errorf("expected '>' after tag %s", tagName)
	}

	var children []interface{}

	if !selfClosing {
		for !p.eof() {
			p.skipWhitespace()

			// Check for closing tag
			if p.match("</") {
				p.pos += 2 // eat '</'
				closeTagName := p.readIdentifier()
				if closeTagName != tagName {
					return nil, fmt.Errorf("mismatched closing tag: expected </%s>, got </%s>", tagName, closeTagName)
				}
				p.skipWhitespace()
				if p.consume() != '>' {
					return nil, fmt.Errorf("expected '>' after closing tag %s", closeTagName)
				}
				break
			}

			// Parse Child
			child, err := p.parseNode()
			if err != nil {
				return nil, err
			}
			if child != nil {
				children = append(children, child)
			}
		}
	}

	return p.createComponent(tagName, attrs, children)
}

func (p *htmParser) parseAttributes() map[string]interface{} {
	attrs := make(map[string]interface{})

	for {
		p.skipWhitespace()
		if p.eof() {
			break
		}
		b := p.peek()
		if b == '>' || b == '/' {
			break
		}

		key := p.readIdentifier()
		if key == "" {
			p.consume()
			continue
		}

		p.skipWhitespace()

		// Boolean attribute?
		if p.peek() != '=' {
			attrs[key] = true
			continue
		}

		p.consume() // eat '='
		p.skipWhitespace()
		val := p.parseAttributeValue()
		attrs[key] = val
	}
	return attrs
}

func (p *htmParser) parseAttributeValue() interface{} {
	if p.eof() {
		return ""
	}

	quote := p.peek()
	if quote == '"' || quote == '\'' {
		p.consume() // eat opening quote
		start := p.pos
		for !p.eof() && p.peek() != quote {
			p.pos++
		}
		val := p.input[start:p.pos]
		if !p.eof() {
			p.consume() // eat closing quote
		}
		return val
	}

	// Unquoted value
	start := p.pos
	for !p.eof() {
		b := p.peek()
		if isWhitespace(b) || b == '>' || b == '/' {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *htmParser) createComponent(tagName string, attrs map[string]interface{}, children []interface{}) (interface{}, error) {
	componentType := p.mapTagToComponentType(tagName)

	component := Component{
		ID:    uuid.New().String(),
		Type:  componentType,
		Props: make(map[string]interface{}),
	}

	// Apply props using the mapper
	p.mapAttributesToProps(componentType, attrs, &component, children)

	return component, nil
}

// mapTagToComponentType maps HTM tags to internal component types
func (p *htmParser) mapTagToComponentType(tagName string) string {
	return strings.ToLower(tagName)
}

// mapAttributesToProps maps HTML attributes to Component props structure
func (p *htmParser) mapAttributesToProps(componentType string, attrs map[string]interface{}, component *Component, children []interface{}) {
	// common props (might not apply to all)
	if v, ok := attrs["className"]; ok {
		component.Props["className"] = v
	}
	if v, ok := attrs["class"]; ok {
		component.Props["className"] = v
	}
	if v, ok := attrs["style"]; ok {
		component.Props["style"] = v
	}
	if v, ok := attrs["key"]; ok {
		component.Key = fmt.Sprint(v)
	}

	// component-specific props
	switch componentType {
	case "text":
		if v, ok := attrs["text"]; ok {
			component.Props["text"] = v
		} else if len(children) > 0 {
			// Extract text from first child component if it's a text node
			if childComp, ok := children[0].(Component); ok {
				if text, ok := childComp.Props["text"]; ok {
					component.Props["text"] = text
				}
			}
		}

	case "button":
		copyProps(attrs, component.Props, getComponentPropNames(buttonComponentProps)...)

	case "input":
		copyProps(attrs, component.Props, getComponentPropNames(inputComponentProps)...)

	case "switch", "checkbox":
		copyProps(attrs, component.Props, getComponentPropNames(switchComponentProps)...)

	case "select":
		copyProps(attrs, component.Props, getComponentPropNames(selectComponentProps)...)

	case "a":
		copyProps(attrs, component.Props, getComponentPropNames(aComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "p":
		copyProps(attrs, component.Props, getComponentPropNames(pComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "span":
		copyProps(attrs, component.Props, getComponentPropNames(spanComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "div":
		copyProps(attrs, component.Props, getComponentPropNames(divComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "flex":
		copyProps(attrs, component.Props, getComponentPropNames(flexComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "stack":
		copyProps(attrs, component.Props, getComponentPropNames(stackComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "alert":
		copyProps(attrs, component.Props, getComponentPropNames(alertComponentProps)...)

	case "badge":
		copyProps(attrs, component.Props, getComponentPropNames(badgeComponentProps)...)

	case "tooltip":
		copyProps(attrs, component.Props, getComponentPropNames(tooltipComponentProps)...)
		if len(children) > 0 {
			component.Props["item"] = children[0]
		}

	case "modal":
		copyProps(attrs, component.Props, getComponentPropNames(modalComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "tabs":
		copyProps(attrs, component.Props, getComponentPropNames(tabsComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "tabs-list":
		copyProps(attrs, component.Props, getComponentPropNames(tabsListComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "tabs-trigger":
		copyProps(attrs, component.Props, getComponentPropNames(tabsTriggerComponentProps)...)
		// Pass the child
		if childComp, ok := children[0].(Component); ok {
			component.Props["item"] = childComp
		}

	case "tabs-content":
		copyProps(attrs, component.Props, getComponentPropNames(tabsContentComponentProps)...)
		if children != nil {
			component.Props["items"] = children
		}

	case "css":
		copyProps(attrs, component.Props, getComponentPropNames(cssComponentProps)...)
		if len(children) > 0 {
			if childComp, ok := children[0].(Component); ok {
				if text, ok := childComp.Props["text"]; ok {
					component.Props["css"] = text
				}
			}
		}

	default:
		// Pass through all attributes for unknown/generic components
		for k, v := range attrs {
			if k != "className" && k != "class" && k != "style" && k != "key" {
				component.Props[k] = v
			}
		}
		if children != nil {
			component.Props["items"] = children
		}
	}
}

func copyProps(source map[string]interface{}, dest map[string]interface{}, keys ...string) {
	for _, k := range keys {
		if v, ok := source[k]; ok {
			dest[k] = v
		}
	}
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func isAlphaNumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func cleanHTMLString(s string) string {
	// Remove backticks if present
	s = strings.Trim(s, "`")
	return strings.TrimSpace(s)
}
