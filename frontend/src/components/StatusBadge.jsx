import React from 'react';

export function StatusBadge({ status }) {
  const normStatus = (status || 'idle').toLowerCase();

  const badgeClass = {
    idle: 'badge-idle',
    pending: 'badge-idle',
    running: 'badge-running',
    completed: 'badge-completed',
    stopped: 'badge-stopped',
    failed: 'badge-failed',
  }[normStatus] || 'badge-idle';

  return (
    <span className={`status-badge ${badgeClass}`}>
      <span className="status-dot" style={{ width: 6, height: 6 }} />
      {normStatus}
    </span>
  );
}
