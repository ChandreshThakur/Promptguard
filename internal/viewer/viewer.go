package viewer

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"promptguard/internal/runner"
)

// Server provides the web interface for viewing test results
type Server struct {
	resultsFile string
	mux         *http.ServeMux
}

// NewServer creates a new viewer server
func NewServer(resultsFile string) *Server {
	server := &Server{
		resultsFile: resultsFile,
		mux:         http.NewServeMux(),
	}

	server.setupRoutes()
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/results", s.handleAPIResults)
	s.mux.HandleFunc("/api/diff", s.handleAPIDiff)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PromptGuard Viewer</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 0; background: #f5f7fa; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 20px 0; text-align: center; }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        .controls { background: white; padding: 20px; border-radius: 8px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .results-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        .results-panel { background: white; border-radius: 8px; padding: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .test-item { border: 1px solid #e1e5e9; border-radius: 6px; margin-bottom: 15px; overflow: hidden; }
        .test-header { padding: 15px; background: #f8f9fa; border-bottom: 1px solid #e1e5e9; cursor: pointer; }
        .test-content { padding: 15px; display: none; }
        .test-content.show { display: block; }
        .status-badge { padding: 3px 8px; border-radius: 12px; font-size: 0.7em; font-weight: bold; text-transform: uppercase; }
        .badge-passed { background: #d4edda; color: #155724; }
        .badge-failed { background: #f8d7da; color: #721c24; }
        .diff-viewer { background: #f8f9fa; border-radius: 4px; padding: 15px; margin: 10px 0; }
        .response-text { font-family: monospace; white-space: pre-wrap; background: #f1f3f4; padding: 10px; border-radius: 4px; }
        .metrics-chart { height: 300px; margin: 20px 0; }
        button { background: #667eea; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; }
        button:hover { background: #5a67d8; }
        .tab-buttons { display: flex; gap: 10px; margin-bottom: 20px; }
        .tab-buttons button { background: #e2e8f0; color: #4a5568; }
        .tab-buttons button.active { background: #667eea; color: white; }
    </style>
</head>
<body>
    <div class="header">
        <h1>PromptGuard Interactive Viewer</h1>
        <p>Explore test results, compare baselines, and analyze prompt performance</p>
    </div>

    <div class="container">
        <div class="controls">
            <div class="tab-buttons">
                <button id="results-tab" class="active" onclick="showTab('results')">Test Results</button>
                <button id="diff-tab" onclick="showTab('diff')">Baseline Comparison</button>
                <button id="metrics-tab" onclick="showTab('metrics')">Historical Metrics</button>
            </div>
            
            <div id="results-controls">
                <button onclick="loadResults()">Refresh Results</button>
                <button onclick="exportResults()">Export Report</button>
            </div>
            
            <div id="diff-controls" style="display: none;">
                <button onclick="loadBaseline()">Load Baseline</button>
                <button onclick="compareResults()">Compare with Current</button>
            </div>
        </div>

        <div id="results-view">
            <div class="results-grid">
                <div class="results-panel">
                    <h3>Current Results</h3>
                    <div id="current-results">Loading...</div>
                </div>
                <div class="results-panel">
                    <h3>Test Details</h3>
                    <div id="test-details">Select a test to view details</div>
                </div>
            </div>
        </div>

        <div id="diff-view" style="display: none;">
            <div class="results-panel">
                <h3>Baseline vs Current Comparison</h3>
                <div id="diff-content">No comparison data available</div>
            </div>
        </div>

        <div id="metrics-view" style="display: none;">
            <div class="results-panel">
                <h3>Historical Performance</h3>
                <div class="metrics-chart" id="cost-chart"></div>
                <div class="metrics-chart" id="success-chart"></div>
            </div>
        </div>
    </div>

    <script>
        let currentResults = null;

        async function loadResults() {
            try {
                const response = await fetch('/api/results');
                currentResults = await response.json();
                displayResults(currentResults);
            } catch (error) {
                console.error('Failed to load results:', error);
                document.getElementById('current-results').innerHTML = 'Error loading results';
            }
        }

        function displayResults(results) {
            const container = document.getElementById('current-results');
            
            let html = '<div class="summary">';
            html += '<h4>Summary</h4>';
            html += '<p><strong>Total:</strong> ' + results.total + '</p>';
            html += '<p><strong>Passed:</strong> ' + results.passed + '</p>';
            html += '<p><strong>Failed:</strong> ' + results.failed + '</p>';
            html += '<p><strong>Cost:</strong> $' + results.totalCost.toFixed(4) + '</p>';
            html += '</div>';

            html += '<div class="test-list">';
            results.testResults.forEach((test, index) => {
                const statusClass = test.status === 'passed' ? 'badge-passed' : 'badge-failed';
                html += '<div class="test-item">';
                html += '<div class="test-header" onclick="toggleTest(' + index + '); showTestDetails(' + index + ')">';
                html += '<span><strong>' + test.name + '</strong></span>';
                html += '<span class="status-badge ' + statusClass + '">' + test.status + '</span>';
                html += '</div>';
                html += '<div id="test-' + index + '" class="test-content">';
                html += '<p><strong>Provider:</strong> ' + test.provider + '</p>';
                html += '<p><strong>Cost:</strong> $' + test.cost.toFixed(4) + '</p>';
                html += '<div class="response-text">' + test.response + '</div>';
                html += '</div>';
                html += '</div>';
            });
            html += '</div>';

            container.innerHTML = html;
        }

        function showTestDetails(index) {
            if (!currentResults) return;
            
            const test = currentResults.testResults[index];
            const container = document.getElementById('test-details');
            
            let html = '<h4>' + test.name + '</h4>';
            html += '<p><strong>File:</strong> ' + test.promptFile + '</p>';
            html += '<p><strong>Provider:</strong> ' + test.provider + '</p>';
            html += '<p><strong>Duration:</strong> ' + test.duration + '</p>';
            
            if (test.error) {
                html += '<div style="color: red;"><strong>Error:</strong> ' + test.error + '</div>';
            }
            
            html += '<h5>Assertions</h5>';
            test.assertions.forEach(assertion => {
                const status = assertion.passed ? '✅' : '❌';
                html += '<div>' + status + ' <strong>' + assertion.type + ':</strong> ' + assertion.message + '</div>';
            });
            
            html += '<h5>Response</h5>';
            html += '<div class="response-text">' + test.response + '</div>';
            
            container.innerHTML = html;
        }

        function toggleTest(index) {
            const content = document.getElementById('test-' + index);
            content.classList.toggle('show');
        }

        function showTab(tabName) {
            // Hide all views
            document.getElementById('results-view').style.display = 'none';
            document.getElementById('diff-view').style.display = 'none';
            document.getElementById('metrics-view').style.display = 'none';
            document.getElementById('results-controls').style.display = 'none';
            document.getElementById('diff-controls').style.display = 'none';
            
            // Remove active class from all tabs
            document.querySelectorAll('.tab-buttons button').forEach(btn => btn.classList.remove('active'));
            
            // Show selected view and controls
            document.getElementById(tabName + '-view').style.display = 'block';
            document.getElementById(tabName + '-controls').style.display = 'block';
            document.getElementById(tabName + '-tab').classList.add('active');
        }

        function exportResults() {
            if (!currentResults) return;
            
            const dataStr = JSON.stringify(currentResults, null, 2);
            const dataBlob = new Blob([dataStr], {type: 'application/json'});
            const url = URL.createObjectURL(dataBlob);
            const link = document.createElement('a');
            link.href = url;
            link.download = 'promptguard-results.json';
            link.click();
        }

        // Load results on page load
        loadResults();
    </script>
</body>
</html>`

	t, err := template.New("index").Parse(tmpl)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, nil)
}

func (s *Server) handleAPIResults(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(s.resultsFile)
	if err != nil {
		http.Error(w, "Failed to read results file", http.StatusInternalServerError)
		return
	}

	var results runner.Results
	if err := json.Unmarshal(data, &results); err != nil {
		http.Error(w, "Failed to parse results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) handleAPIDiff(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement baseline comparison
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Diff functionality not yet implemented"}`))
}
