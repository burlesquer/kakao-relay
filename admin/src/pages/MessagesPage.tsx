import React, { useEffect, useState } from 'react';
import { api, InboundMessage, OutboundMessage } from '../lib/api';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '../components/ui/table';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs';
import { Badge } from '../components/ui/badge';
import { Search } from 'lucide-react';

export function MessagesPage() {
  const [activeTab, setActiveTab] = useState('inbound');
  const [messages, setMessages] = useState<(InboundMessage | OutboundMessage)[]>([]);
  const [loading, setLoading] = useState(true);
  const [offset, setOffset] = useState(0);
  const [total, setTotal] = useState(0);
  const [accountId, setAccountId] = useState('');
  const [status, setStatus] = useState('');
  const limit = 50;

  const fetchMessages = async () => {
    setLoading(true);
    try {
      const fetcher = activeTab === 'inbound' ? api.getInboundMessages : api.getOutboundMessages;
      const data = await fetcher(limit, offset, accountId || undefined, status || undefined);
      setMessages(data.items);
      setTotal(data.total);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    setOffset(0);
    fetchMessages();
  }, [activeTab]);

  useEffect(() => {
    fetchMessages();
  }, [offset]);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setOffset(0);
    fetchMessages();
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'processed':
      case 'sent':
        return 'default';
      case 'failed':
        return 'destructive';
      case 'queued':
      case 'pending':
        return 'secondary';
      default:
        return 'outline';
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Messages</h1>
      </div>

      <Tabs defaultValue="inbound" value={activeTab} onValueChange={setActiveTab} className="w-full">
        <TabsList>
          <TabsTrigger value="inbound">Inbound</TabsTrigger>
          <TabsTrigger value="outbound">Outbound</TabsTrigger>
        </TabsList>

        <div className="my-4 flex gap-2">
          <form onSubmit={handleSearch} className="flex gap-2 w-full max-w-2xl">
            <Input
              placeholder="Filter by Account ID"
              value={accountId}
              onChange={(e) => setAccountId(e.target.value)}
              className="max-w-xs"
            />
            <select
              className="flex h-10 w-full max-w-xs rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={status}
              onChange={(e) => setStatus(e.target.value)}
            >
              <option value="">All Statuses</option>
              <option value="queued">Queued</option>
              <option value="processed">Processed</option>
              <option value="sent">Sent</option>
              <option value="failed">Failed</option>
            </select>
            <Button type="submit" variant="secondary">
              <Search className="h-4 w-4" />
            </Button>
          </form>
        </div>

        <TabsContent value="inbound">
          <MessageTable messages={messages} loading={loading} type="inbound" />
        </TabsContent>
        <TabsContent value="outbound">
          <MessageTable messages={messages} loading={loading} type="outbound" />
        </TabsContent>
      </Tabs>

      <div className="flex items-center justify-end space-x-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => setOffset(Math.max(0, offset - limit))}
          disabled={offset === 0}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => setOffset(offset + limit)}
          disabled={offset + limit >= total}
        >
          Next
        </Button>
      </div>
    </div>
  );
}

interface MessageRow {
  id: string;
  accountId: string;
  kakaoUserId?: string;
  content: string;
  status: string;
  createdAt: string;
}

function MessageTable({ messages, loading, type }: { messages: MessageRow[], loading: boolean, type: string }) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>Account ID</TableHead>
            <TableHead>채널 사용자</TableHead>
            <TableHead>Content</TableHead>
            <TableHead>Status</TableHead>
            <TableHead>Created At</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {loading ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center h-24">Loading...</TableCell>
            </TableRow>
          ) : messages.length === 0 ? (
            <TableRow>
              <TableCell colSpan={6} className="text-center h-24">No messages found</TableCell>
            </TableRow>
          ) : (
            messages.map((msg) => (
              <TableRow key={msg.id}>
                <TableCell className="font-mono text-xs">{msg.id}</TableCell>
                <TableCell className="font-mono text-xs">{msg.accountId}</TableCell>
                <TableCell className="text-xs">{msg.kakaoUserId || '-'}</TableCell>
                <TableCell className="max-w-xs truncate" title={JSON.stringify(msg.content)}>
                  {typeof msg.content === 'string' ? msg.content : JSON.stringify(msg.content)}
                </TableCell>
                <TableCell>
                  <Badge variant={msg.status === 'failed' ? 'destructive' : 'secondary'}>
                    {msg.status}
                  </Badge>
                </TableCell>
                <TableCell className="text-muted-foreground text-xs">
                  {new Date(msg.createdAt).toLocaleString()}
                </TableCell>
              </TableRow>
            ))
          )}
        </TableBody>
      </Table>
    </div>
  );
}
