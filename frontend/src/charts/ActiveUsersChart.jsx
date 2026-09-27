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
import { Users } from 'lucide-react';
import { CustomTooltip } from './CustomTooltip';

export function ActiveUsersChart({ data, currentVUs = 0 }) {
  return (
    <div className="chart-card">
      <div className="chart-header">
        <div className="chart-title-area">
          <Users size={16} color="#a855f7" />
          <h3 className="chart-title">Active Virtual Users (Concurrency)</h3>
        </div>
        <div className="chart-stat-badge" style={{ color: '#a855f7' }}>
          {currentVUs} Workers
        </div>
      </div>

      <div className="chart-wrapper">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="vuGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#a855f7" stopOpacity={0.4} />
                <stop offset="95%" stopColor="#a855f7" stopOpacity={0.0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(51, 65, 85, 0.3)" />
            <XAxis dataKey="time" stroke="#64748b" tick={{ fontSize: 10 }} />
            <YAxis stroke="#64748b" tick={{ fontSize: 10 }} domain={[0, 'auto']} />
            <Tooltip content={<CustomTooltip unit="VUs" />} />
            <Area
              type="stepAfter"
              dataKey="activeUsers"
              name="Active VUs"
              stroke="#a855f7"
              strokeWidth={2}
              fillOpacity={1}
              fill="url(#vuGradient)"
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
