import { useState, useEffect, useRef, useCallback } from 'react';

const MAX_HISTORY_POINTS = 60;

export function useWebSocket(onTestTerminalState) {
  const [wsStatus, setWsStatus] = useState('connecting'); // 'connecting' | 'connected' | 'disconnected' | 'reconnecting'
  const [latestMetrics, setLatestMetrics] = useState(null);
  const [chartHistory, setChartHistory] = useState([]);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const isUnmounted = useRef(false);

  const clearHistory = useCallback(() => {
    setChartHistory([]);
    setLatestMetrics(null);
  }, []);

  const connect = useCallback(() => {
    if (isUnmounted.current) return;

    // Use direct port 8080 in dev or window.location if proxied
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Fallback to localhost:8080 if running on Vite :5173
    const host = window.location.port === '5173' ? 'localhost:8080' : window.location.host;
    const wsURL = `${protocol}//${host}/ws/metrics`;

    try {
      setWsStatus((prev) => (prev === 'connected' ? 'reconnecting' : 'connecting'));
      const ws = new WebSocket(wsURL);
      wsRef.current = ws;

      ws.onopen = () => {
        if (isUnmounted.current) return;
        setWsStatus('connected');
        console.log('[WebSocket] Connected to metrics stream');
      };

      ws.onmessage = (event) => {
        if (isUnmounted.current) return;
        try {
          const msg = JSON.parse(event.data);
          setLatestMetrics(msg);

          const timeLabel = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
          const newPoint = {
            time: timeLabel,
            rps: msg.requests_per_second || 0,
            avgLatency: msg.average_latency_ms || 0,
            p50: msg.p50_latency_ms || 0,
            p95: msg.p95_latency_ms || 0,
            p99: msg.p99_latency_ms || 0,
            errorRate: msg.error_rate || 0,
            activeUsers: msg.active_users || 0,
            totalRequests: msg.total_requests || 0,
            successfulRequests: msg.successful_requests || 0,
            failedRequests: msg.failed_requests || 0,
          };

          setChartHistory((prev) => {
            const next = [...prev, newPoint];
            return next.length > MAX_HISTORY_POINTS ? next.slice(-MAX_HISTORY_POINTS) : next;
          });

          // Notify terminal state
          if ((msg.status === 'completed' || msg.status === 'stopped' || msg.status === 'failed') && onTestTerminalState) {
            onTestTerminalState(msg.status);
          }
        } catch (err) {
          console.warn('[WebSocket] Malformed JSON frame received:', err);
        }
      };

      ws.onerror = () => {
        if (isUnmounted.current) return;
        setWsStatus('disconnected');
      };

      ws.onclose = () => {
        if (isUnmounted.current) return;
        setWsStatus('disconnected');
        // Auto-reconnect after 2 seconds
        reconnectTimeoutRef.current = setTimeout(() => {
          connect();
        }, 2000);
      };
    } catch (err) {
      console.error('[WebSocket] Setup exception:', err);
      setWsStatus('disconnected');
      reconnectTimeoutRef.current = setTimeout(connect, 3000);
    }
  }, [onTestTerminalState]);

  useEffect(() => {
    isUnmounted.current = false;
    connect();

    return () => {
      isUnmounted.current = true;
      if (reconnectTimeoutRef.current) clearTimeout(reconnectTimeoutRef.current);
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [connect]);

  return {
    wsStatus,
    latestMetrics,
    chartHistory,
    clearHistory,
  };
}
