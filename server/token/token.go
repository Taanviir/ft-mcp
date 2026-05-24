package token

import (
	_ "embed"
	"net/http"
	"strings"
)

//go:embed token.html
var page string

// Serve renders the token UI form. Pass errMsg to show an error, tok to show
// a successfully issued token, or empty strings for the blank form.
func Serve(w http.ResponseWriter, errMsg, tok string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var errDiv, resultDiv string
	if errMsg != "" {
		errDiv = `<div class="error">` + errMsg + `</div>`
	}
	if tok != "" {
		resultDiv = `<div class="result">` +
			`<div class="result-title">Your bearer token</div>` +
			`<div class="token-row">` +
			`<input type="text" id="tok" value="` + tok + `" readonly>` +
			`<button id="cpbtn" class="copy-btn" onclick="copyToken()">Copy</button>` +
			`</div>` +
			`<p class="usage">Add to your MCP client config:<br>` +
			`<code>Authorization: Bearer &lt;token&gt;</code><br>` +
			`Token expires after 24 hours.</p>` +
			`</div>`
	}

	strings.NewReplacer("{{ERROR}}", errDiv, "{{RESULT}}", resultDiv).WriteString(w, page)
}
