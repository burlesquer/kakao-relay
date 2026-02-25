import React, { useEffect, useState } from 'react';
import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, MessageSquare, Settings, Key, LogOut, Shield } from 'lucide-react';
import { Button } from './ui/button';
import { Badge } from './ui/badge';
import { api, type User } from '../lib/api';

const navItems = [
  { to: '/', icon: LayoutDashboard, label: '대시보드' },
  { to: '/messages', icon: MessageSquare, label: '메시지' },
  { to: '/settings/token', icon: Key, label: 'API 토큰', requiresAuth: true },
  { to: '/settings', icon: Settings, label: '설정', end: true, requiresAuth: true },
];

interface NavItemProps {
  to: string;
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  end?: boolean;
  variant: 'desktop' | 'mobile';
  requiresAuth?: boolean;
  isCodeSession?: boolean;
}

function NavItem({ to, icon: Icon, label, end, variant, requiresAuth, isCodeSession }: NavItemProps): React.ReactElement {
  // Hide auth-required items in code session
  if (requiresAuth && isCodeSession) {
    return <></>;
  }
  const isMobile = variant === 'mobile';
  const baseStyles = isMobile
    ? 'flex flex-1 items-center justify-center gap-2 rounded-md px-2 py-2 text-xs font-medium transition-colors'
    : 'flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium transition-colors';

  return (
    <NavLink
      to={to}
      end={end}
      className={({ isActive }) =>
        `${baseStyles} ${
          isActive
            ? 'bg-accent text-accent-foreground'
            : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
        }`
      }
    >
      <Icon className="h-4 w-4" />
      {isMobile ? <span className="hidden xs:inline">{label}</span> : label}
    </NavLink>
  );
}

export default function Layout(): React.ReactElement {
  const navigate = useNavigate();
  const [user, setUser] = useState<User | null>(null);
  const [isCodeSession, setIsCodeSession] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Try OAuth session first
    api.me()
      .then((data) => {
        if (data) {
          setUser(data);
          setIsCodeSession(false);
        } else {
          // Try code session
          return api.getCodeStats().then(() => {
            setIsCodeSession(true);
            setUser({ id: 'code-user', email: '코드 세션', accountId: 'code', createdAt: '' });
          });
        }
      })
      .catch(() => {
        // No valid session, redirect to code login
        navigate('/code');
      })
      .finally(() => setLoading(false));
  }, [navigate]);

  const handleLogout = async () => {
    try {
      await api.logout();
    } finally {
      navigate('/code');
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-muted-foreground">로딩 중...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="mx-auto flex h-14 max-w-6xl items-center justify-between px-4">
          {/* Logo */}
          <div className="flex items-center gap-6">
            <NavLink to="/" className="flex items-center gap-2 font-semibold">
              <span className="text-lg">Talk Channel x OpenClaw</span>
            </NavLink>

            {/* Navigation */}
            <nav className="hidden items-center gap-1 md:flex">
              {navItems.map((item) => (
                <NavItem key={item.to} {...item} variant="desktop" isCodeSession={isCodeSession} />
              ))}
            </nav>
          </div>

          {/* User Menu */}
          <div className="flex items-center gap-4">
            {isCodeSession ? (
              <Badge variant="secondary" className="hidden sm:flex gap-1">
                <Shield className="h-3 w-3" />
                읽기 전용
              </Badge>
            ) : (
              <span className="hidden text-sm text-muted-foreground sm:inline">
                {user?.email}
              </span>
            )}
            <Button variant="ghost" size="sm" onClick={handleLogout}>
              <LogOut className="mr-2 h-4 w-4" />
              {isCodeSession ? '종료' : '로그아웃'}
            </Button>
          </div>
        </div>

        {/* Mobile Navigation */}
        <nav className="flex items-center gap-1 border-t px-4 py-2 md:hidden">
          {navItems.map((item) => (
            <NavItem key={item.to} {...item} variant="mobile" isCodeSession={isCodeSession} />
          ))}
        </nav>
      </header>

      {/* Main Content */}
      <main className="mx-auto max-w-6xl p-4 md:p-8">
        <Outlet context={{ user, isCodeSession }} />
      </main>
    </div>
  );
}
