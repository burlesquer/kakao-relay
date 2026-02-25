import { useEffect, useState } from 'react';
import { CheckCircle2, MessageSquare, Users, Activity } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { api, type PublicStats } from '../lib/api';

export default function DemoPage() {
  const [stats, setStats] = useState<PublicStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadStats = async () => {
      try {
        const data = await api.getPublicStats();
        setStats(data);
      } catch (error) {
        console.error('Failed to load stats', error);
      } finally {
        setLoading(false);
      }
    };

    loadStats();
    const interval = setInterval(loadStats, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-muted-foreground">로딩 중...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="mx-auto max-w-5xl space-y-6">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-bold">서비스 상태</h1>
            {stats?.isPublic && (
              <Badge variant="secondary">공개</Badge>
            )}
          </div>
          <p className="text-muted-foreground">
            카카오톡 채널 릴레이 서비스 실시간 현황
          </p>
        </div>

        <Card className="border-blue-500/50 bg-blue-500/5">
          <CardContent className="py-3">
            <p className="text-sm text-blue-600 dark:text-blue-400">
              시스템 전체 통계입니다. 개인 통계를 보려면 카카오 톡채널 채팅방에서 /status 명령어를 사용하세요.
            </p>
          </CardContent>
        </Card>

        <div className="grid gap-4 md:grid-cols-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">서비스 상태</p>
                  <p className="text-2xl font-bold text-green-500">정상</p>
                </div>
                <CheckCircle2 className="h-8 w-8 text-green-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">등록 계정</p>
                  <p className="text-2xl font-bold">
                    {stats?.system.accounts ?? 0}
                  </p>
                </div>
                <Users className="h-8 w-8 text-blue-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">활성 연결</p>
                  <p className="text-2xl font-bold">
                    {stats?.system.connections ?? 0}
                  </p>
                </div>
                <MessageSquare className="h-8 w-8 text-green-500" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">활성 세션</p>
                  <p className="text-2xl font-bold">
                    {stats?.system.sessions.paired ?? 0}
                  </p>
                </div>
                <Activity className="h-8 w-8 text-primary" />
              </div>
              <p className="mt-2 text-xs text-muted-foreground">
                전체 세션: {stats?.system.sessions.total ?? 0}개
                {(stats?.system.sessions.pending ?? 0) > 0 && (
                  <span className="ml-2 text-yellow-600">
                    (대기 중 {stats?.system.sessions.pending}개)
                  </span>
                )}
              </p>
            </CardContent>
          </Card>
        </div>

        {(stats?.messages.inbound.queued ?? 0) > 0 && (
          <Card className="border-yellow-500/50 bg-yellow-500/5">
            <CardContent className="py-4">
              <p className="text-sm text-yellow-600 dark:text-yellow-400">
                현재 처리 대기 중인 메시지가 {stats?.messages.inbound.queued}개 있습니다.
              </p>
            </CardContent>
          </Card>
        )}

        <Card>
          <CardHeader>
            <CardTitle>기능 안내</CardTitle>
            <CardDescription>
              코드 접속 후 사용할 수 있는 기능들입니다
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-3">
              <div className="rounded-lg border p-4">
                <h3 className="font-semibold">페어링 코드 생성</h3>
                <p className="text-sm text-muted-foreground">
                  카카오톡 채널과 OpenClaw를 연결하는 코드를 생성합니다
                </p>
              </div>
              <div className="rounded-lg border p-4">
                <h3 className="font-semibold">연결 관리</h3>
                <p className="text-sm text-muted-foreground">
                  연결된 대화를 확인하고 차단/해제할 수 있습니다
                </p>
              </div>
              <div className="rounded-lg border p-4">
                <h3 className="font-semibold">메시지 기록</h3>
                <p className="text-sm text-muted-foreground">
                  주고받은 메시지 내역을 확인할 수 있습니다
                </p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>카카오톡 명령어</CardTitle>
            <CardDescription>
              카카오톡 채널에서 사용할 수 있는 명령어입니다
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              <div className="flex items-start gap-3">
                <code className="rounded bg-muted px-2 py-1 text-sm">/status</code>
                <p className="text-sm text-muted-foreground">
                  연결 상태 및 개인 메시지 통계를 확인합니다
                </p>
              </div>
              <div className="flex items-start gap-3">
                <code className="rounded bg-muted px-2 py-1 text-sm">/pair &lt;코드&gt;</code>
                <p className="text-sm text-muted-foreground">
                  OpenClaw에 연결합니다
                </p>
              </div>
              <div className="flex items-start gap-3">
                <code className="rounded bg-muted px-2 py-1 text-sm">/unpair</code>
                <p className="text-sm text-muted-foreground">
                  연결을 해제합니다
                </p>
              </div>
              <div className="flex items-start gap-3">
                <code className="rounded bg-muted px-2 py-1 text-sm">/help</code>
                <p className="text-sm text-muted-foreground">
                  도움말을 표시합니다
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
