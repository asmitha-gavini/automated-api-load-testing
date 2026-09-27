import React from 'react';
import { Gauge, Sparkles } from 'lucide-react';

export function EmptyState({ onSelectPreset }) {
  return (
    <div className="card empty-state">
      <div className="empty-state-icon">
        <Gauge size={28} />
      </div>
      <h3 className="empty-state-title">No Load Test Currently Active</h3>
      <p className="empty-state-desc">
        Configure an API endpoint on the left panel or pick a safe mock preset to begin real-time load testing.
      </p>

      {onSelectPreset && (
        <div style={{ marginTop: '1.25rem' }}>
          <button
            className="preset-chip"
            style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem', padding: '0.5rem 1rem' }}
            onClick={() => onSelectPreset('users')}
          >
            <Sparkles size={14} color="#38bdf8" />
            Load Fast Users Preset (Local Mock)
          </button>
        </div>
      )}
    </div>
  );
}
