// App state: subscriptions + signals from server, plus client-only viewed map.

import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { api, getViewed, setViewed, type SignalDTO, type SubscriptionDTO } from './api';

export type EnrichedSub = SubscriptionDTO & {
  signals: SignalDTO[];
  unread: boolean;
  newCount: number;
  lastRunRel: string;
};

export interface Toast { kind: 'error' | 'info'; text: string }

export function useApp() {
  const [subs, setSubs] = useState<SubscriptionDTO[]>([]);
  const [signalsBySub, setSignalsBySub] = useState<Record<string, SignalDTO[]>>({});
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [toast, setToast] = useState<Toast | null>(null);
  const [, setViewedTick] = useState(0);
  const bootRan = useRef(false);

  const showError = useCallback((text: string) => {
    setToast({ kind: 'error', text });
    setTimeout(() => setToast(null), 4000);
  }, []);
  const showInfo = useCallback((text: string) => {
    setToast({ kind: 'info', text });
    setTimeout(() => setToast(null), 2500);
  }, []);

  const loadSignalsFor = useCallback(async (id: string) => {
    try {
      const sgs = await api.listSignals(id);
      setSignalsBySub(prev => ({ ...prev, [id]: sgs }));
    } catch (e: any) { showError(e?.message ?? 'failed to load signals'); }
  }, [showError]);

  const refresh = useCallback(async () => {
    setRefreshing(true);
    try {
      await api.ensureDevice();
      const list = await api.listSubscriptions();
      setSubs(list);
      const all: Record<string, SignalDTO[]> = {};
      await Promise.all(list.map(async s => {
        try { all[s.id] = await api.listSignals(s.id); }
        catch { all[s.id] = []; }
      }));
      setSignalsBySub(all);
    } catch (e: any) {
      showError(e?.message ?? 'network error');
    } finally {
      setRefreshing(false);
      setLoading(false);
    }
  }, [showError]);

  useEffect(() => {
    if (bootRan.current) return;
    bootRan.current = true;
    refresh();
  }, [refresh]);

  const create = useCallback(async (query: string, type: 'event' | 'news', cadenceSeconds: number) => {
    try {
      const sub = await api.createSubscription(query, type, cadenceSeconds);
      setSubs(prev => [sub, ...prev]);
      setSignalsBySub(prev => ({ ...prev, [sub.id]: [] }));
      showInfo('Watching');
      // Trigger first run in background so the user sees signals appear soon
      api.runSubscription(sub.id)
        .then(() => loadSignalsFor(sub.id))
        .catch(() => {});
      return sub;
    } catch (e: any) {
      showError(e?.message ?? 'failed to create');
      throw e;
    }
  }, [showError, showInfo, loadSignalsFor]);

  const remove = useCallback(async (id: string) => {
    try {
      await api.deleteSubscription(id);
      setSubs(prev => prev.filter(s => s.id !== id));
      setSignalsBySub(prev => { const c = { ...prev }; delete c[id]; return c; });
      showInfo('Removed');
    } catch (e: any) { showError(e?.message ?? 'delete failed'); }
  }, [showError, showInfo]);

  const runOne = useCallback(async (id: string) => {
    try {
      const r = await api.runSubscription(id);
      await loadSignalsFor(id);
      showInfo(r.new_signals > 0 ? `${r.new_signals} new` : 'No new signals');
    } catch (e: any) { showError(e?.message ?? 'run failed'); }
  }, [loadSignalsFor, showError, showInfo]);

  const markViewed = useCallback((subId: string) => {
    setViewed(subId);
    setViewedTick(t => t + 1);
  }, []);

  const enriched = useMemo<EnrichedSub[]>(() => {
    const viewed = getViewed();
    return subs.map(s => {
      const sgs = signalsBySub[s.id] ?? [];
      const lastViewed = viewed[s.id] ? Date.parse(viewed[s.id]!) : 0;
      const newCount = sgs.filter(sg => Date.parse(sg.first_seen_at) > lastViewed).length;
      const lastRunRel = relativeShort(s.last_run_at);
      return { ...s, signals: sgs, unread: newCount > 0, newCount, lastRunRel };
    });
  }, [subs, signalsBySub]);

  return {
    loading, refreshing, toast, setToast,
    subs: enriched, refresh, create, remove, runOne, markViewed,
  };
}

function relativeShort(iso: string | null | undefined): string {
  if (!iso) return 'never';
  const diff = Math.max(0, Date.now() - Date.parse(iso));
  const m = Math.round(diff / 60_000);
  if (m < 1)  return 'just now';
  if (m < 60) return `${m}m ago`;
  const h = Math.round(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.round(h / 24);
  return `${d}d ago`;
}
