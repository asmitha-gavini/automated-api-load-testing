import React from 'react';
import {
  ResponsiveContainer,
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
  Legend,
} from 'recharts';
import { Clock } from 'lucide-react';
import { CustomTooltip } from './CustomTooltip';

export function LatencyChart({ data, currentAvg = 0 }) {
  return (
    <div className="chart-card">
      <div className="chart-header">
        <div className="chart-title-area">
          <Clock size={16} color="#818cf8" />
          <h3 className="chart-title">Latency Over Time (ms)</h3>
        </div>
        <div className="chart-stat-badge" style={{ color: '#818cf8' }}>
          Avg: {Number(currentAvg).toFixed(1)} ms
        </div>
      </div>

      <div className="chart-wrapper">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
            <XAxis dataKey="time" stroke="#64748b" tick={{ fontSize: 10 }} />
            <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
            <Tooltip content={<CustomTooltip unit="ms" />} />
            <Legend wrapperStyle={{ fontSize: 11, paddingTop: 6 }} />
            <Line
              type="monotone"
              dataKey="avgLatency"
              name="Avg Latency"
              stroke="#818cf8"
              strokeWidth={2}
              dot={false}
              isAnimationActive={false}
            />
            <Line
              type="monotone"
              dataKey="p95"
              name="P95 Latency"
              stroke="#f59e0b"
              strokeWidth={1.5}
              strokeDasharray="4 4"
              dot={false}
              isAnimationActive={false}
            />
            <Line
              type="monotone"
              dataKey="p99"
              name="P99 Latency"
              stroke="#ec4899"
              strokeWidth={1.5}
              strokeDasharray="2 2"
              dot={false}
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
