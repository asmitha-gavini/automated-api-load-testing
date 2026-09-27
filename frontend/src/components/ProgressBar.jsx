import React from 'react';
import { Timer, Users } from 'lucide-react';

export function ProgressBar({ isRunning, startTime, durationSeconds, activeUsers }) {
  if (!isRunning && !startTime) return null;

  const now = Date.now();
  const start = startTime ? new Date(startTime).getTime() : now;
  const elapsedSec = Math.max(0, Math.floor((now - start) / 1000));
  const progressPercent = durationSeconds > 0 ? Math.min(100, Math.round((elapsedSec / durationSeconds) * 100)) : 100;

  return (
    <div className="live-progress-container">
      <div className="progress-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-accent)' }}>
          <Timer size={14} />
          <span>
            {isRunning ? 'Load Test In Progress' : 'Test Concluded'}: {elapsedSec}s elapsed{' '}
            {durationSeconds > 0 ? `/ ${durationSeconds}s limit` : ''}
          </span>
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', color: 'var(--text-secondary)' }}>
          <Users size={14} />
          <span>{activeUsers} Active Workers</span>
          <span style={{ color: 'var(--text-accent)', fontWeight: 700 }}>({progressPercent}%)</span>
        </div>
      </div>

      <div className="progress-bar-bg">
        <div className="progress-bar-fill" style={{ width: `${progressPercent}%` }} />
      </div>
    </div>
  );
}
