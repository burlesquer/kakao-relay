import { useState, useEffect } from 'react';
import { useNavigate, useOutletContext } from 'react-router-dom';
import { AlertTriangle, Link2, Trash2, Unlink } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Input } from '../components/ui/input';
import { api, type User, type OAuthProvider } from '../lib/api';

interface LayoutContext {
  user: User | null;
}

export default function SettingsPage() {
  const navigate = useNavigate();
  const { user } = useOutletContext<LayoutContext>();

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">설정</h1>
        <p className="text-muted-foreground">
          계정 설정을 관리합니다.
        </p>
      </div>

      {/* Account Info */}
      <Card>
        <CardHeader>
          <CardTitle>계정 정보</CardTitle>
          <CardDescription>현재 로그인된 계정 정보</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            <div className="flex justify-between">
              <span className="text-sm text-muted-foreground">이메일</span>
              <span className="text-sm font-medium">{user?.email}</span>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Linked Accounts */}
      <LinkedAccountsCard />

      {/* Account Deletion */}
      <AccountDeletionCard onDeleted={() => navigate('/auth')} />
    </div>
  );
}

function AccountDeletionCard({ onDeleted }: { onDeleted: () => void }) {
  const [confirmText, setConfirmText] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);

  const handleDelete = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (confirmText !== '계정 삭제') {
      setError('"계정 삭제"를 정확히 입력해주세요.');
      return;
    }

    setLoading(true);
    try {
      await api.deleteAccount();
      onDeleted();
    } catch (err) {
      setError(err instanceof Error ? err.message : '계정 삭제에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Card className="border-destructive/50">
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-destructive">
          <Trash2 className="h-5 w-5" />
          계정 삭제
        </CardTitle>
        <CardDescription>
          계정을 삭제하면 모든 데이터가 영구적으로 삭제되며 복구할 수 없습니다.
        </CardDescription>
      </CardHeader>
      <CardContent>
        {!showForm ? (
          <Button
            variant="destructive"
            onClick={() => setShowForm(true)}
          >
            계정 삭제 진행
          </Button>
        ) : (
          <form onSubmit={handleDelete} className="space-y-4">
            <div className="flex items-start gap-2 rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
              <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0" />
              <div>
                <p className="font-medium">주의: 이 작업은 되돌릴 수 없습니다!</p>
                <p className="mt-1">
                  계정 삭제 시 모든 연결, 메시지 기록, API 토큰이 영구적으로 삭제됩니다.
                </p>
              </div>
            </div>

            {error && (
              <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="confirmText" className="text-sm font-medium">
                확인을 위해 <span className="font-bold">"계정 삭제"</span>를 입력하세요
              </label>
              <Input
                id="confirmText"
                type="text"
                value={confirmText}
                onChange={(e) => setConfirmText(e.target.value)}
                placeholder="계정 삭제"
                required
              />
            </div>

            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => {
                  setShowForm(false);
                  setConfirmText('');
                  setError(null);
                }}
              >
                취소
              </Button>
              <Button
                type="submit"
                variant="destructive"
                disabled={loading || confirmText !== '계정 삭제'}
              >
                {loading ? '삭제 중...' : '계정 영구 삭제'}
              </Button>
            </div>
          </form>
        )}
      </CardContent>
    </Card>
  );
}

const providerNames: Record<string, string> = {
  google: 'Google',
  twitter: 'X (Twitter)',
};

const ProviderIcon = ({ provider }: { provider: string }) => {
  if (provider === 'google') {
    return (
      <svg className="h-5 w-5" viewBox="0 0 24 24">
        <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
        <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
        <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
        <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
      </svg>
    );
  }
  if (provider === 'twitter') {
    return (
      <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
        <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
      </svg>
    );
  }
  return null;
};

function LinkedAccountsCard() {
  const [providers, setProviders] = useState<OAuthProvider[]>([]);
  const [loading, setLoading] = useState(true);
  const [unlinkingProvider, setUnlinkingProvider] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadProviders();
  }, []);

  const loadProviders = async () => {
    try {
      const data = await api.getLinkedProviders();
      setProviders(data.providers);
    } catch (err) {
      setError(err instanceof Error ? err.message : '연결된 계정을 불러오는데 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleUnlink = async (provider: string) => {
    setError(null);
    setUnlinkingProvider(provider);
    try {
      await api.unlinkProvider(provider);
      setProviders(prev => prev.filter(p => p.provider !== provider));
    } catch (err) {
      setError(err instanceof Error ? err.message : '연결 해제에 실패했습니다.');
    } finally {
      setUnlinkingProvider(null);
    }
  };

  const canUnlink = (provider: string) => {
    const otherProviders = providers.filter(p => p.provider !== provider);
    return otherProviders.length > 0;
  };

  if (loading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Link2 className="h-5 w-5" />
            연결된 계정
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-sm text-muted-foreground">불러오는 중...</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Link2 className="h-5 w-5" />
          연결된 계정
        </CardTitle>
        <CardDescription>
          소셜 로그인 연결을 관리합니다.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive">
            {error}
          </div>
        )}

        {providers.length === 0 ? (
          <div className="text-sm text-muted-foreground">
            연결된 소셜 계정이 없습니다.
          </div>
        ) : (
          <div className="space-y-3">
            {providers.map((provider) => (
              <div
                key={provider.provider}
                className="flex items-center justify-between rounded-lg border p-3"
              >
                <div className="flex items-center gap-3">
                  <ProviderIcon provider={provider.provider} />
                  <div>
                    <div className="font-medium">{providerNames[provider.provider] || provider.provider}</div>
                    {provider.email && (
                      <div className="text-sm text-muted-foreground">{provider.email}</div>
                    )}
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleUnlink(provider.provider)}
                  disabled={!canUnlink(provider.provider) || unlinkingProvider === provider.provider}
                  title={!canUnlink(provider.provider) ? '마지막 인증 방법은 해제할 수 없습니다' : undefined}
                >
                  {unlinkingProvider === provider.provider ? (
                    '해제 중...'
                  ) : (
                    <>
                      <Unlink className="h-4 w-4 mr-1" />
                      연결 해제
                    </>
                  )}
                </Button>
              </div>
            ))}
          </div>
        )}

        {/* Link new provider */}
        <div className="pt-4 border-t">
          <div className="text-sm font-medium mb-3">계정 연결하기</div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => window.location.href = '/portal/api/oauth/google'}
              disabled={providers.some(p => p.provider === 'google')}
            >
              <svg className="h-4 w-4 mr-1" viewBox="0 0 24 24">
                <path fill="currentColor" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="currentColor" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="currentColor" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="currentColor" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
              Google
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => window.location.href = '/portal/api/oauth/twitter'}
              disabled={providers.some(p => p.provider === 'twitter')}
            >
              <svg className="h-4 w-4 mr-1" viewBox="0 0 24 24" fill="currentColor">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
              X (Twitter)
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
