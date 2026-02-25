import React, { useEffect, useState } from 'react';
import { Copy, RefreshCw, Eye, EyeOff, AlertTriangle } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { api, type TokenResponse } from '../lib/api';

export default function TokenPage() {
  const [tokenData, setTokenData] = useState<TokenResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [regenerating, setRegenerating] = useState(false);
  const [showToken, setShowToken] = useState(false);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadToken();
  }, []);

  const loadToken = async () => {
    try {
      setError(null);
      const data = await api.getToken();
      setTokenData(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : '토큰을 불러오는데 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleRegenerate = async () => {
    if (!confirm('새 토큰을 발급하면 기존 토큰은 더 이상 사용할 수 없습니다. 계속하시겠습니까?')) {
      return;
    }

    setRegenerating(true);
    setError(null);
    try {
      const data = await api.regenerateToken();
      setTokenData(data);
      setShowToken(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : '토큰 재발급에 실패했습니다.');
    } finally {
      setRegenerating(false);
    }
  };

  const copyToken = async () => {
    if (!tokenData?.token) return;

    await navigator.clipboard.writeText(tokenData.token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const maskedToken = (token: string) => {
    if (token.length <= 8) return '••••••••';
    return token.slice(0, 4) + '••••••••' + token.slice(-4);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-muted-foreground">토큰 정보를 불러오는 중...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">API 토큰</h1>
        <p className="text-muted-foreground">
          OpenClaw API에 접근하기 위한 인증 토큰을 관리합니다.
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>API 토큰</CardTitle>
          <CardDescription>
            이 토큰을 사용하여 외부 애플리케이션에서 OpenClaw API에 접근할 수 있습니다.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error ? (
            <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-destructive">
              {error}
            </div>
          ) : tokenData ? (
            <>
              <div className="space-y-2">
                <label className="text-sm font-medium">토큰</label>
                <div className="flex items-center gap-2">
                  <div className="flex-1 rounded-lg border bg-muted p-3 font-mono text-sm">
                    {showToken ? tokenData.token : maskedToken(tokenData.token)}
                  </div>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={() => setShowToken(!showToken)}
                    title={showToken ? '숨기기' : '보기'}
                  >
                    {showToken ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                  <Button
                    variant="outline"
                    size="icon"
                    onClick={copyToken}
                    title="복사"
                  >
                    <Copy className="h-4 w-4" />
                  </Button>
                </div>
                {copied && (
                  <p className="text-sm text-green-600">클립보드에 복사되었습니다.</p>
                )}
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium">생성일</label>
                <p className="text-sm text-muted-foreground">
                  {new Date(tokenData.createdAt).toLocaleString('ko-KR')}
                </p>
              </div>
            </>
          ) : (
            <div className="py-4 text-center text-muted-foreground">
              토큰이 없습니다. 새 토큰을 발급하세요.
            </div>
          )}

          <div className="border-t pt-4">
            <Button
              onClick={handleRegenerate}
              disabled={regenerating}
              variant="outline"
              className="w-full"
            >
              <RefreshCw className={`mr-2 h-4 w-4 ${regenerating ? 'animate-spin' : ''}`} />
              {regenerating ? '발급 중...' : '새 토큰 발급'}
            </Button>
          </div>

          <div className="flex items-start gap-2 rounded-lg border border-yellow-500/50 bg-yellow-500/10 p-3 text-sm text-yellow-700 dark:text-yellow-400">
            <AlertTriangle className="mt-0.5 h-4 w-4 flex-shrink-0" />
            <div>
              <p className="font-medium">보안 주의사항</p>
              <p className="mt-1 text-yellow-600 dark:text-yellow-500">
                이 토큰은 계정에 대한 전체 접근 권한을 가집니다. 절대로 다른 사람과 공유하거나 공개 저장소에 커밋하지 마세요.
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>사용 방법</CardTitle>
          <CardDescription>API 요청 시 토큰을 사용하는 방법</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <p className="mb-2 text-sm font-medium">HTTP 헤더</p>
              <pre className="overflow-x-auto rounded-lg bg-muted p-3 text-sm">
                <code>Authorization: Bearer {'<your-token>'}</code>
              </pre>
            </div>
            <div>
              <p className="mb-2 text-sm font-medium">cURL 예시</p>
              <pre className="overflow-x-auto rounded-lg bg-muted p-3 text-sm">
                <code>{`curl -H "Authorization: Bearer <your-token>" \\
  https://api.openclaw.io/v1/messages`}</code>
              </pre>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
