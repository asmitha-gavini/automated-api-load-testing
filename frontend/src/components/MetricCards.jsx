import React from 'react';
import {
  Hash,
  CheckCircle2,
  AlertTriangle,
  Percent,
  Clock,
  Gauge,
  Zap,
  TrendingUp,
  Users,
} from 'lucide-react';

export function MetricCards({ metrics }) {
  const m = metrics || {};
  const lat = m.latency || {};

  const total = m.total_requests ?? 0;
  const success = m.successful_requests ?? 0;
  const failed = m.failed_requests ?? 0;
  const errorRate = m.error_rate !== undefined ? m.error_rate.toFixed(1) : '0.0';
  const rps = m.requests_per_second !== undefined ? m.requests_per_second.toFixed(1) : (m.current_rps !== undefined ? m.current_rps.toFixed(1) : '0.0');
  const avgLat = (m.average_latency_ms ?? lat.avg_ms ?? 0).toFixed(1);
  const p95Lat = (m.p95_latency_ms ?? lat.p95_ms ?? 0).toFixed(1);
  const p99Lat = (m.p99_latency_ms ?? lat.p99_ms ?? 0).toFixed(1);
  const activeVUs = m.active_users ?? m.active_vus ?? 0;

  const cards = [
    {
      label: 'Total Requests',
      value: total.toLocaleString(),
      subtext: 'Cumulative dispatches',
      icon: Hash,
      color: '#38bdf8',
      bg: 'rgba(56, 189, 248, 0.12)',
    },
    {
      label: 'Successful (2xx/3xx)',
      value: success.toLocaleString(),
      subtext: `${total > 0 ? ((success / total) * 100).toFixed(1) : 100}% of total`,
      icon: CheckCircle2,
      color: '#10b981',
      bg: 'rgba(16, 185, 129, 0.12)',
    },
    {
      label: 'Failed Requests',
      value: failed.toLocaleString(),
      subtext: failed > 0 ? 'Network or 4xx/5xx' : 'Zero errors detected',
      icon: AlertTriangle,
      color: failed > 0 ? '#f43f5e' : '#94a3b8',
      bg: failed > 0 ? 'rgba(244, 63, 94, 0.15)' : 'rgba(148, 163, 184, 0.1)',
    },
    {
      label: 'Error Rate',
      value: `${errorRate}%`,
      subtext: Number(errorRate) > 5 ? 'High failure rate' : 'Normal threshold',
      icon: Percent,
      color: Number(errorRate) > 5 ? '#f43f5e' : '#10b981',
      bg: Number(errorRate) > 5 ? 'rgba(244, 63, 94, 0.15)' : 'rgba(16, 185, 129, 0.12)',
    },
    {
      label: 'Throughput (RPS)',
      value: rps,
      subtext: 'Current requests / sec',
      icon: TrendingUp,
      color: '#0ea5e9',
      bg: 'rgba(14, 165, 233, 0.15)',
    },
    {
      label: 'Average Latency',
      value: `${avgLat} ms`,
      subtext: 'Mean round-trip time',
      icon: Clock,
      color: '#818cf8',
      bg: 'rgba(129, 140, 248, 0.15)',
    },
    {
      label: 'P95 Latency',
      value: `${p95Lat} ms`,
      subtext: '95th percentile SLA',
      icon: Gauge,
      color: '#f59e0b',
      bg: 'rgba(245, 158, 11, 0.15)',
    },
    {
      label: 'P99 Latency',
      value: `${p99Lat} ms`,
      subtext: 'Tail latency outlier',
      icon: Zap,
      color: '#ec4899',
      bg: 'rgba(236, 72, 153, 0.15)',
    },
    {
      label: 'Active Users',
      value: activeVUs,
      subtext: 'Concurrent goroutines',
      icon: Users,
      color: '#a855f7',
      bg: 'rgba(168, 85, 247, 0.15)',
    },
  ];

  return (
    <div className="metrics-grid">
      {cards.map((card, idx) => {
        const IconComponent = card.icon;
        return (
          <div key={idx} className="metric-card">
            <div className="metric-top">
              <span className="metric-label">{card.label}</span>
              <div className="metric-icon-box" style={{ background: card.bg, color: card.color }}>
                <IconComponent size={16} />
              </div>
            </div>
            <div className="metric-value">{card.value}</div>
            <div className="metric-subtext">{card.subtext}</div>
          </div>
        );
      })}
    </div>
  );
}
