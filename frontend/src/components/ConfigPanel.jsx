import React, { useState } from 'react';
import { Play, Square, RotateCcw, Sliders, Sparkles, ChevronDown, ChevronUp } from 'lucide-react';

const PRESETS = {
  safe: {
    name: 'Safe Demo (2 VUs, 5s, 0% Error)',
    target_url: 'http://localhost:8081/api/users?delay_ms=15',
    method: 'GET',
    virtual_users: 2,
    duration_seconds: 5,
    ramp_up_seconds: 0,
    timeout_ms: 3000,
    headers: '',
    body: '',
  },
  users: {
    name: 'Demo: Users (Fast Baseline)',
    target_url: 'http://localhost:8081/api/users?delay_ms=15',
    method: 'GET',
    virtual_users: 5,
    duration_seconds: 10,
    ramp_up_seconds: 1,
    timeout_ms: 3000,
    headers: '',
    body: '',
  },
  products: {
    name: 'Demo: Products (Jitter Latency)',
    target_url: 'http://localhost:8081/api/products?delay_ms=50',
    method: 'GET',
    virtual_users: 5,
    duration_seconds: 10,
    ramp_up_seconds: 1,
    timeout_ms: 3000,
    headers: '',
    body: '',
  },
  orders: {
    name: 'Demo: Orders (POST Transactions)',
    target_url: 'http://localhost:8081/api/orders',
    method: 'POST',
    virtual_users: 5,
    duration_seconds: 10,
    ramp_up_seconds: 1,
    timeout_ms: 3000,
    headers: '{\n  "Content-Type": "application/json"\n}',
    body: '{\n  "user_id": 1,\n  "product_id": 101,\n  "quantity": 2\n}',
  },
  errors: {
    name: 'Demo: Errors (20% Failure Rate)',
    target_url: 'http://localhost:8081/api/users?error_rate=0.2&delay_ms=15',
    method: 'GET',
    virtual_users: 5,
    duration_seconds: 10,
    ramp_up_seconds: 1,
    timeout_ms: 3000,
    headers: '',
    body: '',
  },
};

const DEFAULT_CONFIG = {
  name: 'Safe Demo Test (Users API)',
  target_url: 'http://localhost:8081/api/users?delay_ms=15',
  method: 'GET',
  virtual_users: 2,
  duration_seconds: 5,
  ramp_up_seconds: 0,
  timeout_ms: 3000,
  headers: '',
  body: '',
};

export function ConfigPanel({ isRunning, onStart, onStop, isLoading }) {
  const [config, setConfig] = useState(DEFAULT_CONFIG);
  const [errors, setErrors] = useState({});
  const [showAdvanced, setShowAdvanced] = useState(false);

  const applyPreset = (key) => {
    if (PRESETS[key]) {
      setConfig({ ...PRESETS[key] });
      setErrors({});
    }
  };

  const validate = () => {
    const errs = {};

    if (!config.target_url || !config.target_url.trim()) {
      errs.target_url = 'Target URL is required.';
    } else {
      try {
        const u = new URL(config.target_url.trim());
        if (u.protocol !== 'http:' && u.protocol !== 'https:') {
          errs.target_url = 'Only http:// and https:// URLs are permitted.';
        }
      } catch {
        errs.target_url = 'Enter a valid URL (e.g. http://localhost:8081/api/users).';
      }
    }

    if (!config.virtual_users || config.virtual_users < 1 || config.virtual_users > 500) {
      errs.virtual_users = 'Virtual users must be between 1 and 500.';
    }

    if (!config.duration_seconds || config.duration_seconds < 1 || config.duration_seconds > 600) {
      errs.duration_seconds = 'Duration must be between 1 and 600 seconds.';
    }

    if (config.ramp_up_seconds < 0 || config.ramp_up_seconds >= config.duration_seconds) {
      errs.ramp_up_seconds = 'Ramp-up must be less than total duration.';
    }

    if (config.timeout_ms < 50 || config.timeout_ms > 30000) {
      errs.timeout_ms = 'Timeout must be between 50 and 30,000 ms.';
    }

    if (config.headers && config.headers.trim()) {
      try {
        JSON.parse(config.headers);
      } catch {
        errs.headers = 'Headers must be valid JSON (or empty).';
      }
    }

    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleStart = (e) => {
    e.preventDefault();
    if (!validate()) return;

    let parsedHeaders = {};
    if (config.headers && config.headers.trim()) {
      try {
        parsedHeaders = JSON.parse(config.headers);
      } catch (err) {
        setErrors({ headers: 'Invalid headers JSON' });
        return;
      }
    }

    onStart({
      name: config.name || 'Load Test',
      target_url: config.target_url.trim(),
      method: config.method,
      virtual_users: parseInt(config.virtual_users, 10),
      duration_seconds: parseInt(config.duration_seconds, 10),
      ramp_up_seconds: parseInt(config.ramp_up_seconds, 10) || 0,
      timeout_ms: parseInt(config.timeout_ms, 10) || 3000,
      headers: parsedHeaders,
      body: config.body || '',
    });
  };

  const handleReset = () => {
    setConfig(DEFAULT_CONFIG);
    setErrors({});
  };

  return (
    <div className="card">
      <div className="card-header">
        <h2 className="card-title">
          <Sliders size={18} color="#38bdf8" />
          Test Configuration
        </h2>
      </div>

      {/* Quick Mock Presets */}
      <div className="presets-container">
        <div className="presets-title">
          <Sparkles size={12} color="#38bdf8" />
          Quick Test Presets
        </div>
        <button type="button" className="preset-chip" style={{ borderColor: 'var(--accent-primary)', color: 'var(--accent-primary)' }} onClick={() => applyPreset('safe')} disabled={isRunning}>
          Safe Baseline (2 VU, 5s)
        </button>
        <button type="button" className="preset-chip" onClick={() => applyPreset('users')} disabled={isRunning}>
          Fast Users (15ms)
        </button>
        <button type="button" className="preset-chip" onClick={() => applyPreset('products')} disabled={isRunning}>
          Products (Jitter)
        </button>
        <button type="button" className="preset-chip" onClick={() => applyPreset('orders')} disabled={isRunning}>
          Orders (POST)
        </button>
        <button type="button" className="preset-chip" onClick={() => applyPreset('errors')} disabled={isRunning}>
          Errors (20% Flake)
        </button>
      </div>

      <form onSubmit={handleStart}>
        {/* Method & URL */}
        <div className="form-group">
          <label className="form-label">
            <span>Target API Endpoint</span>
            <span style={{ fontSize: '0.7rem', color: 'var(--text-muted)' }}>HTTP / HTTPS</span>
          </label>
          <div className="input-with-select">
            <select
              className="method-select"
              value={config.method}
              onChange={(e) => setConfig({ ...config, method: e.target.value })}
              disabled={isRunning}
            >
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
              <option value="DELETE">DELETE</option>
            </select>
            <input
              type="text"
              className={`input-field ${errors.target_url ? 'error' : ''}`}
              placeholder="http://localhost:8081/api/users"
              value={config.target_url}
              onChange={(e) => setConfig({ ...config, target_url: e.target.value })}
              disabled={isRunning}
            />
          </div>
          {errors.target_url && <span className="error-text">{errors.target_url}</span>}
        </div>

        {/* Concurrency & Duration */}
        <div className="form-grid-two">
          <div className="form-group">
            <label className="form-label">Virtual Users (VUs)</label>
            <input
              type="number"
              className={`input-field ${errors.virtual_users ? 'error' : ''}`}
              min="1"
              max="500"
              value={config.virtual_users}
              onChange={(e) => setConfig({ ...config, virtual_users: e.target.value })}
              disabled={isRunning}
            />
            {errors.virtual_users && <span className="error-text">{errors.virtual_users}</span>}
          </div>

          <div className="form-group">
            <label className="form-label">Duration (seconds)</label>
            <input
              type="number"
              className={`input-field ${errors.duration_seconds ? 'error' : ''}`}
              min="1"
              max="600"
              value={config.duration_seconds}
              onChange={(e) => setConfig({ ...config, duration_seconds: e.target.value })}
              disabled={isRunning}
            />
            {errors.duration_seconds && <span className="error-text">{errors.duration_seconds}</span>}
          </div>
        </div>

        {/* Ramp-Up & Timeout */}
        <div className="form-grid-two">
          <div className="form-group">
            <label className="form-label">Ramp-Up Time (s)</label>
            <input
              type="number"
              className={`input-field ${errors.ramp_up_seconds ? 'error' : ''}`}
              min="0"
              max="60"
              value={config.ramp_up_seconds}
              onChange={(e) => setConfig({ ...config, ramp_up_seconds: e.target.value })}
              disabled={isRunning}
            />
            {errors.ramp_up_seconds && <span className="error-text">{errors.ramp_up_seconds}</span>}
          </div>

          <div className="form-group">
            <label className="form-label">Timeout (ms)</label>
            <input
              type="number"
              className={`input-field ${errors.timeout_ms ? 'error' : ''}`}
              min="50"
              max="30000"
              step="100"
              value={config.timeout_ms}
              onChange={(e) => setConfig({ ...config, timeout_ms: e.target.value })}
              disabled={isRunning}
            />
            {errors.timeout_ms && <span className="error-text">{errors.timeout_ms}</span>}
          </div>
        </div>

        {/* Collapsible Headers & Body */}
        <div style={{ marginTop: '0.5rem', marginBottom: '1rem' }}>
          <button
            type="button"
            className="btn btn-secondary"
            style={{ width: '100%', fontSize: '0.78rem', padding: '0.5rem', justifyContent: 'space-between' }}
            onClick={() => setShowAdvanced(!showAdvanced)}
          >
            <span>Advanced (Headers & Body)</span>
            {showAdvanced ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
          </button>

          {showAdvanced && (
            <div style={{ marginTop: '0.75rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
              <div className="form-group">
                <label className="form-label">
                  <span>Request Headers (JSON)</span>
                </label>
                <textarea
                  className={`input-field ${errors.headers ? 'error' : ''}`}
                  rows="3"
                  style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem' }}
                  placeholder='{"Authorization": "Bearer token", "X-Custom": "val"}'
                  value={config.headers}
                  onChange={(e) => setConfig({ ...config, headers: e.target.value })}
                  disabled={isRunning}
                />
                {errors.headers && <span className="error-text">{errors.headers}</span>}
              </div>

              {(config.method === 'POST' || config.method === 'PUT' || config.method === 'PATCH') && (
                <div className="form-group">
                  <label className="form-label">Request Body (JSON / Text)</label>
                  <textarea
                    className="input-field"
                    rows="3"
                    style={{ fontFamily: 'var(--font-mono)', fontSize: '0.75rem' }}
                    placeholder='{"key": "value"}'
                    value={config.body}
                    onChange={(e) => setConfig({ ...config, body: e.target.value })}
                    disabled={isRunning}
                  />
                </div>
              )}
            </div>
          )}
        </div>

        {/* Action Controls */}
        <div className="btn-actions">
          {!isRunning ? (
            <button type="submit" className="btn btn-primary" style={{ flex: 1 }} disabled={isLoading}>
              <Play size={16} fill="currentColor" />
              {isLoading ? 'Starting...' : 'Start Load Test'}
            </button>
          ) : (
            <button
              type="button"
              className="btn btn-danger"
              style={{ flex: 1 }}
              onClick={onStop}
              disabled={isLoading}
            >
              <Square size={16} fill="currentColor" />
              {isLoading ? 'Stopping...' : 'Stop Test'}
            </button>
          )}

          <button
            type="button"
            className="btn btn-secondary"
            onClick={handleReset}
            disabled={isRunning || isLoading}
            title="Reset to defaults"
          >
            <RotateCcw size={16} />
          </button>
        </div>
      </form>
    </div>
  );
}
