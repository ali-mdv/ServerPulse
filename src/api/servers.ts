export interface Service { id: string; name: string; status: 'running' | 'stopped'; cpu?: number; mem?: number }
export interface Server { id: string; name: string; status: 'online' | 'down'; cpu: number; mem: number; pm2?: Service[]; docker?: Service[]; }

const SAMPLE: Server[] = [
  { id: 'srv1', name: 'web-01', status: 'online', cpu: 32, mem: 58 },
  { id: 'srv2', name: 'db-01', status: 'online', cpu: 68, mem: 82 },
  { id: 'srv3', name: 'cache-01', status: 'down', cpu: 0, mem: 0 },
];

async function safeFetchJson(url: string, ms = 3000) {
  try {
    // Use Promise.race to implement timeout without aborting the fetch signal which can surface AbortError
    const fetchPromise = fetch(url);
    const timeout = new Promise<null>((resolve) => setTimeout(() => resolve(null), ms));
    const res = await Promise.race([fetchPromise, timeout]) as Response | null;
    if (res === null) return null;
    if (!res.ok) throw new Error('Network response not ok');
    const data = await res.json();
    return data;
  } catch (err) {
    return null;
  }
}

export async function fetchServers(): Promise<Server[]> {
  const data = await safeFetchJson('/api/servers', 3000);
  if (!Array.isArray(data)) return SAMPLE;
  return data as Server[];
}

export async function fetchServer(id: string): Promise<Server | null> {
  const data = await safeFetchJson(`/api/servers/${encodeURIComponent(id)}`, 3000);
  if (data) return data as Server;
  // fallback: find in sample
  const found = SAMPLE.find(s => s.id === id);
  if (found) {
    return {
      ...found,
      pm2: [
        { id: 'p1', name: 'api', status: 'running', cpu: Math.round(Math.random()*30), mem: Math.round(Math.random()*200) },
        { id: 'p2', name: 'worker', status: 'running', cpu: Math.round(Math.random()*20), mem: Math.round(Math.random()*150) }
      ],
      docker: [
        { id: 'd1', name: 'nginx', status: 'running', cpu: Math.round(Math.random()*5), mem: Math.round(Math.random()*50) },
        { id: 'd2', name: 'redis', status: 'running', cpu: Math.round(Math.random()*4), mem: Math.round(Math.random()*40) }
      ]
    } as Server;
  }
  return null;
}
