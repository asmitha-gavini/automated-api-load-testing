import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { getExportReportUrl } from '../src/services/api.js';

describe('Frontend API Service', () => {
  it('should generate valid JSON export URLs', () => {
    const url = getExportReportUrl('test-123', 'json');
    assert.ok(url.includes('/tests/test-123/export?format=json'));
  });

  it('should generate valid CSV export URLs', () => {
    const url = getExportReportUrl('test-456', 'csv');
    assert.ok(url.includes('/tests/test-456/export?format=csv'));
  });

  it('should generate valid Markdown export URLs', () => {
    const url = getExportReportUrl('test-789', 'markdown');
    assert.ok(url.includes('/tests/test-789/export?format=markdown'));
  });

  it('should support default export format as json', () => {
    const url = getExportReportUrl('test-default');
    assert.ok(url.includes('/tests/test-default/export?format=json'));
  });
});

