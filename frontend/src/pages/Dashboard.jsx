import React, { useState, useEffect, useCallback } from 'react';
import {
  LayoutDashboard,
  PlaySquare,
  History,
  FileText,
  Settings,
  AlertCircle,
  Zap,
  BarChart3,
  GitCompare,
} from 'lucide-react';
import { Header } from '../components/Header';
import { ConfigPanel } from '../components/ConfigPanel';
import { MetricCards } from '../components/MetricCards';
import { ProgressBar } from '../components/ProgressBar';
import { EmptyState } from '../components/EmptyState';
import { LatencyChart } from '../charts/LatencyChart';
import { ThroughputChart } from '../charts/ThroughputChart';
import { ErrorRateChart } from '../charts/ErrorRateChart';
import { ActiveUsersChart } from '../charts/ActiveUsersChart';
import { useWebSocket } from '../hooks/useWebSocket';
import {
  checkBackendHealth,
  startLoadTest,
  stopLoadTest,
  getTestStatus,
  getTestMetrics,
} from '../services/api';
import { HistoryView } from './HistoryView';
import { ReportsView } from './ReportsView';
import { AnalyticsView } from './AnalyticsView';
import { ComparisonView } from './ComparisonView';
import { ErrorBoundary } from '../components/ErrorBoundary';

export function Dashboard() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [comparisonIds, setComparisonIds] = useState([]);
  const [backendConnected, setBackendConnected] = useState(false);
  const [currentStatus, setCurrentStatus] = useState('idle');
  const [activeConfig, setActiveConfig] = useState(null);
  const [startTime, setStartTime] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState(null);

  // Navigate to compare view with selected IDs
  const handleNavigateToCompare = useCallback((ids) => {
    setComparisonIds(ids || []);
    setActiveTab('comparison');
  }, []);

  // Handle terminal status transition from WebSocket
  const handleTerminalState = useCallback((status) => {
    setCurrentStatus(status);
  }, []);

  const { wsStatus, latestMetrics, chartHistory, clearHistory } = useWebSocket(handleTerminalState);

  // Periodic health check
  const verifyBackend = useCallback(async () => {
    try {
      await checkBackendHealth();
      setBackendConnected(true);
    } catch {
      setBackendConnected(false);
    }
  }, []);

  // Sync state with backend on mount
  useEffect(() => {
    verifyBackend();
    const interval = setInterval(verifyBackend, 10000);

    // Fetch initial status if a test is already running
    getTestStatus()
      .then((data) => {
        if (data.status) setCurrentStatus(data.status);
        if (data.config) setActiveConfig(data.config);
      })
      .catch(() => {});

    return () => clearInterval(interval);
  }, [verifyBackend]);

  // Start test handler
  const handleStart = async (config) => {
    setIsLoading(true);
    setErrorMessage(null);
    clearHistory();

    try {
      const res = await startLoadTest(config);
      setCurrentStatus('running');
      setActiveConfig(res.test || config);
      setStartTime(Date.now());
    } catch (err) {
      setErrorMessage(err.message || 'Failed to start load test');
    } finally {
      setIsLoading(false);
    }
  };

  // Stop test handler
  const handleStop = async () => {
    setIsLoading(true);
    setErrorMessage(null);

    try {
      await stopLoadTest();
      setCurrentStatus('stopped');
    } catch (err) {
      setErrorMessage(err.message || 'Failed to stop test');
    } finally {
      setIsLoading(false);
    }
  };

  const hasMetrics = latestMetrics || chartHistory.length > 0;
  const isRunning = currentStatus === 'running';

  return (
    <div className="app-container">
      {/* Sidebar Navigation */}
      <aside className="sidebar">
        <div className="brand-section">
          <div className="brand-icon">
            <Zap size={22} fill="white" />
          </div>
          <div>
            <div className="brand-title">LoadPulse</div>
            <div className="brand-subtitle">API Benchmarker</div>
          </div>
        </div>

        <nav className="nav-menu">
          <button
            className={`nav-item ${activeTab === 'dashboard' ? 'active' : ''}`}
            onClick={() => setActiveTab('dashboard')}
          >
            <LayoutDashboard size={18} />
            <span>Dashboard</span>
          </button>

          <button
            className={`nav-item ${activeTab === 'create-test' ? 'active' : ''}`}
            onClick={() => setActiveTab('dashboard')}
          >
            <PlaySquare size={18} />
            <span>Create Test</span>
          </button>

          <button
            className={`nav-item ${activeTab === 'history' ? 'active' : ''}`}
            onClick={() => setActiveTab('history')}
          >
            <History size={18} />
            <span>Test History</span>
          </button>

          <button
            className={`nav-item ${activeTab === 'analytics' ? 'active' : ''}`}
            onClick={() => setActiveTab('analytics')}
          >
            <BarChart3 size={18} />
            <span>Analytics</span>
          </button>

          <button
            className={`nav-item ${activeTab === 'comparison' ? 'active' : ''}`}
            onClick={() => setActiveTab('comparison')}
          >
            <GitCompare size={18} />
            <span>Compare Tests</span>
            {comparisonIds.length > 0 && (
              <span className="nav-badge" style={{ backgroundColor: 'rgba(56, 189, 248, 0.2)', color: '#38bdf8' }}>
                {comparisonIds.length}
              </span>
            )}
          </button>

          <button
            className={`nav-item ${activeTab === 'reports' ? 'active' : ''}`}
            onClick={() => setActiveTab('reports')}
          >
            <FileText size={18} />
            <span>Reports</span>
          </button>

          <button
            className={`nav-item ${activeTab === 'settings' ? 'active' : ''}`}
            onClick={() => setActiveTab('settings')}
          >
            <Settings size={18} />
            <span>Settings</span>
          </button>
        </nav>

        <div className="sidebar-footer">
          <div>Engine: Go 1.27 + Gin</div>
          <div>Telemetry: WebSocket + Prom</div>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="main-wrapper">
        <Header
          backendConnected={backendConnected}
          wsStatus={wsStatus}
          currentStatus={currentStatus}
        />

        <main className="dashboard-content">
          {/* Error Banner */}
          {errorMessage && (
            <div
              className="card"
              style={{
                backgroundColor: 'rgba(239, 68, 68, 0.15)',
                borderColor: 'rgba(239, 68, 68, 0.4)',
                padding: '0.85rem 1.25rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                color: '#fca5a5',
              }}
            >
              <AlertCircle size={18} color="#f87171" />
              <span style={{ fontSize: '0.85rem', flex: 1 }}>{errorMessage}</span>
              <button
                className="btn btn-secondary"
                style={{ padding: '0.2rem 0.6rem', fontSize: '0.75rem' }}
                onClick={() => setErrorMessage(null)}
              >
                Dismiss
              </button>
            </div>
          )}

          {/* Backend Offline Warning */}
          {!backendConnected && (
            <div
              className="card"
              style={{
                backgroundColor: 'rgba(245, 158, 11, 0.12)',
                borderColor: 'rgba(245, 158, 11, 0.35)',
                padding: '0.85rem 1.25rem',
                display: 'flex',
                alignItems: 'center',
                gap: '0.75rem',
                color: '#fcd34d',
              }}
            >
              <AlertCircle size={18} color="#f59e0b" />
              <span style={{ fontSize: '0.85rem' }}>
                Backend service at <code>http://localhost:8080</code> is not responding. Ensure the Go backend server is running.
              </span>
            </div>
          )}

          {/* Dashboard View */}
          {activeTab === 'dashboard' && (
            <div className="grid-two-columns">
              {/* Left Column: Configuration Controls */}
              <ConfigPanel
                isRunning={isRunning}
                onStart={handleStart}
                onStop={handleStop}
                isLoading={isLoading}
              />

              {/* Right Column: Live Telemetry & Visualizations */}
              <div style={{ display: 'flex', flexDirection: 'column' }}>
                {/* Progress bar during run */}
                <ProgressBar
                  isRunning={isRunning}
                  startTime={startTime}
                  durationSeconds={activeConfig?.duration_seconds || 10}
                  activeUsers={latestMetrics?.active_users || 0}
                />

                {/* Metrics Cards Grid */}
                <MetricCards metrics={latestMetrics} />

                {/* Real-time Recharts */}
                {hasMetrics ? (
                  <div className="charts-grid">
                    <LatencyChart
                      data={chartHistory}
                      currentAvg={latestMetrics?.average_latency_ms || 0}
                    />
                    <ThroughputChart
                      data={chartHistory}
                      currentRPS={latestMetrics?.requests_per_second || 0}
                    />
                    <ErrorRateChart
                      data={chartHistory}
                      currentErrorRate={latestMetrics?.error_rate || 0}
                    />
                    <ActiveUsersChart
                      data={chartHistory}
                      currentVUs={latestMetrics?.active_users || 0}
                    />
                  </div>
                ) : (
                  <div style={{ marginTop: '1.5rem' }}>
                    <EmptyState />
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Phase 5 Test History View */}
          {activeTab === 'history' && (
            <ErrorBoundary>
              <HistoryView onCompare={handleNavigateToCompare} />
            </ErrorBoundary>
          )}

          {/* Phase 6 Advanced Analytics View */}
          {activeTab === 'analytics' && (
            <ErrorBoundary>
              <AnalyticsView />
            </ErrorBoundary>
          )}

          {/* Phase 6 Test Comparison View */}
          {activeTab === 'comparison' && (
            <ErrorBoundary>
              <ComparisonView
                preselectedIds={comparisonIds}
                onBackToHistory={() => setActiveTab('history')}
              />
            </ErrorBoundary>
          )}

          {/* Phase 5 Reports View */}
          {activeTab === 'reports' && (
            <ErrorBoundary>
              <ReportsView />
            </ErrorBoundary>
          )}

          {/* Settings Tab */}
          {activeTab === 'settings' && (
            <div className="card" style={{ maxWidth: 650 }}>
              <div className="card-header">
                <h2 className="card-title">Environment & Connection Settings</h2>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', fontSize: '0.85rem' }}>
                <div>
                  <strong>Backend URL:</strong> <code>http://localhost:8080</code>
                </div>
                <div>
                  <strong>WebSocket URL:</strong> <code>ws://localhost:8080/ws/metrics</code>
                </div>
                <div>
                  <strong>Prometheus Endpoint:</strong> <code>http://localhost:8080/metrics</code>
                </div>
                <div>
                  <strong>Local Mock API Target:</strong> <code>http://localhost:8081</code>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
