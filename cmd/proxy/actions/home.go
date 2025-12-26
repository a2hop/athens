package actions

import (
	"errors"
	"html/template"
	"net/http"
	"os"
	"strings"

	"github.com/gomods/athens/pkg/config"
	"github.com/gomods/athens/pkg/log"
)

const homepage = `<!DOCTYPE html>
<html>
<head>
	<meta charset="utf-8"></meta>
	<title>Athens - Go Module Proxy</title>
	<style>
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
			margin: 0;
			padding: 20px;
			background: #f5f5f5;
			color: #333;
		}

		.container {
			max-width: 900px;
			margin: 0 auto;
			background: white;
			padding: 40px;
			border-radius: 8px;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}

		h1 {
			color: #00ADD8;
			border-bottom: 3px solid #00ADD8;
			padding-bottom: 10px;
		}

		h2 {
			color: #00ADD8;
			margin-top: 30px;
		}

		h3 {
			color: #5DC9E2;
			margin-top: 20px;
		}

		pre {
			background-color: #2d2d2d;
			color: #f8f8f2;
			padding: 15px;
			border-radius: 5px;
			overflow-x: auto;
			border-left: 4px solid #00ADD8;
		}

		code {
			background-color: #f4f4f4;
			padding: 2px 6px;
			border-radius: 3px;
			font-family: 'Courier New', monospace;
		}

		.setup-box {
			background: #e7f5ff;
			border: 2px solid #00ADD8;
			border-radius: 8px;
			padding: 20px;
			margin: 20px 0;
		}

		.download-btn {
			display: inline-block;
			background: #00ADD8;
			color: white;
			padding: 12px 24px;
			text-decoration: none;
			border-radius: 5px;
			font-weight: bold;
			margin: 10px 10px 10px 0;
			transition: background 0.3s;
		}

		.download-btn:hover {
			background: #0099C1;
		}

		.warning {
			background: #fff3cd;
			border-left: 4px solid #ffc107;
			padding: 15px;
			margin: 20px 0;
			border-radius: 4px;
		}

		.success {
			background: #d4edda;
			border-left: 4px solid #28a745;
			padding: 15px;
			margin: 20px 0;
			border-radius: 4px;
		}

		a {
			color: #00ADD8;
			text-decoration: none;
		}

		a:hover {
			text-decoration: underline;
		}

		ul {
			line-height: 1.8;
		}

	</style>
</head>
<body>
	<div class="container">
	
	<h1>🏛️ Welcome to Athens Go Module Proxy</h1>

	<div class="success">
		<strong>✓ Athens is running!</strong> Your Go module proxy is ready to serve modules.
	</div>

	<h2>🚀 Quick Setup</h2>
	<div class="setup-box">
		<h3>One-Line Setup (Recommended)</h3>
		<p>Copy and paste this command to configure your Go environment:</p>
		<pre>eval "$(curl -fsSL {{ .Host }}/setup.sh)"</pre>
		
		<h3>Manual Setup</h3>
		<p>Or set these environment variables in your terminal:</p>
		<pre>export GOPROXY="{{ .Host }}"
export GONOSUMDB="*"
export GONOPROXY=""</pre>
		
		<p>Make it permanent by adding to your <code>~/.bashrc</code> or <code>~/.zshrc</code>:</p>
		<pre>echo 'export GOPROXY="{{ .Host }}"' >> ~/.bashrc
echo 'export GONOSUMDB="*"' >> ~/.bashrc
echo 'export GONOPROXY=""' >> ~/.bashrc
source ~/.bashrc</pre>
	</div>

	<div class="warning">
		<strong>⚙️ Configuration Explained:</strong>
		<ul>
			<li><code>GOPROXY</code> - Routes all module requests through Athens</li>
			<li><code>GONOSUMDB="*"</code> - Skips checksum verification for all modules (needed for private repos)</li>
			<li><code>GONOPROXY=""</code> - Ensures everything goes through the proxy (prevents direct fetching)</li>
		</ul>
	</div>

	<h2>✅ Verify Setup</h2>
	<pre>go env GOPROXY GONOSUMDB GONOPROXY</pre>

	<h2>📦 Test It Out</h2>
	<pre>go get github.com/your-org/your-private-repo</pre>

	<h2>📚 Athens API Reference</h2>
	<p>Use the <a href="/catalog">catalog</a> endpoint to get a list of all modules in the proxy.</p>

	<h3>List Versions</h3>
	<p>Get all versions for a module:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/list</pre>

	<h3>Version Info</h3>
	<p>Get metadata about a specific version:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/v1.0.0.info</pre>

	<h3>go.mod File</h3>
	<p>Download the go.mod file:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/v1.0.0.mod</pre>

	<h3>Module Source</h3>
	<p>Download module source as zip:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@v/v1.0.0.zip</pre>

	<h3>Latest Version</h3>
	<p>Get the latest version info:</p>
	<pre>GET {{ .Host }}/github.com/acidburn/htp/@latest</pre>

	<hr style="margin: 40px 0; border: none; border-top: 1px solid #ddd;">
	<p style="text-align: center; color: #666;">
		<a href="https://docs.gomods.io" target="_blank">Documentation</a> • 
		<a href="https://github.com/gomods/athens" target="_blank">GitHub</a> • 
		Version {{ .Version }}
	</p>

	</div>
</body>
</html>
`

func proxyHomeHandler(c *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		lggr := log.EntryFromContext(r.Context())

		templateData := make(map[string]string)

		templateContents := homepage

		// load the template from the file system if it exists, otherwise revert to default
		rawTemplateFileContents, err := os.ReadFile(c.HomeTemplatePath)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				// this is some other error, log it and revert to default
				lggr.SystemErr(err)
			}
		} else {
			templateContents = string(rawTemplateFileContents)
		}

		// This should be correct in most cases. If it is not, users can supply their own template
		templateData["Host"] = r.Host

		// use host from URL, if it exists
		if r.URL.Host != "" {
			templateData["Host"] = r.URL.Host
		}

		// if the host does not have a scheme, add one based on the request
		if !strings.HasPrefix(templateData["Host"], "http://") && !strings.HasPrefix(templateData["Host"], "https://") {
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				templateData["Host"] = "https://" + templateData["Host"]
			} else {
				templateData["Host"] = "http://" + templateData["Host"]
			}
		}

		templateData["NoSumPatterns"] = strings.Join(c.NoSumPatterns, ",")
		templateData["Version"] = "Custom Build"

		tmp, err := template.New("home").Parse(templateContents)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)

		err = tmp.ExecuteTemplate(w, "home", templateData)
		if err != nil {
			lggr.SystemErr(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// setupScriptHandler serves the setup.sh script for clients to download
func setupScriptHandler(w http.ResponseWriter, r *http.Request) {
	// Get the host from the request
	host := r.Host
	if r.URL.Host != "" {
		host = r.URL.Host
	}

	// Add scheme
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	proxyURL := scheme + "://" + host

	script := `#!/bin/bash
# Athens Go Module Proxy Client Setup
# Usage: eval "$(curl -fsSL ` + proxyURL + `/setup.sh)"

export GOPROXY="` + proxyURL + `"
export GONOSUMDB="*"
export GONOPROXY=""

# Output confirmation if running interactively
if [ -t 1 ]; then
    echo "✓ Go environment configured for Athens:" >&2
    echo "  GOPROXY=$GOPROXY" >&2
    echo "  GONOSUMDB=$GONOSUMDB" >&2
    echo "  GONOPROXY=$GONOPROXY" >&2
    echo "" >&2
    echo "To make permanent, add to ~/.bashrc or ~/.zshrc:" >&2
    echo "  echo 'export GOPROXY=\"` + proxyURL + `\"' >> ~/.bashrc" >&2
    echo "  echo 'export GONOSUMDB=\"*\"' >> ~/.bashrc" >&2
    echo "  echo 'export GONOPROXY=\"\"' >> ~/.bashrc" >&2
fi
`

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(script))
}
