export interface Service { id: string; name: string; status: 'running' | 'stopped'; cpu?: number; mem?: number }
export interface Server { id: string; name: string; status: 'online' | 'down'; cpu: number; mem: number; pm2?: Service[]; docker?: Service[]; }

const SAMPLE: Server[] = [
  { id: 'srv1', name: 'web-01', status: 'online', cpu: 32, mem: 58 },
  { id: 'srv2', name: 'db-01', status: 'online', cpu: 68, mem: 82 },
  { id: 'srv3', name: 'cache-01', status: 'down', cpu: 0, mem: 0 },
];

// There is no /api/servers endpoint in the backend; requesting it on every
// mount produces a 404 in the browser console, so serve the sample data
// directly until an endpoint is added.
export async function fetchServers(): Promise<Server[]> {
  return SAMPLE;
}

export async function fetchServer(id: string): Promise<Server | null> {
  const found = SAMPLE.find(s => s.id === id);
  if (!found) return null;
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
  };
}
