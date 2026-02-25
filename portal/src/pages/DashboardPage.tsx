import React, { useEffect, useState, useMemo } from 'react';
import { useOutletContext } from 'react-router-dom';
import { Unlink, ShieldBan, ShieldCheck, RefreshCw, AlertCircle, CheckCircle2, MessageSquare, ArrowDownToLine, ArrowUpFromLine, Shield } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { Tabs, TabsList, TabsTrigger } from '../components/ui/tabs';
import { api, type Connection, type UserStats, type ConversationStats } from '../lib/api';

type FilterType = 'all' | 'paired' | 'blocked';

export default function DashboardPage() {
  const { isCodeSession } = useOutletContext<{ isCodeSession?: boolean }>();
  const [connections, setConnections] = useState<Connection[]>([]);
  const [pairingCode, setPairingCode] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState<FilterType>('all');
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [stats, setStats] = useState<UserStats | ConversationStats | null>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  const loadConnections = async () => {
    try {
      const { connections: data } = await api.getConnections();
      setConnections(data);
    } catch (error) {
      console.error('Failed to load connections', error);
    } finally {
      setLoading(false);
    }
  };

  const loadStats = async () => {
    try {
      const data = isCodeSession ? await api.getCodeStats() : await api.getStats();
      setStats(data);
    } catch (error) {
      console.error('Failed to load stats', error);
    } finally {
      setStatsLoading(false);
    }
  };

  useEffect(() => {
    if (!isCodeSession) {
      loadConnections();
    } else {
      setLoading(false);
    }
    loadStats();

    const interval = setInterval(loadStats, 30000);
    return () => clearInterval(interval);
  }, [isCodeSession]);

  const filteredConnections = useMemo(() => {
    if (filter === 'all') return connections;
    return connections.filter((conn) => {
      if (filter === 'blocked') return conn.state === 'blocked';
      return conn.state === 'paired' || conn.state === 'active';
    });
  }, [connections, filter]);

  const generateCode = async () => {
    try {
      const { code } = await api.generatePairingCode();
      setPairingCode(code);
    } catch (error) {
      console.error('Failed to generate code', error);
    }
  };

  const copyCode = async () => {
    if (pairingCode) {
      await navigator.clipboard.writeText(pairingCode);
    }
  };

  const handleUnpair = async (conversationKey: string) => {
    if (!confirm('이 연결을 해제하시겠습니까?')) return;

    setActionLoading(conversationKey);
    try {
      await api.unpairConnection(conversationKey);
      setConnections((prev) =>
        prev.filter((conn) => conn.conversationKey !== conversationKey)
      );
    } catch (error) {
      console.error('Failed to unpair', error);
      alert('연결 해제에 실패했습니다.');
    } finally {
      setActionLoading(null);
    }
  };

  const handleBlock = async (conversationKey: string, currentState: Connection['state']) => {
    const isBlocking = currentState !== 'blocked';
    const message = isBlocking
      ? '이 연결을 차단하시겠습니까?'
      : '이 연결의 차단을 해제하시겠습니까?';

    if (!confirm(message)) return;

    setActionLoading(conversationKey);
    try {
      const { state } = await api.blockConnection(conversationKey);
      setConnections((prev) =>
        prev.map((conn) =>
          conn.conversationKey === conversationKey
            ? { ...conn, state }
            : conn
        )
      );
    } catch (error) {
      console.error('Failed to toggle block', error);
      alert('작업에 실패했습니다.');
    } finally {
      setActionLoading(null);
    }
  };

  const getStateBadge = (state: Connection['state']) => {
    switch (state) {
      case 'blocked':
        return <Badge variant="destructive">차단됨</Badge>;
      case 'active':
        return <Badge variant="default">활성</Badge>;
      default:
        return <Badge variant="secondary">연결됨</Badge>;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-muted-foreground">연결 정보를 불러오는 중...</div>
      </div>
    );
  }

  const getHealthStatus = () => {
    if (!stats) return { status: 'unknown', color: 'text-muted-foreground', label: '확인 중...' };
    
    const hasErrors = stats.messages.outbound.failed > 0;
    const hasQueued = stats.messages.inbound.queued > 5;
    const hasConnections = stats.connections.paired > 0;
    
    if (hasErrors) {
      return { status: 'warning', color: 'text-yellow-500', label: '주의 필요' };
    }
    if (!hasConnections) {
      return { status: 'inactive', color: 'text-muted-foreground', label: '연결 없음' };
    }
    if (hasQueued) {
      return { status: 'busy', color: 'text-blue-500', label: '처리 중' };
    }
    return { status: 'healthy', color: 'text-green-500', label: '정상' };
  };

  const healthStatus = getHealthStatus();

  return (
    <div className="space-y-6">
      <div>
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold">대시보드</h1>
          {isCodeSession && (
            <Badge variant="secondary" className="flex gap-1">
              <Shield className="h-3 w-3" />
              읽기 전용 모드
            </Badge>
          )}
        </div>
        {!isCodeSession && (
          <p className="text-muted-foreground">
            대화를 위한 채널 연결:{' '}
            <a
              href="http://pf.kakao.com/_scexbC"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary hover:underline"
            >
              http://pf.kakao.com/_scexbC
            </a>
            {' '}혹은 카카오 톡채널 채팅방에서 'samantha' 검색
          </p>
        )}
        {isCodeSession && (
          <p className="text-muted-foreground">
            코드로 접속하여 relay 통계를 조회하고 있습니다
          </p>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">상태</p>
                <p className={`text-2xl font-bold ${healthStatus.color}`}>
                  {healthStatus.label}
                </p>
              </div>
              {healthStatus.status === 'healthy' ? (
                <CheckCircle2 className="h-8 w-8 text-green-500" />
              ) : healthStatus.status === 'warning' ? (
                <AlertCircle className="h-8 w-8 text-yellow-500" />
              ) : (
                <MessageSquare className="h-8 w-8 text-muted-foreground" />
              )}
            </div>
            {stats?.lastActivity && (
              <p className="mt-2 text-xs text-muted-foreground">
                마지막 활동: {new Date(stats.lastActivity).toLocaleString('ko-KR')}
              </p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">오늘 수신</p>
                <p className="text-2xl font-bold">
                  {statsLoading ? '-' : stats?.messages.inbound.today ?? 0}
                </p>
              </div>
              <ArrowDownToLine className="h-8 w-8 text-blue-500" />
            </div>
            <p className="mt-2 text-xs text-muted-foreground">
              전체: {stats?.messages.inbound.total ?? 0}개
              {(stats?.messages.inbound.queued ?? 0) > 0 && (
                <span className="ml-2 text-yellow-600">
                  (대기 중 {stats?.messages.inbound.queued}개)
                </span>
              )}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">오늘 발신</p>
                <p className="text-2xl font-bold">
                  {statsLoading ? '-' : stats?.messages.outbound.today ?? 0}
                </p>
              </div>
              <ArrowUpFromLine className="h-8 w-8 text-green-500" />
            </div>
            <p className="mt-2 text-xs text-muted-foreground">
              전체: {stats?.messages.outbound.total ?? 0}개
              {(stats?.messages.outbound.failed ?? 0) > 0 && (
                <span className="ml-2 text-red-500">
                  (실패 {stats?.messages.outbound.failed}개)
                </span>
              )}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="pt-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-muted-foreground">연결</p>
                <p className="text-2xl font-bold">
                  {statsLoading ? '-' : stats?.connections.paired ?? 0}
                </p>
              </div>
              <MessageSquare className="h-8 w-8 text-primary" />
            </div>
            <p className="mt-2 text-xs text-muted-foreground">
              전체: {stats?.connections.total ?? 0}개
              {(stats?.connections.blocked ?? 0) > 0 && (
                <span className="ml-2 text-muted-foreground">
                  (차단 {stats?.connections.blocked}개)
                </span>
              )}
            </p>
          </CardContent>
        </Card>
      </div>

      {stats && stats.recentErrors.length > 0 && (
        <Card className="border-destructive/50">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base text-destructive">
              <AlertCircle className="h-4 w-4" />
              최근 오류
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {stats.recentErrors.map((error) => (
                <div
                  key={error.id}
                  className="flex items-start justify-between rounded-md bg-destructive/5 p-2 text-sm"
                >
                  <div className="min-w-0 flex-1">
                    <p className="truncate font-medium text-destructive">
                      {error.errorMessage || '알 수 없는 오류'}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {error.conversationKey} · {new Date(error.createdAt).toLocaleString('ko-KR')}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {!isCodeSession && (
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Pairing Code Card */}
          <Card>
          <CardHeader>
            <CardTitle>페어링 코드</CardTitle>
            <CardDescription>
              카카오톡 채널을 연결하기 위한 코드를 생성하세요
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {pairingCode ? (
              <div className="space-y-4 text-center">
                <div className="rounded-lg bg-muted p-4 font-mono text-4xl font-bold tracking-wider">
                  {pairingCode}
                </div>
                <Button onClick={copyCode} variant="secondary" className="w-full">
                  코드 복사
                </Button>
                <Button onClick={() => setPairingCode(null)} variant="ghost" size="sm">
                  닫기
                </Button>
              </div>
            ) : (
              <Button onClick={generateCode} className="w-full">
                새 코드 생성
              </Button>
            )}
          </CardContent>
        </Card>

        {/* Connections Card */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div className="flex items-center justify-between">
              <div>
                <CardTitle>연결 관리</CardTitle>
                <CardDescription>
                  연결된 카카오톡 대화 목록 ({filteredConnections.length}개)
                </CardDescription>
              </div>
              <Button
                variant="ghost"
                size="icon"
                onClick={loadConnections}
                disabled={loading}
              >
                <RefreshCw className="h-4 w-4" />
              </Button>
            </div>
            <Tabs defaultValue="all" value={filter} onValueChange={(v) => setFilter(v as FilterType)}>
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="all">전체</TabsTrigger>
                <TabsTrigger value="paired">활성</TabsTrigger>
                <TabsTrigger value="blocked">차단됨</TabsTrigger>
              </TabsList>
            </Tabs>
          </CardHeader>
          <CardContent>
            {filteredConnections.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground">
                {filter === 'all'
                  ? '연결된 대화가 없습니다'
                  : filter === 'blocked'
                    ? '차단된 연결이 없습니다'
                    : '활성 연결이 없습니다'}
              </div>
            ) : (
              <div className="space-y-3">
                {filteredConnections.map((conn) => {
                  const isLoading = actionLoading === conn.conversationKey;
                  const isBlocked = conn.state === 'blocked';

                  return (
                    <div
                      key={conn.conversationKey}
                      className="flex items-center justify-between rounded-lg border bg-card p-3"
                    >
                      <div className="mr-4 min-w-0 flex-1">
                        <div className="flex items-center gap-2">
                          <span className="truncate font-medium">
                            {conn.conversationKey}
                          </span>
                          {getStateBadge(conn.state)}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          마지막 활동: {new Date(conn.lastSeenAt).toLocaleString('ko-KR')}
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleBlock(conn.conversationKey, conn.state)}
                          disabled={isLoading}
                          title={isBlocked ? '차단 해제' : '차단'}
                        >
                          {isBlocked ? (
                            <ShieldCheck className="h-4 w-4 text-green-600" />
                          ) : (
                            <ShieldBan className="h-4 w-4 text-yellow-600" />
                          )}
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleUnpair(conn.conversationKey)}
                          disabled={isLoading}
                          title="연결 해제"
                        >
                          <Unlink className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
        </div>
      )}
    </div>
  );
}
