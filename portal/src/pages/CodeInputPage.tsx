import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Key, ArrowRight, Info, CheckCircle2, BarChart3 } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import * as api from '../lib/api';

export default function CodeInputPage() {
  const [code, setCode] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const formatCode = (input: string): string => {
    // Remove non-alphanumeric characters
    const cleaned = input.toUpperCase().replace(/[^A-Z0-9]/g, '');

    // Format as XXXX-XXXX
    if (cleaned.length <= 4) {
      return cleaned;
    }
    return `${cleaned.slice(0, 4)}-${cleaned.slice(4, 8)}`;
  };

  const handleCodeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const formatted = formatCode(e.target.value);
    setCode(formatted);
    setError('');
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (code.length !== 9) { // XXXX-XXXX = 9 characters
      setError('코드는 8자리여야 합니다 (XXXX-XXXX)');
      return;
    }

    setLoading(true);
    setError('');

    try {
      await api.loginWithCode(code);
      navigate('/dashboard');
    } catch (err: any) {
      console.error('Login failed:', err);
      setError(err.message || '코드가 유효하지 않거나 만료되었습니다');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background py-12 px-4 sm:px-6 lg:px-8">
      <div className="w-full max-w-md space-y-6">
        {/* Header */}
        <div className="text-center">
          <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-primary/10 mb-4">
            <Key className="h-6 w-6 text-primary" />
          </div>
          <h1 className="text-3xl font-bold tracking-tight">포털 접속</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            카카오 톡채널 채팅방에서 발급받은 접속 코드를 입력하세요
          </p>
        </div>

        {/* Main Card */}
        <Card>
          <CardHeader>
            <CardTitle>접속 코드 입력</CardTitle>
            <CardDescription>
              8자리 코드를 입력하면 relay 통계를 확인할 수 있습니다
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="space-y-2">
                <label htmlFor="code" className="text-sm font-medium">
                  접속 코드
                </label>
                <input
                  id="code"
                  name="code"
                  type="text"
                  required
                  className="flex h-12 w-full rounded-md border border-input bg-background px-3 py-2 text-center text-lg font-mono tracking-widest ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  placeholder="XXXX-XXXX"
                  value={code}
                  onChange={handleCodeChange}
                  maxLength={9}
                  autoComplete="off"
                  disabled={loading}
                  autoFocus
                />
              </div>

              {error && (
                <div className="rounded-md bg-destructive/10 border border-destructive/20 p-3">
                  <p className="text-sm text-destructive">
                    {error}
                  </p>
                </div>
              )}

              <Button
                type="submit"
                disabled={loading || code.length !== 9}
                className="w-full"
                size="lg"
              >
                {loading ? (
                  '로그인 중...'
                ) : (
                  <>
                    로그인
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </>
                )}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Instructions Card */}
        <Card className="border-primary/20 bg-primary/5">
          <CardContent className="pt-6">
            <div className="flex items-start gap-3">
              <Info className="h-5 w-5 text-primary mt-0.5 flex-shrink-0" />
              <div className="space-y-3 text-sm">
                <div>
                  <p className="font-medium mb-2">코드 받는 방법</p>
                  <ol className="space-y-1.5 text-muted-foreground">
                    <li className="flex items-start gap-2">
                      <span className="font-medium text-primary">1.</span>
                      <span>카카오톡 채팅방에서 <code className="px-1.5 py-0.5 rounded bg-muted font-mono text-xs">/code</code> 입력</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="font-medium text-primary">2.</span>
                      <span>받은 8자리 코드를 위에 입력</span>
                    </li>
                    <li className="flex items-start gap-2">
                      <span className="font-medium text-primary">3.</span>
                      <span>코드는 30분 동안 유효합니다</span>
                    </li>
                  </ol>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Footer Note */}
        <div className="space-y-3">
          <div className="flex items-center justify-center gap-2 text-xs text-muted-foreground">
            <CheckCircle2 className="h-3.5 w-3.5" />
            <p>읽기 전용 모드로 안전하게 조회할 수 있습니다</p>
          </div>

          <div className="flex items-center justify-center">
            <Link
              to="/total"
              className="flex items-center gap-2 text-sm text-primary hover:underline"
            >
              <BarChart3 className="h-4 w-4" />
              <span>전체 통계 보기</span>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
