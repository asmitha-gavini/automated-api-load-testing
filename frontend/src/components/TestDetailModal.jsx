import React from 'react';
import {
  X,
  Clock,
  Gauge,
  CheckCircle2,
  AlertTriangle,
  FileJson,
  FileSpreadsheet,
  FileText,
  Printer,
  Trash2,
  ExternalLink,
  Zap,
  Activity,
  Layers,
  Calendar,
} from 'lucide-react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  Cell,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';
import { getExportReportUrl } from '../services/api';

export function TestDetailModal({ test, onClose, onDelete }) {
  if (!test) return null;

  const total = test.total_requests || 0;
  const success = test.successful_requests || 0;
  const failed = test.failed_requests || 0;
  const errorRate = test.error_rate !== undefined ? test.error_rate.toFixed(1) : '0.0';
  const successRate = total > 0 ? ((success / total) * 100).toFixed(1) : '100.0';
  const duration = test.duration_seconds || 1;
  const rps = (total / duration).toFixed(1);

  const getStatusBadge = (status) => {
    switch ((status || '').toLowerCase()) {
      case 'completed':
        return <span className="status-badge badge-completed">Completed</span>;
      case 'stopped':
        return <span className="status-badge badge-stopped">Stopped</span>;
      case 'failed':
        return <span className="status-badge badge-failed">Failed</span>;
      default:
        return <span className="status-badge badge-idle">{status || 'Idle'}</span>;
    }
  };

  const getMethodBadgeClass = (method) => {
    switch ((method || '').toUpperCase()) {
      case 'GET': return 'badge-method-get';
      case 'POST': return 'badge-method-post';
      case 'PUT': return 'badge-method-put';
      case 'DELETE': return 'badge-method-delete';
      default: return 'badge-method-default';
    }
  };

  const handlePrint = () => {
    window.print();
  };

  return (
    <div className="modal-backdrop" onClick={onClose}>
      <div className="modal-content print-container" onClick={(e) => e.stopPropagation()}>
        {/* Modal Header */}
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', flexWrap: 'wrap' }}>
            <span className={`method-badge ${getMethodBadgeClass(test.method)}`}>
              {test.method || 'GET'}
            </span>
            <h2 className="modal-title">{test.name || 'Benchmark Test Details'}</h2>
            {getStatusBadge(test.status)}
          </div>
          <button className="btn-icon no-print" onClick={onClose} title="Close Modal">
            <X size={20} />
          </button>
        </div>

        {/* Target & Config Info */}
        <div className="modal-section meta-banner">
          <div className="meta-grid">
            <div>
              <span className="meta-label">Target Endpoint</span>
              <div className="meta-value code-font" style={{ wordBreak: 'break-all' }}>
                {test.target_url}
              </div>
            </div>
            <div>
              <span className="meta-label">Test Run ID</span>
              <div className="meta-value code-font">{test.id}</div>
            </div>
            <div>
              <span className="meta-label">Execution Time</span>
              <div className="meta-value">
                <Calendar size={13} style={{ display: 'inline', marginRight: '4px', verticalAlign: '-1px' }} />
                {test.started_at ? new Date(test.started_at).toLocaleString() : 'N/A'}
              </div>
            </div>
            <div>
              <span className="meta-label">Configuration</span>
              <div className="meta-value">
                {test.virtual_users} Virtual Users &bull; {test.duration_seconds}s Duration
              </div>
            </div>
          </div>
        </div>

        {/* Primary Metrics Grid */}
        <div className="modal-section">
          <h3 className="section-subtitle">Performance Overview</h3>
          <div className="metrics-grid">
            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(56, 189, 248, 0.12)' }}>
                <Activity size={20} color="#38bdf8" />
              </div>
              <div className="metric-value">{total.toLocaleString()}</div>
              <div className="metric-label">Total Requests</div>
              <div className="metric-subtext">{rps} req/sec throughput</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(16, 185, 129, 0.12)' }}>
                <CheckCircle2 size={20} color="#10b981" />
              </div>
              <div className="metric-value">{success.toLocaleString()}</div>
              <div className="metric-label">Successful (2xx/3xx)</div>
              <div className="metric-subtext">{successRate}% pass rate</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: failed > 0 ? 'rgba(239, 68, 68, 0.15)' : 'rgba(148, 163, 184, 0.1)' }}>
                <AlertTriangle size={20} color={failed > 0 ? '#ef4444' : '#94a3b8'} />
              </div>
              <div className="metric-value" style={{ color: failed > 0 ? '#ef4444' : 'inherit' }}>
                {failed.toLocaleString()}
              </div>
              <div className="metric-label">Failed Requests</div>
              <div className="metric-subtext">Error rate: {errorRate}%</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(129, 140, 248, 0.15)' }}>
                <Clock size={20} color="#818cf8" />
              </div>
              <div className="metric-value">{(test.avg_latency_ms || 0).toFixed(1)} ms</div>
              <div className="metric-label">Average Latency</div>
              <div className="metric-subtext">Min: {(test.min_latency_ms || 0).toFixed(1)}ms &bull; Max: {(test.max_latency_ms || 0).toFixed(1)}ms</div>
            </div>
          </div>
        </div>

        {/* SLA Latency Percentiles */}
        <div className="modal-section">
          <h3 className="section-subtitle">
            <Gauge size={16} color="#f59e0b" style={{ display: 'inline', marginRight: '6px', verticalAlign: '-2px' }} />
            Latency Percentiles (SLA Breakdown)
          </h3>

          {/* Mini Percentiles Bar Chart */}
          <div style={{ height: 160, marginBottom: '1rem', background: 'rgba(15, 23, 42, 0.4)', borderRadius: 'var(--radius-md)', padding: '0.5rem' }}>
            <ResponsiveContainer width="100%" height="100%">
              <BarChart
                data={[
                  { tier: 'P50 (Median)', latency: test.p50_latency_ms || 0 },
                  { tier: 'P90 (90th)', latency: test.p90_latency_ms || 0 },
                  { tier: 'P95 (SLA)', latency: test.p95_latency_ms || 0 },
                  { tier: 'P99 (Tail)', latency: test.p99_latency_ms || 0 },
                ]}
                margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.25)" />
                <XAxis dataKey="tier" stroke="#64748b" tick={{ fontSize: 10 }} />
                <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
                <Tooltip
                  contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '11px' }}
                  formatter={(val) => [`${Number(val).toFixed(2)} ms`, 'Latency']}
                />
                <Bar dataKey="latency" radius={[4, 4, 0, 0]} isAnimationActive={false}>
                  <Cell fill="#38bdf8" />
                  <Cell fill="#818cf8" />
                  <Cell fill="#f59e0b" />
                  <Cell fill="#ec4899" />
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </div>

          <div className="percentiles-table-wrapper">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Percentile Tier</th>
                  <th>Latency (ms)</th>
                  <th>SLA Compliance Target</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                <tr>
                  <td><strong>P50 (Median)</strong></td>
                  <td className="code-font">{(test.p50_latency_ms || 0).toFixed(2)} ms</td>
                  <td>&lt; 50 ms</td>
                  <td>
                    {(test.p50_latency_ms || 0) < 50 ? (
                      <span className="badge-sla badge-sla-pass">Pass</span>
                    ) : (
                      <span className="badge-sla badge-sla-warn">High</span>
                    )}
                  </td>
                </tr>
                <tr>
                  <td><strong>P90 (90th Percentile)</strong></td>
                  <td className="code-font">{(test.p90_latency_ms || 0).toFixed(2)} ms</td>
                  <td>&lt; 100 ms</td>
                  <td>
                    {(test.p90_latency_ms || 0) < 100 ? (
                      <span className="badge-sla badge-sla-pass">Pass</span>
                    ) : (
                      <span className="badge-sla badge-sla-warn">Attention</span>
                    )}
                  </td>
                </tr>
                <tr>
                  <td><strong>P95 (SLA Threshold)</strong></td>
                  <td className="code-font">{(test.p95_latency_ms || 0).toFixed(2)} ms</td>
                  <td>&lt; 200 ms</td>
                  <td>
                    {(test.p95_latency_ms || 0) < 200 ? (
                      <span className="badge-sla badge-sla-pass">Pass</span>
                    ) : (
                      <span className="badge-sla badge-sla-fail">Breached</span>
                    )}
                  </td>
                </tr>
                <tr>
                  <td><strong>P99 (Tail Outlier)</strong></td>
                  <td className="code-font">{(test.p99_latency_ms || 0).toFixed(2)} ms</td>
                  <td>&lt; 500 ms</td>
                  <td>
                    {(test.p99_latency_ms || 0) < 500 ? (
                      <span className="badge-sla badge-sla-pass">Pass</span>
                    ) : (
                      <span className="badge-sla badge-sla-fail">Breached</span>
                    )}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        {/* Modal Actions / Export Controls */}
        <div className="modal-footer no-print">
          <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
            <a
              href={getExportReportUrl(test.id, 'json')}
              download={`report-${test.id}.json`}
              className="btn btn-secondary"
              title="Download full JSON dataset"
            >
              <FileJson size={15} color="#38bdf8" />
              <span>JSON Report</span>
            </a>
            <a
              href={getExportReportUrl(test.id, 'csv')}
              download={`report-${test.id}.csv`}
              className="btn btn-secondary"
              title="Download CSV spreadsheet"
            >
              <FileSpreadsheet size={15} color="#10b981" />
              <span>CSV Data</span>
            </a>
            <a
              href={getExportReportUrl(test.id, 'markdown')}
              download={`report-${test.id}.md`}
              className="btn btn-secondary"
              title="Download Markdown summary"
            >
              <FileText size={15} color="#f59e0b" />
              <span>Markdown</span>
            </a>
            <button className="btn btn-secondary" onClick={handlePrint} title="Print or Save as PDF">
              <Printer size={15} />
              <span>Print / PDF</span>
            </button>
          </div>

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {onDelete && (
              <button
                className="btn btn-danger"
                onClick={() => {
                  if (window.confirm(`Are you sure you want to permanently delete test "${test.name}"?`)) {
                    onDelete(test.id);
                  }
                }}
              >
                <Trash2 size={15} />
                <span>Delete</span>
              </button>
            )}
            <button className="btn btn-primary" onClick={onClose}>
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
