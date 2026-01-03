package plugin_ui

import "testing"

func TestContainsDangerousHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Safe HTML", "<div>Hello World</div>", false},
		{"Safe HTML with Attributes", "<div class='container' id='main'></div>", false},
		{"Safe Text", "This is a script about java", false},

		{"Script Tag", "<script>alert(1)</script>", true},
		{"Script Tag Upper", "<SCRIPT>alert(1)</SCRIPT>", true},
		{"Script Tag Mixed", "<ScRiPt>alert(1)</sCrIpT>", true},

		{"Javascript Protocol", "<a href='javascript:void(0)'>", true},
		{"Javascript Protocol Mixed", "<a href='JaVaScRiPt:void(0)'>", true},

		{"Event Handler Simple", "<div onclick='alert(1)'>", true},
		{"Event Handler Spaces", "<div onclick = 'alert(1)'>", true},
		{"Event Handler Mixed", "<div oNcLiCk='alert(1)'>", true},
		{"Event Handler Suffix", "<div onmouseover='alert(1)'>", true},

		{"False Positive Check", "This contains onclick text but no equals", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsDangerousHTML(tt.input); got != tt.expected {
				t.Errorf("containsDangerousHTML(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsDangerousAttribute(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Safe Attribute", "class", false},
		{"Safe Attribute ID", "id", false},
		{"Safe Attribute Href", "href", false},

		{"Dangerous OnClick", "onclick", true},
		{"Dangerous OnClick Mixed", "oNcLiCk", true},
		{"Dangerous OnLoad", "onload", true},
		{"Dangerous OnError", "onerror", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDangerousAttribute(tt.input); got != tt.expected {
				t.Errorf("isDangerousAttribute(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsDangerousAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		attrName string
		value    string
		expected bool
	}{
		{"Safe Href", "href", "https://google.com", false},
		{"Safe Src", "src", "/images/logo.png", false},
		{"Safe Data Image", "src", "data:image/png;base64,xyz", false},

		{"Javascript Href", "href", "javascript:alert(1)", true},
		{"Javascript Href Mixed", "href", "JaVaScRiPt:alert(1)", true},
		{"Javascript Href Spaces", "href", "  javascript:alert(1)", true},

		{"Data HTML Href", "href", "data:text/html,<b>hi</b>", true},
		{"Data HTML Mixed", "href", "DaTa:TeXt/HtMl,<b>hi</b>", true},

		{"Wrong Attribute Name", "class", "javascript:alert(1)", false}, // class allows weird values
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDangerousAttributeValue(tt.attrName, tt.value); got != tt.expected {
				t.Errorf("isDangerousAttributeValue(%q, %q) = %v, want %v", tt.name, tt.value, got, tt.expected)
			}
		})
	}
}

func TestIsDangerousProperty(t *testing.T) {
	tests := []struct {
		name     string
		propName string
		expected bool
	}{
		{"innerHTML", "innerHTML", true},
		{"outerHTML", "outerHTML", true},
		{"onclick", "onclick", true},
		{"textContent", "textContent", false},
		{"innerText", "innerText", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDangerousProperty(tt.propName); got != tt.expected {
				t.Errorf("isDangerousProperty(%q) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestContainsDangerousCSS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Safe CSS", "color: red; background: blue;", false},
		{"Dangerous JS", "background: url('javascript:alert(1)')", true},
		{"Dangerous Expression", "width: expression(alert(1))", true},
		{"Dangerous Binding", "-moz-binding: url('xml')", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsDangerousCSS(tt.input); got != tt.expected {
				t.Errorf("containsDangerousCSS(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

// expect 0 allocs/op here
func BenchmarkIsDangerousAttribute(b *testing.B) {
	inputs := []string{"class", "onclick", "href", "ONCLICK", "data-test", "OnMouseOver"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = isDangerousAttribute(inputs[i%len(inputs)])
	}
}

func BenchmarkContainsDangerousHTML_Safe(b *testing.B) {
	input := `
		<div class="container">
			<h1>Hello World</h1>
			<p>This is a test paragraph with some random text.</p>
			<ul>
				<li>Item 1</li>
				<li>Item 2</li>
			</ul>
		</div>
	`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = containsDangerousHTML(input)
	}
}

func BenchmarkContainsDangerousHTML_Dangerous(b *testing.B) {
	input := `<div class="container">
			<h1>Hello World</h1>
			<p>This is a test paragraph with some random text.</p>
			<ul>
				<li>Item 1</li>
				<li>Item 2</li>
			</ul>
			<button onclick="alert(1)">Click Me</button>
		</div>`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = containsDangerousHTML(input)
	}

}

func BenchmarkIsDangerousAttributeValue(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isDangerousAttributeValue("href", "javascript:void(0)")
		_ = isDangerousAttributeValue("src", "https://google.com/image.png")
		_ = isDangerousAttributeValue("class", "btn btn-primary")
	}
}
