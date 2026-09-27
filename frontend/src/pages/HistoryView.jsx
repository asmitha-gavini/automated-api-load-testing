import React, { useState, useEffect, useCallback } from 'react';
import {
  History,
  Search,
  Filter,
  ArrowUpDown,
  Trash2,
  RefreshCw,
  ExternalLink,
  FileJson,
  FileSpreadsheet,
  FileText,
  Calendar,
  Layers,
  Activity,
  CheckCircle2,
  AlertTriangle,
  Clock,
  Gauge,
  SlidersHorizontal,
  Scale,
} from 'lucide-react';
import { fetchTestHistory, deleteTest, clearAllTests, getExportReportUrl } from '../services/api';
import { TestDetailModal } from '../components/TestDetailModal';

export function HistoryView({ onNavigateToComparison }) {
  const [tests, setTests] = useState([]);
  const [summary, setSummary] = useState(null);
  const [totalCount, setTotalCount] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  // Filters and sorting
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('all');
  const [methodFilter, setMethodFilter] = useState('all');
  const [sortBy, setSortBy] = useState('started_at');
  const [sortOrder, setSortOrder] = useState('DESC');

  // Multi-select for comparison
  const [checkedIds, setCheckedIds] = useState([]);

  // Selected test for detail modal
  const [selectedTest, setSelectedTest] = useState(null);

  const loadHistory = useCallback(async () => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await fetchTestHistory({
        search: search.trim(),
        status: statusFilter,
        method: methodFilter,
        sortBy,
        order: sortOrder,
        limit: 100,
        offset: 0,
      });
      setTests(data.tests || []);
      setTotalCount(data.total || 0);
      setSummary(data.summary || null);
    } catch (err) {
      setError(err.message || 'Failed to load test history');
    } finally {
      setIsLoading(false);
    }
  }, [search, statusFilter, methodFilter, sortBy, sortOrder]);

  useEffect(() => {
    loadHistory();
  }, [loadHistory]);

  const handleDelete = async (id) => {
    try {
      await deleteTest(id);
      setCheckedIds(checkedIds.filter((item) => item !== id));
      if (selectedTest?.id === id) {
        setSelectedTest(null);
      }
      loadHistory();
    } catch (err) {
      alert(`Delete error: ${err.message}`);
    }
  };

  const handleClearAll = async () => {
    if (window.confirm('Are you sure you want to permanently clear ALL test history? This action cannot be undone.')) {
      try {
        await clearAllTests();
        setCheckedIds([]);
        setSelectedTest(null);
        loadHistory();
      } catch (err) {
        alert(`Clear all error: ${err.message}`);
      }
    }
  };

  const handleToggleCheck = (id, e) => {
    e.stopPropagation();
    if (checkedIds.includes(id)) {
      setCheckedIds(checkedIds.filter((item) => item !== id));
    } else {
      if (checkedIds.length < 4) {
        setCheckedIds([...checkedIds, id]);
      } else {
        alert('You can select up to 4 tests for comparison.');
      }
    }
  };

  const handleCompareClick = () => {
    if (onNavigateToComparison && checkedIds.length >= 2) {
      onNavigateToComparison(checkedIds);
    }
  };

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

  return (
    <div className="history-page">
      {/* Top Summary Banner */}
      {summary && summary.total_tests > 0 && (
        <div className="card summary-banner">
          <div className="summary-banner-grid">
            <div className="summary-item">
              <span className="summary-label">Total Benchmarks</span>
              <span className="summary-value" style={{ color: '#38bdf8' }}>{summary.total_tests}</span>
              <span className="summary-subtext">{summary.completed_tests} completed, {summary.stopped_tests} stopped</span>
            </div>
            <div className="summary-item">
              <span className="summary-label">Dispatched Requests</span>
              <span className="summary-value" style={{ color: '#10b981' }}>{summary.total_requests.toLocaleString()}</span>
              <span className="summary-subtext">{summary.successful_requests.toLocaleString()} successful</span>
            </div>
            <div className="summary-item">
              <span className="summary-label">Overall Error Rate</span>
              <span className="summary-value" style={{ color: summary.overall_error_rate > 5 ? '#f43f5e' : '#10b981' }}>
                {summary.overall_error_rate.toFixed(1)}%
              </span>
              <span className="summary-subtext">{summary.failed_requests.toLocaleString()} total failures</span>
            </div>
            <div className="summary-item">
              <span className="summary-label">Average Response Time</span>
              <span className="summary-value" style={{ color: '#818cf8' }}>
                {summary.avg_latency_ms.toFixed(1)} ms
              </span>
              <span className="summary-subtext">Min: {summary.min_latency_ms}ms &bull; Max: {summary.max_latency_ms}ms</span>
            </div>
          </div>
        </div>
      )}

      {/* Control & Filter Toolbar */}
      <div className="card toolbar-card">
        <div className="toolbar-wrapper">
          {/* Search box */}
          <div className="search-input-wrapper">
            <Search size={16} className="search-icon" />
            <input
              type="text"
              placeholder="Search tests by name or target URL..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="search-input"
            />
          </div>

          {/* Status Filter Chips */}
          <div className="status-filters">
            {['all', 'completed', 'stopped', 'failed'].map((st) => (
              <button
                key={st}
                className={`filter-chip ${statusFilter === st ? 'active' : ''}`}
                onClick={() => setStatusFilter(st)}
              >
                {st.charAt(0).toUpperCase() + st.slice(1)}
              </button>
            ))}
          </div>

          {/* Method Filter */}
          <div className="sort-controls">
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>Method:</span>
            <select
              className="sort-select"
              value={methodFilter}
              onChange={(e) => setMethodFilter(e.target.value)}
            >
              <option value="all">All Methods</option>
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="DELETE">DELETE</option>
            </select>
          </div>

          {/* Sort Selector */}
          <div className="sort-controls">
            <SlidersHorizontal size={14} color="#94a3b8" />
            <select
              className="sort-select"
              value={`${sortBy}-${sortOrder}`}
              onChange={(e) => {
                const [sb, so] = e.target.value.split('-');
                setSortBy(sb);
                setSortOrder(so);
              }}
            >
              <option value="started_at-DESC">Date (Newest First)</option>
              <option value="started_at-ASC">Date (Oldest First)</option>
              <option value="total_requests-DESC">Requests (High to Low)</option>
              <option value="duration_seconds-DESC">Duration (Longest First)</option>
              <option value="duration_seconds-ASC">Duration (Shortest First)</option>
              <option value="avg_latency_ms-ASC">Latency (Lowest First)</option>
              <option value="error_rate-ASC">Error Rate (Lowest First)</option>
              <option value="error_rate-DESC">Error Rate (Highest First)</option>
            </select>
          </div>

          {/* Actions */}
          <div className="toolbar-actions">
            {checkedIds.length >= 2 && onNavigateToComparison && (
              <button
                className="btn btn-primary btn-sm"
                onClick={handleCompareClick}
                style={{ background: 'linear-gradient(90deg, #0284c7, #38bdf8)', color: '#090d16', fontWeight: 700 }}
              >
                <Scale size={14} />
                <span>Compare Selected ({checkedIds.length})</span>
              </button>
            )}
            <button className="btn-icon" onClick={loadHistory} title="Refresh test history">
              <RefreshCw size={16} />
            </button>
            {tests.length > 0 && (
              <button className="btn btn-secondary" onClick={handleClearAll} style={{ color: '#f87171' }} title="Clear all history">
                <Trash2 size={14} />
                <span>Clear All</span>
              </button>
            )}
          </div>
        </div>
      </div>

      {/* History Table / Records List */}
      <div className="card table-card">
        <div className="card-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <History size={18} color="#38bdf8" />
            <h2 className="card-title">Test Run History ({totalCount})</h2>
          </div>
          {checkedIds.length > 0 && (
            <span style={{ fontSize: '0.8rem', color: '#38bdf8' }}>
              {checkedIds.length} test{checkedIds.length > 1 ? 's' : ''} selected for comparison
            </span>
          )}
        </div>

        {error && (
          <div className="error-banner">
            <AlertTriangle size={16} />
            <span>{error}</span>
          </div>
        )}

        {isLoading ? (
          <div className="loading-state">Loading test history...</div>
        ) : tests.length === 0 ? (
          <div className="empty-history-state">
            <History size={36} color="#64748b" style={{ marginBottom: '0.75rem' }} />
            <h3 style={{ fontSize: '1rem', color: 'var(--text-secondary)' }}>No Test Runs Found</h3>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-muted)' }}>
              {search || statusFilter !== 'all' || methodFilter !== 'all'
                ? 'Try adjusting your search query or filter criteria.'
                : 'Run a load test from the Dashboard to record benchmark results here.'}
            </p>
          </div>
        ) : (
          <div className="table-responsive">
            <table className="data-table">
              <thead>
                <tr>
                  <th style={{ width: 40 }}>Compare</th>
                  <th>Status</th>
                  <th>Test Name & Endpoint</th>
                  <th>Config</th>
                  <th>Requests</th>
                  <th>Success %</th>
                  <th>Avg Latency</th>
                  <th>P95 SLA</th>
                  <th>Date & Time</th>
                  <th style={{ textAlign: 'right' }}>Actions</th>
                </tr>
              </thead>
              <tbody>
                {tests.map((run) => {
                  const passPct = run.total_requests > 0
                    ? ((run.successful_requests / run.total_requests) * 100).toFixed(1)
                    : '100.0';
                  const isChecked = checkedIds.includes(run.id);

                  return (
                    <tr
                      key={run.id}
                      className="history-row"
                      style={{ background: isChecked ? 'rgba(56, 189, 248, 0.08)' : undefined }}
                      onClick={() => setSelectedTest(run)}
                    >
                      <td onClick={(e) => e.stopPropagation()} style={{ textAlign: 'center' }}>
                        <input
                          type="checkbox"
                          checked={isChecked}
                          onChange={(e) => handleToggleCheck(run.id, e)}
                          title="Select for comparison"
                          style={{ cursor: 'pointer', accentColor: '#38bdf8' }}
                        />
                      </td>
                      <td>{getStatusBadge(run.status)}</td>
                      <td>
                        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                          <span className={`method-badge ${getMethodBadgeClass(run.method)}`}>
                            {run.method}
                          </span>
                          <span style={{ fontWeight: 600, color: 'var(--text-primary)' }}>
                            {run.name || 'Load Test'}
                          </span>
                        </div>
                        <div className="code-font text-muted" style={{ fontSize: '0.75rem', marginTop: '2px' }}>
                          {run.target_url}
                        </div>
                      </td>
                      <td>
                        <span style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                          {run.virtual_users} VUs &bull; {run.duration_seconds}s
                        </span>
                      </td>
                      <td className="code-font" style={{ fontWeight: 600 }}>
                        {run.total_requests.toLocaleString()}
                      </td>
                      <td>
                        <span style={{
                          color: Number(passPct) === 100 ? '#10b981' : (Number(passPct) >= 90 ? '#f59e0b' : '#ef4444'),
                          fontWeight: 600,
                        }}>
                          {passPct}%
                        </span>
                      </td>
                      <td className="code-font">{(run.avg_latency_ms || 0).toFixed(1)} ms</td>
                      <td className="code-font">{(run.p95_latency_ms || 0).toFixed(1)} ms</td>
                      <td style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                        {run.started_at ? new Date(run.started_at).toLocaleString([], { dateStyle: 'short', timeStyle: 'short' }) : 'N/A'}
                      </td>
                      <td style={{ textAlign: 'right' }} onClick={(e) => e.stopPropagation()}>
                        <div style={{ display: 'inline-flex', gap: '0.35rem' }}>
                          <button
                            className="btn btn-secondary btn-sm"
                            onClick={() => setSelectedTest(run)}
                            title="View test details"
                          >
                            Details
                          </button>
                          <a
                            href={getExportReportUrl(run.id, 'json')}
                            download={`report-${run.id}.json`}
                            className="btn btn-secondary btn-sm"
                            title="Download JSON report"
                          >
                            <FileJson size={13} color="#38bdf8" />
                          </a>
                          <a
                            href={getExportReportUrl(run.id, 'markdown')}
                            download={`report-${run.id}.md`}
                            className="btn btn-secondary btn-sm"
                            title="Download Markdown report"
                          >
                            <FileText size={13} color="#f59e0b" />
                          </a>
                          <button
                            className="btn btn-danger btn-sm"
                            onClick={() => {
                              if (window.confirm(`Delete test "${run.name}"?`)) {
                                handleDelete(run.id);
                              }
                            }}
                            title="Delete record"
                          >
                            <Trash2 size={13} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Test Detail Modal */}
      {selectedTest && (
        <TestDetailModal
          test={selectedTest}
          onClose={() => setSelectedTest(null)}
          onDelete={handleDelete}
        />
      )}
    </div>
  );
}
