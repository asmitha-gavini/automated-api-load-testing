import React from 'react';
import { Dashboard } from './pages/Dashboard';
import { ErrorBoundary } from './components/ErrorBoundary';
import './styles/index.css';
import './styles/dashboard.css';
import './styles/components.css';
import './styles/charts.css';

export default function App() {
  return (
    <ErrorBoundary>
      <Dashboard />
    </ErrorBoundary>
  );
}
