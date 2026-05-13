export const config = { runtime: 'edge' };

export default async function handler(req: Request): Promise<Response> {
  const base = process.env.API_BASE;
  if (!base) {
    return new Response('API_BASE env var not set on server', { status: 500 });
  }
  const url = new URL(req.url);
  const downstreamPath = url.pathname.replace(/^\/api\/proxy/, '');
  const target = base.replace(/\/$/, '') + downstreamPath + url.search;

  const headers = new Headers(req.headers);
  headers.delete('host');
  headers.delete('connection');

  const hasBody = !['GET', 'HEAD'].includes(req.method);
  const init: RequestInit = {
    method: req.method,
    headers,
    body: hasBody ? await req.arrayBuffer() : undefined,
    redirect: 'manual',
  };

  return fetch(target, init);
}
