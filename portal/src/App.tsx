import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import Layout from './components/Layout';
import CodeInputPage from './pages/CodeInputPage';
import DashboardPage from './pages/DashboardPage';
import DemoPage from './pages/DemoPage';
import TokenPage from './pages/TokenPage';
import SettingsPage from './pages/SettingsPage';
import MessagesPage from './pages/MessagesPage';

export default function App() {
  return (
    <BrowserRouter basename="/portal">
      <Routes>
        <Route path="/code" element={<CodeInputPage />} />
        <Route path="/total" element={<DemoPage />} />

        <Route element={<Layout />}>
          <Route path="/" element={<DashboardPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/messages" element={<MessagesPage />} />
          <Route path="/settings/token" element={<TokenPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>

        <Route path="*" element={<Navigate to="/code" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
