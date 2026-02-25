import React from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Plug, Building2, Users, Link as LinkIcon, MessageSquare, LogOut } from 'lucide-react';
import { cn } from '../lib/utils';
import { api } from '../lib/api';
import { Button } from './ui/button';

interface LayoutProps {
  children: React.ReactNode;
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const navigate = useNavigate();

  const handleLogout = async () => {
    try {
      await api.logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed', error);
    }
  };

  const navItems = [
    { href: '/', label: '대시보드', icon: LayoutDashboard },
    { href: '/sessions', label: '플러그인 세션', icon: Plug },
    { href: '/accounts', label: 'API 계정', icon: Building2 },
    { href: '/users', label: '포털 관리자', icon: Users },
    { href: '/mappings', label: '연결 관리', icon: LinkIcon },
    { href: '/messages', label: '메시지', icon: MessageSquare },
  ];

  return (
    <div className="flex h-screen w-full bg-background">
      <aside className="w-64 border-r bg-card">
        <div className="flex h-14 items-center border-b px-6">
          <span className="font-bold text-lg">Relay Admin</span>
        </div>
        <nav className="flex flex-col gap-2 p-4">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = location.pathname === item.href;
            return (
              <Link
                key={item.href}
                to={item.href}
                className={cn(
                  "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground",
                  isActive ? "bg-accent text-accent-foreground" : "text-muted-foreground"
                )}
              >
                <Icon className="h-4 w-4" />
                {item.label}
              </Link>
            );
          })}
        </nav>
        <div className="absolute bottom-4 left-4 right-4">
          <Button variant="outline" className="w-full justify-start gap-3" onClick={handleLogout}>
            <LogOut className="h-4 w-4" />
            Logout
          </Button>
        </div>
      </aside>

      <main className="flex-1 overflow-y-auto p-8">
        {children}
      </main>
    </div>
  );
}
