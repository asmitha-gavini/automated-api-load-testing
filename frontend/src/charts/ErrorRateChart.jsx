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
import { AlertTriangle } from 'lucide-react';
import { CustomTooltip } from './CustomTooltip';

export function ErrorRateChart({ data, currentErrorRate = 0 }) {
  return (
    <div className="chart-card">
      <div className="chart-header">
        <div className="chart-title-area">
          <AlertTriangle size={16} color="#f43f5e" />
          <h3 className="chart-title">Error Rate (%)</h3>
        </div>
        <div
          className="chart-stat-badge"
          style={{ color: Number(currentErrorRate) > 5 ? '#f43f5e' : '#10b981' }}
        >
          {Number(currentErrorRate).toFixed(1)}%
        </div>
      </div>

      <div className="chart-wrapper">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="errorGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#f43f5e" stopOpacity={0.4} />
                <stop offset="95%" stopColor="#f43f5e" stopOpacity={0.0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
            <XAxis dataKey="time" stroke="#64748b" tick={{ fontSize: 10 }} />
            <YAxis stroke="#64748b" tick={{ fontSize: 10 }} domain={[0, 'auto']} />
            <Tooltip content={<CustomTooltip unit="%" />} />
            <Area
              type="monotone"
              dataKey="errorRate"
              name="Error Rate"
              stroke="#f43f5e"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#errorGradient)"
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
