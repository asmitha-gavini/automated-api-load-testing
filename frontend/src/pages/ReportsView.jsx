import React, { useState, useEffect } from 'react';
import {
  FileText,
  Download,
  Printer,
  FileSpreadsheet,
  FileJson,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Gauge,
  Activity,
  Layers,
  Sparkles,
  TrendingUp,
} from 'lucide-react';
import { fetchTestHistory, getExportReportUrl } from '../services/api';

export function ReportsView() {
  const [tests, setTests] = useState([]);
  const [selectedTestId, setSelectedTestId] = useState('');
  const [summary, setSummary] = useState(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      setIsLoading(true);
      try {
        const data = await fetchTestHistory({ limit: 100, sortBy: 'started_at', order: 'DESC' });
        const list = data.tests || [];
        setTests(list);
        setSummary(data.summary || null);
        if (list.length > 0) {
          setSelectedTestId(list[0].id);
        }
      } catch (err) {
        console.error('Failed to load reports data:', err);
      } finally {
        setIsLoading(false);
      }
    }
    loadData();
  }, []);

  const selectedTest = tests.find((t) => t.id === selectedTestId) || tests[0] || null;

  const total = selectedTest?.total_requests || 0;
  const success = selectedTest?.successful_requests || 0;
  const failed = selectedTest?.failed_requests || 0;
  const passPct = total > 0 ? ((success / total) * 100).toFixed(1) : '100.0';
  const duration = selectedTest?.duration_seconds || 1;
  const throughput = (total / duration).toFixed(1);

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="reports-page">
      {/* Page Header */}
      <div className="card no-print" style={{ marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <FileText size={20} color="#38bdf8" />
              <h2 className="card-title" style={{ fontSize: '1.25rem', marginBottom: 0 }}>
                Performance Benchmark Reports
              </h2>
            </div>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
              Generate, preview, and export executive load testing summaries and latency SLA reports.
            </p>
          </div>

          {/* Test Selector Dropdown */}
          {tests.length > 0 && (
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
              <label style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>Select Benchmark Run:</label>
              <select
                className="input-field"
                style={{ width: 'auto', minWidth: '240px' }}
                value={selectedTestId}
                onChange={(e) => setSelectedTestId(e.target.value)}
              >
                {tests.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name} ({t.method} &bull; {new Date(t.started_at).toLocaleDateString()})
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>
      </div>

      {isLoading ? (
        <div className="card loading-state">Loading benchmark reports...</div>
      ) : !selectedTest ? (
        <div className="card empty-state">
          <FileText size={36} color="#64748b" style={{ marginBottom: '0.75rem' }} />
          <h3>No Test Reports Available</h3>
          <p>Execute load tests from the Dashboard to generate exportable reports here.</p>
        </div>
      ) : (
        <div className="report-container">
          {/* Export Action Controls */}
          <div className="card no-print" style={{ marginBottom: '1.25rem', padding: '0.85rem 1.25rem' }}>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '0.75rem' }}>
              <span style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)' }}>
                Export Selected Report:
              </span>
              <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
                <a
                  href={getExportReportUrl(selectedTest.id, 'json')}
                  download={`report-${selectedTest.id}.json`}
                  className="btn btn-secondary btn-sm"
                  title="Export raw JSON"
                >
                  <FileJson size={14} color="#38bdf8" />
                  <span>Download JSON</span>
                </a>
                <a
                  href={getExportReportUrl(selectedTest.id, 'csv')}
                  download={`report-${selectedTest.id}.csv`}
                  className="btn btn-secondary btn-sm"
                  title="Export CSV metrics table"
                >
                  <FileSpreadsheet size={14} color="#10b981" />
                  <span>Download CSV</span>
                </a>
                <a
                  href={getExportReportUrl(selectedTest.id, 'markdown')}
                  download={`report-${selectedTest.id}.md`}
                  className="btn btn-secondary btn-sm"
                  title="Export GitHub Markdown"
                >
                  <FileText size={14} color="#f59e0b" />
                  <span>Download Markdown</span>
                </a>
                <button className="btn btn-primary btn-sm" onClick={handlePrint}>
                  <Printer size={14} />
                  <span>Print / Save PDF</span>
                </button>
              </div>
            </div>
          </div>

          {/* Printable Report Document */}
          <div className="card printable-report print-document">
            {/* Report Header */}
            <div className="report-doc-header">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', borderBottom: '2px solid var(--border-subtle)', paddingBottom: '1rem' }}>
                <div>
                  <h1 style={{ fontSize: '1.5rem', fontWeight: 800, color: 'var(--text-primary)' }}>
                    Executive Benchmark Assessment
                  </h1>
                  <div style={{ fontSize: '0.9rem', color: 'var(--text-accent)', fontWeight: 600, marginTop: '0.2rem' }}>
                    {selectedTest.name}
                  </div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>Report Generated:</div>
                  <div style={{ fontSize: '0.85rem', fontWeight: 600 }}>{new Date().toLocaleString()}</div>
                </div>
              </div>

              {/* Benchmark Metadata Table */}
              <div className="report-meta-grid" style={{ marginTop: '1rem', display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '1rem', background: 'rgba(15, 23, 42, 0.6)', padding: '1rem', borderRadius: 'var(--radius-md)' }}>
                <div>
                  <span className="meta-label">Target Endpoint</span>
                  <div className="code-font" style={{ fontWeight: 600, fontSize: '0.85rem', color: 'var(--text-primary)' }}>
                    [{selectedTest.method}] {selectedTest.target_url}
                  </div>
                </div>
                <div>
                  <span className="meta-label">Run Identification</span>
                  <div className="code-font" style={{ fontSize: '0.8rem' }}>{selectedTest.id}</div>
                </div>
                <div>
                  <span className="meta-label">Concurrency & Duration</span>
                  <div style={{ fontSize: '0.85rem', fontWeight: 600 }}>
                    {selectedTest.virtual_users} Virtual Users &bull; {selectedTest.duration_seconds} Seconds
                  </div>
                </div>
                <div>
                  <span className="meta-label">Result Assessment</span>
                  <div style={{ fontSize: '0.85rem', fontWeight: 700, color: selectedTest.status === 'completed' ? '#34d399' : '#fbbf24' }}>
                    {selectedTest.status.toUpperCase()}
                  </div>
                </div>
              </div>
            </div>

            {/* Executive KPIs */}
            <div style={{ marginTop: '1.5rem' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 700, marginBottom: '0.75rem', color: 'var(--text-primary)' }}>
                Core Performance KPIs
              </h3>
              <div className="metrics-grid">
                <div className="metric-card">
                  <div className="metric-value">{total.toLocaleString()}</div>
                  <div className="metric-label">Dispatched Requests</div>
                  <div className="metric-subtext">{throughput} req/s throughput</div>
                </div>
                <div className="metric-card">
                  <div className="metric-value" style={{ color: '#10b981' }}>{success.toLocaleString()}</div>
                  <div className="metric-label">Successful Requests</div>
                  <div className="metric-subtext">{passPct}% SLA pass rate</div>
                </div>
                <div className="metric-card">
                  <div className="metric-value" style={{ color: failed > 0 ? '#ef4444' : '#10b981' }}>
                    {(selectedTest.error_rate || 0).toFixed(1)}%
                  </div>
                  <div className="metric-label">Error Rate</div>
                  <div className="metric-subtext">{failed.toLocaleString()} failures</div>
                </div>
                <div className="metric-card">
                  <div className="metric-value">{(selectedTest.avg_latency_ms || 0).toFixed(1)} ms</div>
                  <div className="metric-label">Average Response Time</div>
                  <div className="metric-subtext">Min: {(selectedTest.min_latency_ms || 0).toFixed(1)}ms &bull; Max: {(selectedTest.max_latency_ms || 0).toFixed(1)}ms</div>
                </div>
              </div>
            </div>

            {/* SLA Percentiles Table */}
            <div style={{ marginTop: '1.75rem' }}>
              <h3 style={{ fontSize: '1rem', fontWeight: 700, marginBottom: '0.75rem', color: 'var(--text-primary)' }}>
                Latency SLA Distribution
              </h3>
              <table className="data-table">
                <thead>
                  <tr>
                    <th>Percentile Tier</th>
                    <th>Measured Latency</th>
                    <th>Target SLA Limit</th>
                    <th>Compliance</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td><strong>P50 (Median Response)</strong></td>
                    <td className="code-font">{(selectedTest.p50_latency_ms || 0).toFixed(2)} ms</td>
                    <td>&lt; 50 ms</td>
                    <td>
                      {(selectedTest.p50_latency_ms || 0) < 50 ? (
                        <span className="badge-sla badge-sla-pass">Compliant</span>
                      ) : (
                        <span className="badge-sla badge-sla-warn">Elevated</span>
                      )}
                    </td>
                  </tr>
                  <tr>
                    <td><strong>P90 (90th Percentile)</strong></td>
                    <td className="code-font">{(selectedTest.p90_latency_ms || 0).toFixed(2)} ms</td>
                    <td>&lt; 100 ms</td>
                    <td>
                      {(selectedTest.p90_latency_ms || 0) < 100 ? (
                        <span className="badge-sla badge-sla-pass">Compliant</span>
                      ) : (
                        <span className="badge-sla badge-sla-warn">Warning</span>
                      )}
                    </td>
                  </tr>
                  <tr>
                    <td><strong>P95 (95th Percentile)</strong></td>
                    <td className="code-font">{(selectedTest.p95_latency_ms || 0).toFixed(2)} ms</td>
                    <td>&lt; 200 ms</td>
                    <td>
                      {(selectedTest.p95_latency_ms || 0) < 200 ? (
                        <span className="badge-sla badge-sla-pass">Compliant</span>
                      ) : (
                        <span className="badge-sla badge-sla-fail">Breached</span>
                      )}
                    </td>
                  </tr>
                  <tr>
                    <td><strong>P99 (Tail Latency)</strong></td>
                    <td className="code-font">{(selectedTest.p99_latency_ms || 0).toFixed(2)} ms</td>
                    <td>&lt; 500 ms</td>
                    <td>
                      {(selectedTest.p99_latency_ms || 0) < 500 ? (
                        <span className="badge-sla badge-sla-pass">Compliant</span>
                      ) : (
                        <span className="badge-sla badge-sla-fail">Breached</span>
                      )}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* Benchmark Analysis Notes */}
            <div style={{ marginTop: '1.75rem', padding: '1rem', border: '1px solid var(--border-subtle)', borderRadius: 'var(--radius-md)', background: 'rgba(15, 23, 42, 0.4)' }}>
              <h4 style={{ fontSize: '0.85rem', fontWeight: 700, color: 'var(--text-secondary)', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
                Engineering Assessment Notes
              </h4>
              <ul style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', paddingLeft: '1.25rem', marginTop: '0.5rem', lineHeight: '1.6' }}>
                <li>
                  Total throughput sustained at <strong>{throughput} requests/second</strong> across {selectedTest.virtual_users} parallel virtual user goroutines.
                </li>
                <li>
                  {selectedTest.error_rate === 0 ? (
                    <span style={{ color: '#10b981' }}>Zero request failures observed. Target API exhibited 100% stability.</span>
                  ) : (
                    <span style={{ color: '#ef4444' }}>Observed error rate of {selectedTest.error_rate.toFixed(1)}%. Inspect error logs and upstream dependencies.</span>
                  )}
                </li>
                <li>
                  95% of all client calls resolved within <strong>{(selectedTest.p95_latency_ms || 0).toFixed(1)} ms</strong>.
                </li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
