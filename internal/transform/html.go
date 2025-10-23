package transform

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// ApplyTransformations applies a list of operations to an HTML document
func ApplyTransformations(doc *html.Node, operations []Operation) error {
	for _, op := range operations {
		if err := applyOperation(doc, op); err != nil {
			// Log error but continue with other operations
			fmt.Printf("Warning: failed to apply operation %v: %v\n", op, err)
		}
	}
	return nil
}

// applyOperation applies a single operation to the HTML document
func applyOperation(doc *html.Node, op Operation) error {
	// Find the target element(s)
	nodes := findNodesBySelector(doc, op.Selector)
	if len(nodes) == 0 {
		return fmt.Errorf("no elements found for selector: %s", op.Selector)
	}

	for _, node := range nodes {
		switch op.Type {
		case OpSetText:
			setText(node, op.Value)
		case OpSetStyle:
			setStyle(node, op.Property, op.Value)
		case OpSetAttr:
			setAttr(node, op.Property, op.Value)
		case OpSetHTML:
			setHTML(node, op.Value)
		case OpRemove:
			removeNode(node)
		case OpHide:
			setStyle(node, "display", "none")
		case OpShow:
			setStyle(node, "display", "")
		default:
			return fmt.Errorf("unknown operation type: %s", op.Type)
		}
	}

	return nil
}

// setText replaces the text content of a node
func setText(node *html.Node, text string) {
	// Remove all child nodes
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		node.RemoveChild(child)
		child = next
	}

	// Add new text node
	node.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}

// setStyle sets or updates a CSS property in the style attribute
func setStyle(node *html.Node, property, value string) {
	if node.Type != html.ElementNode {
		return
	}

	// Get existing style attribute
	var styleAttr *html.Attribute
	for i, attr := range node.Attr {
		if attr.Key == "style" {
			styleAttr = &node.Attr[i]
			break
		}
	}

	// Parse existing styles
	styles := make(map[string]string)
	if styleAttr != nil {
		for _, part := range strings.Split(styleAttr.Val, ";") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				styles[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// Update style
	if value == "" {
		delete(styles, property)
	} else {
		styles[property] = value
	}

	// Rebuild style string
	var styleStr strings.Builder
	for k, v := range styles {
		if styleStr.Len() > 0 {
			styleStr.WriteString("; ")
		}
		styleStr.WriteString(k)
		styleStr.WriteString(": ")
		styleStr.WriteString(v)
	}

	// Update or create style attribute
	if styleAttr != nil {
		styleAttr.Val = styleStr.String()
	} else {
		node.Attr = append(node.Attr, html.Attribute{
			Key: "style",
			Val: styleStr.String(),
		})
	}
}

// setAttr sets an HTML attribute
func setAttr(node *html.Node, key, value string) {
	if node.Type != html.ElementNode {
		return
	}

	// Update existing or add new
	found := false
	for i, attr := range node.Attr {
		if attr.Key == key {
			node.Attr[i].Val = value
			found = true
			break
		}
	}

	if !found {
		node.Attr = append(node.Attr, html.Attribute{
			Key: key,
			Val: value,
		})
	}
}

// setHTML replaces the inner HTML of a node
func setHTML(node *html.Node, htmlContent string) {
	// Remove all children
	for child := node.FirstChild; child != nil; {
		next := child.NextSibling
		node.RemoveChild(child)
		child = next
	}

	// Parse new HTML content
	nodes, err := html.ParseFragment(strings.NewReader(htmlContent), &html.Node{
		Type:     html.ElementNode,
		Data:     "div",
		DataAtom: 0,
	})
	if err != nil {
		return
	}

	// Append parsed nodes
	for _, child := range nodes {
		node.AppendChild(child)
	}
}

// removeNode removes a node from the tree
func removeNode(node *html.Node) {
	if node.Parent != nil {
		node.Parent.RemoveChild(node)
	}
}

// findNodesBySelector finds nodes matching a simple CSS selector
// Supports: .class, #id, element, [attr], [attr=value]
func findNodesBySelector(doc *html.Node, selector string) []*html.Node {
	var results []*html.Node

	selector = strings.TrimSpace(selector)

	// Determine selector type
	var matchFunc func(*html.Node) bool

	if strings.HasPrefix(selector, ".") {
		// Class selector
		className := selector[1:]
		matchFunc = func(n *html.Node) bool {
			return hasClass(n, className)
		}
	} else if strings.HasPrefix(selector, "#") {
		// ID selector
		id := selector[1:]
		matchFunc = func(n *html.Node) bool {
			return getAttr(n, "id") == id
		}
	} else if strings.HasPrefix(selector, "[") && strings.HasSuffix(selector, "]") {
		// Attribute selector
		attrStr := selector[1 : len(selector)-1]
		parts := strings.SplitN(attrStr, "=", 2)
		attrKey := strings.TrimSpace(parts[0])

		if len(parts) == 1 {
			// [attr] - has attribute
			matchFunc = func(n *html.Node) bool {
				return hasAttr(n, attrKey)
			}
		} else {
			// [attr=value]
			attrValue := strings.Trim(strings.TrimSpace(parts[1]), "\"'")
			matchFunc = func(n *html.Node) bool {
				return getAttr(n, attrKey) == attrValue
			}
		}
	} else {
		// Element selector
		matchFunc = func(n *html.Node) bool {
			return n.Type == html.ElementNode && n.Data == selector
		}
	}

	// Walk the tree
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if matchFunc(n) {
			results = append(results, n)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return results
}

// Helper functions
func hasClass(node *html.Node, className string) bool {
	if node.Type != html.ElementNode {
		return false
	}
	classAttr := getAttr(node, "class")
	classes := strings.Fields(classAttr)
	for _, c := range classes {
		if c == className {
			return true
		}
	}
	return false
}

func getAttr(node *html.Node, key string) string {
	if node.Type != html.ElementNode {
		return ""
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func hasAttr(node *html.Node, key string) bool {
	if node.Type != html.ElementNode {
		return false
	}
	for _, attr := range node.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// RenderHTML renders an HTML node tree to a string
func RenderHTML(doc *html.Node) (string, error) {
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}
