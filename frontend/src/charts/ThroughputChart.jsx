import React from 'react';
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  CartesianGrid,
} from 'recharts';
import { TrendingUp } from 'lucide-react';
import { CustomTooltip } from './CustomTooltip';

export function ThroughputChart({ data, currentRPS = 0 }) {
  return (
    <div className="chart-card">
      <div className="chart-header">
        <div className="chart-title-area">
          <TrendingUp size={16} color="#38bdf8" />
          <h3 className="chart-title">Throughput (Requests / Sec)</h3>
        </div>
        <div className="chart-stat-badge" style={{ color: '#38bdf8' }}>
          {Number(currentRPS).toFixed(1)} RPS
        </div>
      </div>

      <div className="chart-wrapper">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="rpsGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#38bdf8" stopOpacity={0.4} />
                <stop offset="95%" stopColor="#38bdf8" stopOpacity={0.0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
            <XAxis dataKey="time" stroke="#64748b" tick={{ fontSize: 10 }} />
            <YAxis stroke="#64748b" tick={{ fontSize: 10 }} />
            <Tooltip content={<CustomTooltip unit="req/s" />} />
            <Area
              type="monotone"
              dataKey="rps"
              name="Throughput"
              stroke="#38bdf8"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#rpsGradient)"
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
