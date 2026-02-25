import React, { useEffect, useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import { RefreshCw, ArrowDownLeft, ArrowUpRight, ChevronLeft, ChevronRight, Shield } from 'lucide-react';
import { Button } from '../components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Badge } from '../components/ui/badge';
import { Tabs, TabsList, TabsTrigger } from '../components/ui/tabs';
import { api, type Message } from '../lib/api';

type MessageType = 'all' | 'inbound' | 'outbound';

const LIMIT = 20;

export default function MessagesPage() {
  const { isCodeSession } = useOutletContext<{ isCodeSession?: boolean }>();
  const [messages, setMessages] = useState<Message[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [type, setType] = useState<MessageType>('all');
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [hasMore, setHasMore] = useState(false);

  useEffect(() => {
    loadMessages();
  }, [type, offset]);

  const loadMessages = async () => {
    setLoading(true);
    try {
      setError(null);
      const params: { type?: 'inbound' | 'outbound'; limit: number; offset: number } = {
        limit: LIMIT,
        offset,
      };
      if (type !== 'all') {
        params.type = type;
      }
      const data = isCodeSession ? await api.getCodeMessages(params) : await api.getMessages(params);
      setMessages(data.messages);
      setTotal(data.total);
      setHasMore(data.hasMore);
    } catch (err) {
      setError(err instanceof Error ? err.message : '메시지를 불러오는데 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  const handleTypeChange = (newType: string) => {
    setType(newType as MessageType);
    setOffset(0);
  };

  const handlePrev = () => {
    setOffset(Math.max(0, offset - LIMIT));
  };

  const handleNext = () => {
    if (hasMore) {
      setOffset(offset + LIMIT);
    }
  };

  const currentPage = Math.floor(offset / LIMIT) + 1;
  const totalPages = Math.ceil(total / LIMIT);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3">
            <h1 className="text-2xl font-bold">메시지 히스토리</h1>
            {isCodeSession && (
              <Badge variant="secondary" className="flex gap-1">
                <Shield className="h-3 w-3" />
                읽기 전용 모드
              </Badge>
            )}
          </div>
          <p className="text-muted-foreground">
            {isCodeSession
              ? '이 대화의 메시지 기록을 확인합니다.'
              : '송수신된 메시지 기록을 확인합니다.'}
          </p>
        </div>
        <Button variant="outline" onClick={loadMessages} disabled={loading}>
          <RefreshCw className={`mr-2 h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          새로고침
        </Button>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>메시지 목록</CardTitle>
              <CardDescription>총 {total.toLocaleString()}개의 메시지</CardDescription>
            </div>
          </div>
          <Tabs value={type} onValueChange={handleTypeChange}>
            <TabsList className="grid w-full max-w-md grid-cols-3">
              <TabsTrigger value="all">전체</TabsTrigger>
              <TabsTrigger value="inbound" className="flex items-center gap-1">
                <ArrowDownLeft className="h-3 w-3" />
                수신
              </TabsTrigger>
              <TabsTrigger value="outbound" className="flex items-center gap-1">
                <ArrowUpRight className="h-3 w-3" />
                발신
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </CardHeader>
        <CardContent>
          {error ? (
            <div className="rounded-lg border border-destructive/50 bg-destructive/10 p-4 text-destructive">
              {error}
            </div>
          ) : loading ? (
            <div className="flex items-center justify-center py-12">
              <div className="text-muted-foreground">메시지를 불러오는 중...</div>
            </div>
          ) : messages.length === 0 ? (
            <div className="py-12 text-center text-muted-foreground">
              {type === 'all'
                ? '메시지가 없습니다.'
                : type === 'inbound'
                  ? '수신된 메시지가 없습니다.'
                  : '발신된 메시지가 없습니다.'}
            </div>
          ) : (
            <div className="space-y-3">
              {messages.map((message) => (
                <div
                  key={message.id}
                  className="rounded-lg border bg-card p-4"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <Badge
                          variant={message.direction === 'inbound' ? 'default' : 'secondary'}
                          className="flex items-center gap-1"
                        >
                          {message.direction === 'inbound' ? (
                            <>
                              <ArrowDownLeft className="h-3 w-3" />
                              수신
                            </>
                          ) : (
                            <>
                              <ArrowUpRight className="h-3 w-3" />
                              발신
                            </>
                          )}
                        </Badge>
                        <span className="truncate font-mono text-xs text-muted-foreground">
                          {message.conversationKey}
                        </span>
                      </div>
                      <p className="mt-2 whitespace-pre-wrap break-words text-sm">
                        {message.content}
                      </p>
                      <p className="mt-2 text-xs text-muted-foreground">
                        {new Date(message.createdAt).toLocaleString('ko-KR')}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Pagination */}
          {total > LIMIT && (
            <div className="mt-6 flex items-center justify-between border-t pt-4">
              <p className="text-sm text-muted-foreground">
                {currentPage} / {totalPages} 페이지
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handlePrev}
                  disabled={offset === 0 || loading}
                >
                  <ChevronLeft className="mr-1 h-4 w-4" />
                  이전
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleNext}
                  disabled={!hasMore || loading}
                >
                  다음
                  <ChevronRight className="ml-1 h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
