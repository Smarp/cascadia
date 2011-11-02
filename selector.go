package cascadia

import (
	"fmt"
	"html"
	"strings"
)

// the Selector type, and functions for creating them

// A Selector is a function which tells whether a node matches or not.
type Selector func(*html.Node) bool

// Compile parses a selector and returns, if successful, a Selector object
// that can be used to match against html.Node objects.
func Compile(sel string) (Selector, error) {
	p := &parser{s: sel}
	compiled, err := p.parseSimpleSelectorSequence() // TODO: more complicated selectors
	if err != nil {
		return nil, err
	}

	if p.i < len(sel) {
		return nil, fmt.Errorf("parsing %q: %d bytes left over", sel, len(sel)-p.i)
	}

	return compiled, nil
}

// MatchAll returns a slice of the nodes that match the selector,
// from n and its children.
func (s Selector) MatchAll(n *html.Node) (result []*html.Node) {
	if s(n) {
		result = append(result, n)
	}

	for _, child := range n.Child {
		result = append(result, s.MatchAll(child)...)
	}

	return
}

// typeSelector returns a Selector that matches nodes with a given tag name.
func typeSelector(tag string) Selector {
	tag = toLowerASCII(tag)
	return func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag
	}
}

// toLowerASCII returns s with all ASCII capital letters lowercased.
func toLowerASCII(s string) string {
	var b []byte
	for i := 0; i < len(s); i++ {
		if c := s[i]; 'A' <= c && c <= 'Z' {
			if b == nil {
				b = make([]byte, len(s))
				copy(b, s)
			}
			b[i] = s[i] + ('a' - 'A')
		}
	}

	if b == nil {
		return s
	}

	return string(b)
}

// attributeExistsSelector returns a Selector that matches nodes that have
// an attribute named key.
func attributeExistsSelector(key string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return true
			}
		}
		return false
	}
}

// attributeEqualsSelector returns a Selector that matches nodes where
// the attribute named key has the value val.
func attributeEqualsSelector(key, val string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return a.Val == val
			}
		}
		return false
	}
}

// attributeIncludesSelector returns a Selector that matches nodes where 
// the attribute named key is a whitespace-separated list that includes val.
func attributeIncludesSelector(key, val string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				s := a.Val
				for s != "" {
					i := strings.IndexAny(s, " \t\r\n\f")
					if i == -1 {
						return s == val
					}
					if s[:i] == val {
						return true
					}
					s = s[i+1:]
				}
			}
		}
		return false
	}
}

// attributeDashmatchSelector returns a Selector that matches nodes where
// the attribute named key equals val or starts with val plus a hyphen.
func attributeDashmatchSelector(key, val string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				if a.Val == val {
					return true
				}
				if len(a.Val) <= len(val) {
					return false
				}
				if a.Val[:len(val)] == val && a.Val[len(val)] == '-' {
					return true
				}
				return false
			}
		}
		return false
	}
}

// attributePrefixSelector returns a Selector that matches nodes where
// the attribute named key starts with val.
func attributePrefixSelector(key, val string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return strings.HasPrefix(a.Val, val)
			}
		}
		return false
	}
}

// attributeSuffixSelector returns a Selector that matches nodes where
// the attribute named key ends with val.
func attributeSuffixSelector(key, val string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return strings.HasSuffix(a.Val, val)
			}
		}
		return false
	}
}

// attributeSubstringSelector returns a Selector that matches nodes where
// the attribute named key contains val.
func attributeSubstringSelector(key, val string) Selector {
	key = toLowerASCII(key)
	return func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == key {
				return strings.Contains(a.Val, val)
			}
		}
		return false
	}
}

// intersectionSelector returns a selector that matches nodes that match
// both a and b.
func intersectionSelector(a, b Selector) Selector {
	return func(n *html.Node) bool {
		return a(n) && b(n)
	}
}

// negatedSelector returns a selector that matches nodes that do not match a.
func negatedSelector(a Selector) Selector {
	return func(n *html.Node) bool {
		return !a(n)
	}
}

// nthChildSelector returns a selector that implements :nth-child(an+b).
// If last is true, implements :nth-last-child instead.
func nthChildSelector(a, b int, last bool) Selector {
	return func(n *html.Node) bool {
		parent := n.Parent
		if parent == nil {
			return false
		}

		var i int
		c := parent.Child
		for i = 0; i < len(c); i++ {
			if c[i] == n {
				break
			}
		}

		if i == len(c) {
			return false
		}

		if last {
			i = len(c) - i
		} else {
			i++
		}

		i -= b
		if a == 0 {
			return i == 0
		}

		return i%a == 0 && i/a >= 0
	}
}
