import React, { useEffect, useState, useMemo } from 'react';
import { api, PluginSession } from '../lib/api';
import { Button } from '../components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table';
import { Badge } from '../components/ui/badge';
import { Trash2, Unplug, Search } from 'lucide-react';
import { Input } from '../components/ui/input';

const statusColors: Record<PluginSession['status'], 'default' | 'secondary' | 'destructive' | 'outline'> = {
  pending_pairing: 'secondary',
  paired: 'default',
  expired: 'outline',
  disconnected: 'destructive',
};

const statusLabels: Record<PluginSession['status'], string> = {
  pending_pairing: '대기중',
  paired: '연결됨',
  expired: '만료됨',
  disconnected: '연결해제',
};

export function SessionsPage() {
  const [sessions, setSessions] = useState<PluginSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const limit = 50;

  // Filter states
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('all');

  // Client-side filtering
  const filteredSessions = useMemo(() => {
    return sessions.filter((session) => {
      const matchesSearch = searchQuery === '' ||
        session.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        session.pairingCode.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (session.accountId && session.accountId.toLowerCase().includes(searchQuery.toLowerCase()));
      const matchesStatus = statusFilter === 'all' || session.status === statusFilter;
      return matchesSearch && matchesStatus;
    });
  }, [sessions, searchQuery, statusFilter]);

  const fetchSessions = async () => {
    setLoading(true);
    try {
      const data = await api.getSessions(limit, offset);
      setSessions(data.items);
      setTotal(data.total);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSessions();
  }, [offset]);

  const handleDelete = async (id: string) => {
    if (!confirm('이 세션을 삭제하시겠습니까?')) return;
    try {
      await api.deleteSession(id);
      fetchSessions();
    } catch (error) {
      alert('세션 삭제에 실패했습니다.');
    }
  };

  const handleDisconnect = async (id: string) => {
    if (!confirm('이 세션의 연결을 해제하시겠습니까? 플러그인은 더 이상 메시지를 수신할 수 없습니다.')) return;
    try {
      await api.disconnectSession(id);
      fetchSessions();
    } catch (error) {
      alert('세션 연결 해제에 실패했습니다.');
    }
  };

  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return '-';
    return new Date(dateStr).toLocaleString('ko-KR', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">플러그인 세션</h1>
        <p className="text-muted-foreground mt-1">
          OpenClaw 플러그인의 연결 세션을 관리합니다. 세션은 플러그인이 /pair 명령으로 카카오 채널과 연결할 때 생성됩니다.
        </p>
      </div>

      {/* Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="ID, 페어링 코드, Account ID로 검색..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <select
          className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
        >
          <option value="all">모든 상태</option>
          <option value="pending_pairing">대기중</option>
          <option value="paired">연결됨</option>
          <option value="expired">만료됨</option>
          <option value="disconnected">연결해제</option>
        </select>
        <Button variant="outline" onClick={() => { setSearchQuery(''); setStatusFilter('all'); }}>
          초기화
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>페어링 코드</TableHead>
              <TableHead>상태</TableHead>
              <TableHead>Account ID</TableHead>
              <TableHead>연결 시간</TableHead>
              <TableHead>만료 시간</TableHead>
              <TableHead>생성일</TableHead>
              <TableHead className="text-right">작업</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center h-24">불러오는 중...</TableCell>
              </TableRow>
            ) : filteredSessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-center h-24">
                  {sessions.length === 0 ? '등록된 세션이 없습니다.' : '검색 결과가 없습니다.'}
                </TableCell>
              </TableRow>
            ) : (
              filteredSessions.map((session) => (
                <TableRow key={session.id}>
                  <TableCell className="font-mono">{session.pairingCode}</TableCell>
                  <TableCell>
                    <Badge variant={statusColors[session.status]}>
                      {statusLabels[session.status]}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-mono text-xs">
                    {session.accountId ? session.accountId.slice(0, 8) + '...' : '-'}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {formatDate(session.pairedAt)}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {formatDate(session.expiresAt)}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-xs">
                    {formatDate(session.createdAt)}
                  </TableCell>
                  <TableCell className="text-right space-x-2">
                    {session.status === 'paired' && (
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDisconnect(session.id)}
                        title="연결 해제"
                      >
                        <Unplug className="h-4 w-4" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="text-destructive"
                      onClick={() => handleDelete(session.id)}
                      title="삭제"
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-end space-x-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0}
        >
          이전
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setOffset(offset + limit)}
          disabled={offset + limit >= total}
        >
          다음
        </Button>
      </div>
    </div>
  );
}
