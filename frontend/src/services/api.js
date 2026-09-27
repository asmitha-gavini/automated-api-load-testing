// API service layer interacting with the Go backend REST endpoints

const API_BASE = typeof window !== 'undefined' && window.location.port === '5173'
  ? 'http://localhost:8080/api/v1'
  : '/api/v1';

export async function checkBackendHealth() {
  try {
    const res = await fetch(`${API_BASE}/health`, {
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) {
      throw new Error(`HTTP ${res.status}`);
    }
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Backend unreachable');
  }
}

export async function startLoadTest(config) {
  try {
    const res = await fetch(`${API_BASE}/tests`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      body: JSON.stringify(config),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || data.details || `Failed to start test (${res.status})`);
    }
    return data;
  } catch (err) {
    throw new Error(err.message || 'Failed to start load test');
  }
}

export async function stopLoadTest() {
  try {
    const res = await fetch(`${API_BASE}/tests/stop`, {
      method: 'POST',
      headers: { 'Accept': 'application/json' },
    });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Failed to stop load test');
    }
    return data;
  } catch (err) {
    throw new Error(err.message || 'Failed to stop load test');
  }
}

export async function getTestStatus() {
  try {
    const res = await fetch(`${API_BASE}/tests/status`, {
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to retrieve test status');
  }
}

export async function getTestMetrics() {
  try {
    const res = await fetch(`${API_BASE}/tests/metrics`, {
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to retrieve metrics');
  }
}

export async function fetchTestHistory(params = {}) {
  try {
    const query = new URLSearchParams();
    if (params.search) query.set('q', params.search);
    if (params.status && params.status !== 'all') query.set('status', params.status);
    if (params.sortBy) query.set('sort_by', params.sortBy);
    if (params.order) query.set('order', params.order);
    if (params.limit) query.set('limit', params.limit);
    if (params.offset) query.set('offset', params.offset);

    const res = await fetch(`${API_BASE}/tests?${query.toString()}`, {
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to fetch test history');
  }
}

export async function fetchTestById(id) {
  try {
    const res = await fetch(`${API_BASE}/tests/${id}`, {
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to retrieve test details');
  }
}

export async function deleteTest(id) {
  try {
    const res = await fetch(`${API_BASE}/tests/${id}`, {
      method: 'DELETE',
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to delete test run');
  }
}

export async function clearAllTests() {
  try {
    const res = await fetch(`${API_BASE}/tests`, {
      method: 'DELETE',
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to clear test history');
  }
}

export async function fetchPerformanceSummary() {
  try {
    const res = await fetch(`${API_BASE}/tests/summary`, {
      headers: { 'Accept': 'application/json' },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } catch (err) {
    throw new Error(err.message || 'Failed to fetch performance summary');
  }
}

export function getExportReportUrl(id, format = 'json') {
  return `${API_BASE}/tests/${id}/export?format=${format}`;
}

