// HTTP client for Go server. Mirrors shared/ApiClient.kt route surface.

export interface SubscriptionDTO {
  id: string;
  query: string;
  type: 'event' | 'news';
  cadence_seconds: number;
  last_run_at: string | null;
  next_run_at: string;
  created_at: string;
}

export interface SignalDTO {
  id: string;
  subscription_id: string;
  title: string;
  body?: string | null;
  url?: string | null;
  occurs_at?: string | null;
  source_domains: string[];
  confidence: number;
  first_seen_at: string;
}

export interface RunResponse { new_signals: number }

const DEVICE_KEY  = 'notify.device_id';
const VIEWED_KEY  = 'notify.viewed_at'; // { [subId]: ISO string }

export type ViewedMap = Record<string, string>;
export const getViewed = (): ViewedMap => {
  try { return JSON.parse(localStorage.getItem(VIEWED_KEY) ?? '{}'); }
  catch { return {}; }
};
export const setViewed = (subId: string, iso = new Date().toISOString()) => {
  const m = getViewed(); m[subId] = iso;
  localStorage.setItem(VIEWED_KEY, JSON.stringify(m));
};

class ApiError extends Error {
  constructor(public status: number, message: string) { super(`[${status}] ${message}`); }
}

async function request<T>(method: string, path: string, body?: unknown, deviceId?: string | null): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (deviceId) headers['X-Device-Id'] = deviceId;
  const r = await fetch(path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  if (!r.ok) {
    const text = await r.text().catch(() => '');
    throw new ApiError(r.status, text || r.statusText);
  }
  if (r.status === 204) return undefined as T;
  const ct = r.headers.get('content-type') ?? '';
  if (!ct.includes('json')) return undefined as T;
  return r.json() as Promise<T>;
}

export class Api {
  deviceId: string | null = null;

  constructor() {
    this.deviceId = localStorage.getItem(DEVICE_KEY);
  }

  /** Register once. Web has no APNS — synthesize a stable token. */
  async ensureDevice(): Promise<string> {
    if (this.deviceId) return this.deviceId;
    let token = localStorage.getItem('notify.web_token');
    if (!token) {
      token = 'web-' + crypto.randomUUID();
      localStorage.setItem('notify.web_token', token);
    }
    const resp = await request<{ device_id: string }>('POST', '/v1/devices', { apns_token: token });
    this.deviceId = resp.device_id;
    localStorage.setItem(DEVICE_KEY, resp.device_id);
    return resp.device_id;
  }

  listSubscriptions = () => request<SubscriptionDTO[]>('GET', '/v1/subscriptions', undefined, this.deviceId);

  createSubscription = (query: string, type: 'event' | 'news', cadenceSeconds: number) =>
    request<SubscriptionDTO>('POST', '/v1/subscriptions',
      { query, type, cadence_seconds: cadenceSeconds }, this.deviceId);

  deleteSubscription = (id: string) =>
    request<void>('DELETE', `/v1/subscriptions/${id}`, undefined, this.deviceId);

  listSignals = (subId: string, limit = 50) =>
    request<SignalDTO[]>('GET', `/v1/subscriptions/${subId}/signals?limit=${limit}`, undefined, this.deviceId);

  runSubscription = (id: string) =>
    request<RunResponse>('POST', `/v1/subscriptions/${id}/run`, undefined, this.deviceId);
}

export const api = new Api();

export const CADENCE_SECONDS: Record<string, number> = {
  '5m':  5 * 60,
  '15m': 15 * 60,
  '1h':  60 * 60,
  '6h':  6 * 60 * 60,
  '1d':  24 * 60 * 60,
};

export function cadenceLabel(seconds: number): string {
  if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
  if (seconds < 86400) return `${Math.round(seconds / 3600)}h`;
  return `${Math.round(seconds / 86400)}d`;
}
