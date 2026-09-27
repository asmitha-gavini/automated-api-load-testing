import React from 'react';

export function CustomTooltip({ active, payload, label, unit = '' }) {
  if (!active || !payload || !payload.length) return null;

  return (
    <div className="custom-tooltip">
      <div className="tooltip-time">{label}</div>
      {payload.map((entry, index) => (
        <div key={index} className="tooltip-item">
          <span className="tooltip-item-label">
            <span
              style={{
                display: 'inline-block',
                width: 8,
                height: 8,
                borderRadius: '50%',
                backgroundColor: entry.color,
              }}
            />
            {entry.name}:
          </span>
          <span className="tooltip-item-value">
            {typeof entry.value === 'number' ? entry.value.toFixed(1) : entry.value} {unit}
          </span>
        </div>
      ))}
    </div>
  );
}
