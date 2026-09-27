import React, { useState, useEffect } from 'react';
import {
  BarChart3,
  TrendingUp,
  Clock,
  Gauge,
  CheckCircle2,
  AlertTriangle,
  Zap,
  Users,
  Activity,
  Layers,
  Sparkles,
  PieChart as PieIcon,
  RefreshCw,
} from 'lucide-react';
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Cell,
  Legend,
} from 'recharts';
import { fetchTestHistory } from '../services/api';

export function AnalyticsView() {
  const [tests, setTests] = useState([]);
  const [selectedId, setSelectedId] = useState('');
  const [isLoading, setIsLoading] = useState(true);

  const loadData = async () => {
    setIsLoading(true);
    try {
      const data = await fetchTestHistory({ limit: 50, sortBy: 'started_at', order: 'DESC' });
      const list = data.tests || [];
      setTests(list);
      if (list.length > 0 && !selectedId) {
        setSelectedId(list[0].id);
      }
    } catch (err) {
      console.error('Failed to load analytics data:', err);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, []);

  const activeTest = tests.find((t) => t.id === selectedId) || tests[0] || null;

  const total = activeTest?.total_requests || 0;
  const success = activeTest?.successful_requests || 0;
  const failed = activeTest?.failed_requests || 0;
  const duration = activeTest?.duration_seconds || 1;
  const throughput = (total / duration).toFixed(1);
  const successRatio = total > 0 ? ((success / total) * 100).toFixed(1) : '100.0';
  const peakVUs = activeTest?.virtual_users || 0;

  // Chart data: P50 / P90 / P95 / P99 Percentiles
  const percentilesData = activeTest ? [
    { name: 'P50 (Median)', latency: activeTest.p50_latency_ms || 0, target: 50, fill: '#38bdf8' },
    { name: 'P90 (90%)', latency: activeTest.p90_latency_ms || 0, target: 100, fill: '#818cf8' },
    { name: 'P95 (SLA)', latency: activeTest.p95_latency_ms || 0, target: 200, fill: '#f59e0b' },
    { name: 'P99 (Tail)', latency: activeTest.p99_latency_ms || 0, target: 500, fill: '#ec4899' },
  ] : [];

  // Chart data: Latency Extremes (Min, Avg, Max)
  const extremesData = activeTest ? [
    { metric: 'Min Latency', ms: activeTest.min_latency_ms || 0, fill: '#10b981' },
    { metric: 'Avg Latency', ms: activeTest.avg_latency_ms || 0, fill: '#38bdf8' },
    { metric: 'Max Latency', ms: activeTest.max_latency_ms || 0, fill: '#f43f5e' },
  ] : [];

  // Chart data: Historical Throughput Trend across recent runs
  const historicalTrend = [...tests].reverse().map((t, idx) => ({
    name: `#${idx + 1}`,
    testName: t.name || t.id.substring(0, 8),
    throughput: Number(((t.total_requests || 0) / (t.duration_seconds || 1)).toFixed(1)),
    avgLatency: Number((t.avg_latency_ms || 0).toFixed(1)),
    errorRate: Number((t.error_rate || 0).toFixed(1)),
  }));

  return (
    <div className="analytics-page">
      {/* Page Header */}
      <div className="card" style={{ marginBottom: '1.25rem' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
              <BarChart3 size={20} color="#38bdf8" />
              <h2 className="card-title" style={{ fontSize: '1.25rem', marginBottom: 0 }}>
                Advanced Benchmark Analytics & SLA Telemetry
              </h2>
            </div>
            <p style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
              In-depth percentile distributions, throughput trends, and latency extremes.
            </p>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            {tests.length > 0 && (
              <select
                className="input-field"
                style={{ width: 'auto', minWidth: '240px' }}
                value={selectedId}
                onChange={(e) => setSelectedId(e.target.value)}
              >
                {tests.map((t) => (
                  <option key={t.id} value={t.id}>
                    {t.name} ({t.method} &bull; {new Date(t.started_at).toLocaleDateString()})
                  </option>
                ))}
              </select>
            )}
            <button className="btn-icon" onClick={loadData} title="Refresh Analytics">
              <RefreshCw size={16} />
            </button>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="card loading-state">Computing analytics models...</div>
      ) : !activeTest ? (
        <div className="card empty-state">
          <Activity size={36} color="#64748b" style={{ marginBottom: '0.75rem' }} />
          <h3>No Benchmark Data Available</h3>
          <p>Run a load test from the Dashboard to populate advanced analytics.</p>
        </div>
      ) : (
        <div>
          {/* Advanced Analytics KPI Grid */}
          <div className="metrics-grid" style={{ marginBottom: '1.25rem' }}>
            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(56, 189, 248, 0.12)' }}>
                <TrendingUp size={20} color="#38bdf8" />
              </div>
              <div className="metric-value">{throughput} RPS</div>
              <div className="metric-label">Peak Throughput</div>
              <div className="metric-subtext">{total.toLocaleString()} total requests</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(129, 140, 248, 0.15)' }}>
                <Clock size={20} color="#818cf8" />
              </div>
              <div className="metric-value">{(activeTest.avg_latency_ms || 0).toFixed(1)} ms</div>
              <div className="metric-label">Mean Round-Trip</div>
              <div className="metric-subtext">P95: {(activeTest.p95_latency_ms || 0).toFixed(1)}ms</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(168, 85, 247, 0.15)' }}>
                <Users size={20} color="#a855f7" />
              </div>
              <div className="metric-value">{peakVUs} VUs</div>
              <div className="metric-label">Peak Concurrency</div>
              <div className="metric-subtext">goroutines active</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(16, 185, 129, 0.12)' }}>
                <CheckCircle2 size={20} color="#10b981" />
              </div>
              <div className="metric-value" style={{ color: '#10b981' }}>{successRatio}%</div>
              <div className="metric-label">Success Ratio</div>
              <div className="metric-subtext">{success.toLocaleString()} successful calls</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: failed > 0 ? 'rgba(239, 68, 68, 0.15)' : 'rgba(148, 163, 184, 0.1)' }}>
                <AlertTriangle size={20} color={failed > 0 ? '#ef4444' : '#94a3b8'} />
              </div>
              <div className="metric-value" style={{ color: failed > 0 ? '#ef4444' : 'inherit' }}>
                {(activeTest.error_rate || 0).toFixed(1)}%
              </div>
              <div className="metric-label">Failure Rate</div>
              <div className="metric-subtext">{failed.toLocaleString()} failures</div>
            </div>

            <div className="metric-card">
              <div className="metric-icon-wrapper" style={{ background: 'rgba(245, 158, 11, 0.15)' }}>
                <Activity size={20} color="#f59e0b" />
              </div>
              <div className="metric-value">{duration}s</div>
              <div className="metric-label">Test Duration</div>
              <div className="metric-subtext">Sustained benchmark</div>
            </div>
          </div>

          {/* Charts Row 1: Percentile Distribution & Extremes */}
          <div className="grid-two-columns" style={{ marginBottom: '1.25rem' }}>
            {/* Percentile Distribution Bar Chart */}
            <div className="chart-card">
              <div className="chart-header">
                <div className="chart-title-area">
                  <Gauge size={16} color="#f59e0b" />
                  <h3 className="chart-title">Latency Percentiles vs SLA Thresholds (ms)</h3>
                </div>
              </div>
              <div className="chart-wrapper">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={percentilesData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
                    <XAxis dataKey="name" stroke="#64748b" tick={{ fontSize: 10 }} />
                    <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
                    <Tooltip
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '11px' }}
                      formatter={(val) => [`${val.toFixed(2)} ms`, 'Measured Latency']}
                    />
                    <Bar dataKey="latency" radius={[4, 4, 0, 0]} isAnimationActive={false}>
                      {percentilesData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={entry.fill} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Latency Extremes (Min vs Avg vs Max) */}
            <div className="chart-card">
              <div className="chart-header">
                <div className="chart-title-area">
                  <Clock size={16} color="#38bdf8" />
                  <h3 className="chart-title">Response Time Spread (Min / Avg / Max)</h3>
                </div>
              </div>
              <div className="chart-wrapper">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={extremesData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
                    <XAxis dataKey="metric" stroke="#64748b" tick={{ fontSize: 10 }} />
                    <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
                    <Tooltip
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '11px' }}
                      formatter={(val) => [`${val.toFixed(2)} ms`, 'Latency']}
                    />
                    <Bar dataKey="ms" radius={[4, 4, 0, 0]} isAnimationActive={false}>
                      {extremesData.map((entry, index) => (
                        <Cell key={`cell-ext-${index}`} fill={entry.fill} />
                      ))}
                    </Bar>
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          </div>

          {/* Charts Row 2: Multi-Run Historical Trend */}
          {historicalTrend.length > 1 && (
            <div className="chart-card" style={{ marginBottom: '1.25rem' }}>
              <div className="chart-header">
                <div className="chart-title-area">
                  <TrendingUp size={16} color="#38bdf8" />
                  <h3 className="chart-title">Historical Throughput Trend Across Benchmark Runs (RPS)</h3>
                </div>
              </div>
              <div className="chart-wrapper" style={{ height: 260 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={historicalTrend} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
                    <XAxis dataKey="name" stroke="#64748b" tick={{ fontSize: 10 }} />
                    <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
                    <Tooltip
                      contentStyle={{ backgroundColor: '#0f172a', borderColor: '#334155', borderRadius: '8px', fontSize: '11px' }}
                      labelFormatter={(label, payload) => payload?.[0]?.payload?.testName || label}
                    />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    <Line
                      type="monotone"
                      dataKey="throughput"
                      name="Throughput (RPS)"
                      stroke="#38bdf8"
                      strokeWidth={2}
                      dot={{ r: 3 }}
                      isAnimationActive={false}
                    />
                    <Line
                      type="monotone"
                      dataKey="avgLatency"
                      name="Avg Latency (ms)"
                      stroke="#818cf8"
                      strokeWidth={1.5}
                      strokeDasharray="4 4"
                      dot={{ r: 3 }}
                      isAnimationActive={false}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {/* SLA Assessment & Status Code Distribution Card */}
          <div className="card">
            <div className="card-header">
              <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <Sparkles size={18} color="#38bdf8" />
                <h3 className="card-title">SLA Compliance & Outcome Distribution</h3>
              </div>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))', gap: '1.5rem' }}>
              {/* Success vs Error ratio visual progress */}
              <div>
                <h4 style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
                  Success vs Failure Ratio
                </h4>
                <div style={{ height: '14px', background: 'rgba(239, 68, 68, 0.25)', borderRadius: '9999px', overflow: 'hidden', display: 'flex' }}>
                  <div
                    style={{
                      width: `${successRatio}%`,
                      background: 'linear-gradient(90deg, #10b981, #34d399)',
                      height: '100%',
                      transition: 'width 0.4s ease',
                    }}
                  />
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.75rem', marginTop: '0.35rem', color: 'var(--text-muted)' }}>
                  <span style={{ color: '#10b981' }}>{successRatio}% Success ({success.toLocaleString()})</span>
                  <span style={{ color: failed > 0 ? '#ef4444' : 'inherit' }}>{activeTest.error_rate?.toFixed(1) || '0.0'}% Failed ({failed.toLocaleString()})</span>
                </div>
              </div>

              {/* SLA Tiers Compliance */}
              <div>
                <h4 style={{ fontSize: '0.85rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
                  SLA Compliance Summary
                </h4>
                <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', display: 'flex', flexDirection: 'column', gap: '0.35rem' }}>
                  <div>
                    P95 Target (&lt;200ms):{' '}
                    {(activeTest.p95_latency_ms || 0) < 200 ? (
                      <span className="badge-sla badge-sla-pass">Compliant ({(activeTest.p95_latency_ms || 0).toFixed(1)}ms)</span>
                    ) : (
                      <span className="badge-sla badge-sla-fail">Breached ({(activeTest.p95_latency_ms || 0).toFixed(1)}ms)</span>
                    )}
                  </div>
                  <div>
                    P99 Target (&lt;500ms):{' '}
                    {(activeTest.p99_latency_ms || 0) < 500 ? (
                      <span className="badge-sla badge-sla-pass">Compliant ({(activeTest.p99_latency_ms || 0).toFixed(1)}ms)</span>
                    ) : (
                      <span className="badge-sla badge-sla-fail">Breached ({(activeTest.p99_latency_ms || 0).toFixed(1)}ms)</span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
