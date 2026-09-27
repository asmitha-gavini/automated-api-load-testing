import React from 'react';
import { Activity, Radio, Server } from 'lucide-react';
import { StatusBadge } from './StatusBadge';

export function Header({ backendConnected, wsStatus, currentStatus }) {
  return (
    <header className="topbar">
      <div className="topbar-left">
        <h1 className="page-title">Real-Time Performance Dashboard</h1>
      </div>

      <div className="topbar-right">
        {/* Backend API Status Pill */}
        <div className="status-pill" title="Go Backend REST API connectivity">
          <Server size={14} color={backendConnected ? '#10b981' : '#ef4444'} />
          <span style={{ color: 'var(--text-secondary)' }}>API:</span>
          <span className={`status-dot ${backendConnected ? 'online' : 'offline'}`} />
          <span style={{ color: backendConnected ? '#10b981' : '#ef4444' }}>
            {backendConnected ? 'Ready' : 'Offline'}
          </span>
        </div>

        {/* WebSocket Stream Status Pill */}
        <div className="status-pill" title="Gorilla WebSocket live telemetry connection">
          <Radio size={14} color={wsStatus === 'connected' ? '#38bdf8' : '#f59e0b'} />
          <span style={{ color: 'var(--text-secondary)' }}>Stream:</span>
          <span
            className={`status-dot ${
              wsStatus === 'connected' ? 'online' : wsStatus === 'connecting' || wsStatus === 'reconnecting' ? 'connecting' : 'offline'
            }`}
          />
          <span style={{ textTransform: 'capitalize', color: wsStatus === 'connected' ? '#38bdf8' : '#f59e0b' }}>
            {wsStatus}
          </span>
        </div>

        {/* Test Lifecycle Status */}
        <StatusBadge status={currentStatus} />
      </div>
    </header>
  );
}
