import React, { useState, useEffect } from 'react';
import {
  Scale,
  TrendingUp,
  Clock,
  Gauge,
  CheckCircle2,
  AlertTriangle,
  ArrowRight,
  ArrowUpRight,
  ArrowDownRight,
  Printer,
  Sparkles,
  Layers,
  Activity,
  Award,
} from 'lucide-react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
  Cell,
} from 'recharts';
import { fetchTestHistory } from '../services/api';

export function ComparisonView({ preselectedIds = [], onBackToHistory }) {
  const [allTests, setAllTests] = useState([]);
  const [selectedIds, setSelectedIds] = useState(preselectedIds);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadData() {
      setIsLoading(true);
      try {
        const data = await fetchTestHistory({ limit: 100, sortBy: 'started_at', order: 'DESC' });
        const list = data.tests || [];
        setAllTests(list);
        if (selectedIds.length === 0 && list.length >= 2) {
          setSelectedIds([list[0].id, list[1].id]);
        } else if (selectedIds.length === 0 && list.length === 1) {
          setSelectedIds([list[0].id]);
        }
      } catch (err) {
        console.error('Failed to load tests for comparison:', err);
      } finally {
        setIsLoading(false);
      }
    }
    loadData();
  }, []);

  const handleToggleSelect = (id) => {
    if (selectedIds.includes(id)) {
      if (selectedIds.length > 1) {
        setSelectedIds(selectedIds.filter((item) => item !== id));
      }
    } else {
      if (selectedIds.length < 4) {
        setSelectedIds([...selectedIds, id]);
      } else {
        alert('You can compare up to 4 tests simultaneously.');
      }
    }
  };

  const comparedTests = allTests.filter((t) => selectedIds.includes(t.id));

  // Compute Winners
  const getWinner = (metricKey, lowerIsBetter = false) => {
    if (comparedTests.length < 2) return null;
    let best = comparedTests[0];
    for (let i = 1; i < comparedTests.length; i++) {
      const current = comparedTests[i];
      const bestVal = best[metricKey] || 0;
      const curVal = current[metricKey] || 0;
      if (lowerIsBetter ? curVal < bestVal : curVal > bestVal) {
        best = current;
      }
    }
    return best.id;
  };

  const fastestWinnerId = getWinner('avg_latency_ms', true);
  const lowestP95WinnerId = getWinner('p95_latency_ms', true);
  const highestRpsWinnerId = (() => {
    if (comparedTests.length < 2) return null;
    let best = comparedTests[0];
    let bestRps = (best.total_requests || 0) / (best.duration_seconds || 1);
    for (let i = 1; i < comparedTests.length; i++) {
      const cur = comparedTests[i];
      const curRps = (cur.total_requests || 0) / (cur.duration_seconds || 1);
      if (curRps > bestRps) {
        best = cur;
        bestRps = curRps;
      }
    }
    return best.id;
  })();
  const bestReliabilityWinnerId = getWinner('error_rate', true);

  // Chart data: Latency percentiles comparison
  const latencyChartData = [
    {
      metric: 'Avg Latency',
      ...comparedTests.reduce((acc, t, idx) => ({ ...acc, [`Test ${idx + 1}`]: t.avg_latency_ms || 0 }), {}),
    },
    {
      metric: 'P50 Median',
      ...comparedTests.reduce((acc, t, idx) => ({ ...acc, [`Test ${idx + 1}`]: t.p50_latency_ms || 0 }), {}),
    },
    {
      metric: 'P90 SLA',
      ...comparedTests.reduce((acc, t, idx) => ({ ...acc, [`Test ${idx + 1}`]: t.p90_latency_ms || 0 }), {}),
    },
    {
      metric: 'P95 SLA',
      ...comparedTests.reduce((acc, t, idx) => ({ ...acc, [`Test ${idx + 1}`]: t.p95_latency_ms || 0 }), {}),
    },
    {
      metric: 'P99 Tail',
      ...comparedTests.reduce((acc, t, idx) => ({ ...acc, [`Test ${idx + 1}`]: t.p99_latency_ms || 0 }), {}),
    },
  ];

  // Chart data: Throughput & Total Requests comparison
  const volumeChartData = comparedTests.map((t, idx) => ({
    name: `Test ${idx + 1}: ${t.name?.substring(0, 14) || t.id.substring(0, 8)}`,
    RPS: Number(((t.total_requests || 0) / (t.duration_seconds || 1)).toFixed(1)),
    Total: t.total_requests || 0,
    Success: t.successful_requests || 0,
  }));

  const testColors = ['#38bdf8', '#10b981', '#f59e0b', '#ec4899'];

  // Two-test comparative deltas
  const isPair = comparedTests.length === 2;
  const t1 = comparedTests[0];
  const t2 = comparedTests[1];

  const rps1 = t1 ? (t1.total_requests / (t1.duration_seconds || 1)) : 0;
  const rps2 = t2 ? (t2.total_requests / (t2.duration_seconds || 1)) : 0;
  const rpsDelta = rps2 - rps1;
  const rpsPct = rps1 > 0 ? ((rpsDelta / rps1) * 100).toFixed(1) : '0';

  const latDelta = t2 && t1 ? (t2.avg_latency_ms - t1.avg_latency_ms) : 0;
  const latPct = t1 && t1.avg_latency_ms > 0 ? ((latDelta / t1.avg_latency_ms) * 100).toFixed(1) : '0';

  const p95Delta = t2 && t1 ? (t2.p95_latency_ms - t1.p95_latency_ms) : 0;
  const p95Pct = t1 && t1.p95_latency_ms > 0 ? ((p95Delta / t1.p95_latency_ms) * 100).toFixed(1) : '0';

  const errDelta = t2 && t1 ? (t2.error_rate - t1.error_rate) : 0;

  return (
    <div className="comparison-page">
      {/* Header */}
      <div className="card no-print" style={{ marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <Scale size={20} color="#38bdf8" />
              <h2 className="card-title" style={{ fontSize: '1.25rem', marginBottom: 0 }}>
                Test Comparison & Differential Analysis
              </h2>
            </div>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
              Compare multiple load test runs side-by-side to evaluate optimizations and regressions.
            </p>
          </div>
          <div style={{ display: 'flex', gap: '0.5rem' }}>
            <button className="btn btn-secondary btn-sm" onClick={() => window.print()}>
              <Printer size={14} />
              <span>Print Report</span>
            </button>
            {onBackToHistory && (
              <button className="btn btn-primary btn-sm" onClick={onBackToHistory}>
                Back to History
              </button>
            )}
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="card loading-state">Loading benchmark comparison data...</div>
      ) : allTests.length < 2 ? (
        <div className="card empty-state">
          <Scale size={36} color="#64748b" style={{ marginBottom: '0.75rem' }} />
          <h3>At Least 2 Tests Required For Comparison</h3>
          <p>You have {allTests.length} test recorded. Execute additional tests to enable comparative analytics.</p>
        </div>
      ) : (
        <div>
          {/* Test Selector Strip */}
          <div className="card no-print" style={{ marginBottom: '1.25rem', padding: '1rem 1.25rem' }}>
            <div style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
              Select Tests to Compare (Choose 2 to 4):
            </div>
            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
              {allTests.map((t, idx) => {
                const isSelected = selectedIds.includes(t.id);
                const colorIdx = selectedIds.indexOf(t.id);
                return (
                  <button
                    key={t.id}
                    className={`btn ${isSelected ? 'btn-primary' : 'btn-secondary'} btn-sm`}
                    style={{
                      borderColor: isSelected ? testColors[colorIdx % testColors.length] : undefined,
                      boxShadow: isSelected ? `0 0 10px -2px ${testColors[colorIdx % testColors.length]}40` : undefined,
                    }}
                    onClick={() => handleToggleSelect(t.id)}
                  >
                    <span>{isSelected ? `✓ Test ${colorIdx + 1}` : '+ Add'}:</span>
                    <span style={{ fontWeight: 600 }}>{t.name || t.id.substring(0, 8)}</span>
                    <span style={{ opacity: 0.7, fontSize: '0.7rem' }}>({t.method})</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Differential KPI Cards (when exactly 2 tests are selected) */}
          {isPair && (
            <div className="card" style={{ marginBottom: '1.25rem', background: 'linear-gradient(135deg, rgba(15, 23, 42, 0.9), rgba(30, 41, 59, 0.7))' }}>
              <div style={{ fontSize: '0.85rem', fontWeight: 700, color: '#38bdf8', marginBottom: '0.75rem', display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
                <Sparkles size={15} />
                Comparative Delta: Test 2 ({t2.name}) vs Test 1 Baseline ({t1.name})
              </div>
              <div className="summary-banner-grid">
                {/* Throughput Delta */}
                <div className="summary-item">
                  <span className="summary-label">Throughput Delta</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <span className="summary-value" style={{ color: rpsDelta >= 0 ? '#10b981' : '#f43f5e' }}>
                      {rpsDelta >= 0 ? '+' : ''}{rpsDelta.toFixed(1)} RPS
                    </span>
                    {rpsDelta >= 0 ? <ArrowUpRight size={18} color="#10b981" /> : <ArrowDownRight size={18} color="#f43f5e" />}
                  </div>
                  <span className="summary-subtext">
                    {rpsDelta >= 0 ? `+${rpsPct}% throughput gain` : `${rpsPct}% throughput reduction`}
                  </span>
                </div>

                {/* Latency Delta */}
                <div className="summary-item">
                  <span className="summary-label">Average Latency Delta</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <span className="summary-value" style={{ color: latDelta <= 0 ? '#10b981' : '#f43f5e' }}>
                      {latDelta >= 0 ? '+' : ''}{latDelta.toFixed(1)} ms
                    </span>
                    {latDelta <= 0 ? <ArrowDownRight size={18} color="#10b981" /> : <ArrowUpRight size={18} color="#f43f5e" />}
                  </div>
                  <span className="summary-subtext">
                    {latDelta <= 0 ? `${Math.abs(Number(latPct))}% faster response time` : `+${latPct}% latency increase`}
                  </span>
                </div>

                {/* P95 SLA Delta */}
                <div className="summary-item">
                  <span className="summary-label">P95 SLA Delta</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <span className="summary-value" style={{ color: p95Delta <= 0 ? '#10b981' : '#f43f5e' }}>
                      {p95Delta >= 0 ? '+' : ''}{p95Delta.toFixed(1)} ms
                    </span>
                    {p95Delta <= 0 ? <ArrowDownRight size={18} color="#10b981" /> : <ArrowUpRight size={18} color="#f43f5e" />}
                  </div>
                  <span className="summary-subtext">
                    {p95Delta <= 0 ? `SLA improved by ${Math.abs(Number(p95Pct))}%` : `SLA degraded by +${p95Pct}%`}
                  </span>
                </div>

                {/* Error Rate Delta */}
                <div className="summary-item">
                  <span className="summary-label">Error Rate Delta</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.35rem' }}>
                    <span className="summary-value" style={{ color: errDelta <= 0 ? '#10b981' : '#f43f5e' }}>
                      {errDelta >= 0 ? '+' : ''}{errDelta.toFixed(1)}%
                    </span>
                  </div>
                  <span className="summary-subtext">
                    {errDelta === 0 ? 'Identical reliability' : (errDelta < 0 ? 'Fewer failures' : 'Higher error rate')}
                  </span>
                </div>
              </div>
            </div>
          )}

          {/* Visual Charts Comparison */}
          <div className="grid-two-columns" style={{ marginBottom: '1.25rem' }}>
            {/* Chart 1: Latency Percentile Grouped Bar */}
            <div className="chart-card">
              <div className="chart-header">
                <div className="chart-title-area">
                  <Clock size={16} color="#818cf8" />
                  <h3 className="chart-title">Latency Percentiles by Test (ms)</h3>
                </div>
              </div>
              <div className="chart-wrapper">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={latencyChartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
                    <XAxis dataKey="metric" stroke="#64748b" tick={{ fontSize: 10 }} />
                    <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
                    <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '11px' }} />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    {comparedTests.map((_, idx) => (
                      <Bar
                        key={idx}
                        dataKey={`Test ${idx + 1}`}
                        fill={testColors[idx % testColors.length]}
                        radius={[4, 4, 0, 0]}
                        isAnimationActive={false}
                      />
                    ))}
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Chart 2: Throughput RPS Comparison */}
            <div className="chart-card">
              <div className="chart-header">
                <div className="chart-title-area">
                  <TrendingUp size={16} color="#0ea5e9" />
                  <h3 className="chart-title">Throughput Comparison (RPS)</h3>
                </div>
              </div>
              <div className="chart-wrapper">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={volumeChartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
                    <XAxis dataKey="name" stroke="#64748b" tick={{ fontSize: 10 }} />
                    <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
                    <Tooltip contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '11px' }} />
                    <Bar dataKey="RPS" name="Requests / Sec" radius={[4, 4, 0, 0]} isAnimationActive={false}>
                      {volumeChartData.map((_, idx) => (
                        <Cell key={idx} fill={testColors[idx % testColors.length]} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          {/* Side-by-Side Comparison Table */}
          <div className="card table-card print-container">
            <div className="card-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Layers size={18} color="#38bdf8" />
                <h3 className="card-title">Side-by-Side Benchmark Matrix</h3>
              </div>
            </div>

            <div className="table-responsive">
              <table className="data-table">
                <thead>
                  <tr>
                    <th style={{ width: '22%' }}>Benchmark Metric</th>
                    {comparedTests.map((t, idx) => (
                      <th key={t.id} style={{ color: testColors[idx % testColors.length] }}>
                        Test {idx + 1}: {t.name}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td><strong>Target Endpoint</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font" style={{ fontSize: '0.75rem' }}>
                        [{t.method}] {t.target_url}
                      </td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>Virtual Users &bull; Duration</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id}>
                        {t.virtual_users} VUs &bull; {t.duration_seconds}s
                      </td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>Throughput (RPS)</strong></td>
                    {comparedTests.map((t) => {
                      const rps = ((t.total_requests || 0) / (t.duration_seconds || 1)).toFixed(1);
                      const isWinner = t.id === highestRpsWinnerId;
                      return (
                        <td key={t.id} className="code-font" style={{ fontWeight: isWinner ? 700 : 500, color: isWinner ? '#38bdf8' : undefined }}>
                          {rps} req/s {isWinner && <span className="badge-sla badge-sla-pass" style={{ marginLeft: 4 }}>Fastest</span>}
                        </td>
                      );
                    })}
                  </tr>
                  <tr>
                    <td><strong>Total Requests</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font">
                        {(t.total_requests || 0).toLocaleString()}
                      </td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>Successful Requests</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font" style={{ color: '#10b981' }}>
                        {(t.successful_requests || 0).toLocaleString()}
                      </td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>Failed Requests &bull; Error Rate</strong></td>
                    {comparedTests.map((t) => {
                      const isWinner = t.id === bestReliabilityWinnerId && t.error_rate === 0;
                      return (
                        <td key={t.id}>
                          <span style={{ color: t.failed_requests > 0 ? '#ef4444' : '#10b981', fontWeight: 600 }}>
                            {t.failed_requests} ({t.error_rate?.toFixed(1) || '0.0'}%)
                          </span>
                          {isWinner && <span className="badge-sla badge-sla-pass" style={{ marginLeft: 6 }}>100% Stable</span>}
                        </td>
                      );
                    })}
                  </tr>
                  <tr>
                    <td><strong>Average Round-Trip Latency</strong></td>
                    {comparedTests.map((t) => {
                      const isWinner = t.id === fastestWinnerId;
                      return (
                        <td key={t.id} className="code-font" style={{ fontWeight: isWinner ? 700 : 500, color: isWinner ? '#10b981' : undefined }}>
                          {(t.avg_latency_ms || 0).toFixed(1)} ms
                          {isWinner && <span className="badge-sla badge-sla-pass" style={{ marginLeft: 6 }}>Winner</span>}
                        </td>
                      );
                    })}
                  </tr>
                  <tr>
                    <td><strong>P50 (Median Response)</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font">{(t.p50_latency_ms || 0).toFixed(2)} ms</td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>P90 SLA Latency</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font">{(t.p90_latency_ms || 0).toFixed(2)} ms</td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>P95 SLA Latency</strong></td>
                    {comparedTests.map((t) => {
                      const isWinner = t.id === lowestP95WinnerId;
                      return (
                        <td key={t.id} className="code-font" style={{ fontWeight: isWinner ? 700 : 500, color: isWinner ? '#f59e0b' : undefined }}>
                          {(t.p95_latency_ms || 0).toFixed(2)} ms
                          {isWinner && <span className="badge-sla badge-sla-warn" style={{ marginLeft: 6 }}>Best SLA</span>}
                        </td>
                      );
                    })}
                  </tr>
                  <tr>
                    <td><strong>P99 Tail Outlier Latency</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font">{(t.p99_latency_ms || 0).toFixed(2)} ms</td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>Min &bull; Max Latency</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} className="code-font" style={{ fontSize: '0.8rem' }}>
                        {(t.min_latency_ms || 0).toFixed(1)}ms &bull; {(t.max_latency_ms || 0).toFixed(1)}ms
                      </td>
                    ))}
                  </tr>
                  <tr>
                    <td><strong>Execution Date</strong></td>
                    {comparedTests.map((t) => (
                      <td key={t.id} style={{ fontSize: '0.8rem', color: 'var(--text-secondary)' }}>
                        {t.started_at ? new Date(t.started_at).toLocaleString() : 'N/A'}
                      </td>
                    ))}
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
