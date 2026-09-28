import React, { useEffect, useState } from 'react';
import { Activity, RefreshCw, ShieldCheck, X } from 'lucide-react';
import { getMeshStatus } from '../services/api';

export default function MeshDashboard({ onClose }) {
  const [status, setStatus] = useState(null);
  const [error, setError] = useState(null);
  const load = async () => {
    try { setError(null); setStatus(await getMeshStatus()); }
    catch { setError('Mesh is offline or unavailable. Start the NEXUS service on port 8001.'); }
  };
  useEffect(() => { load(); }, []);
  return <section className="fixed inset-0 z-50 bg-black/80 p-6 overflow-auto">
    <div className="max-w-4xl mx-auto bg-[#181818] border border-gray-700 rounded-xl p-6 shadow-2xl">
      <div className="flex items-center justify-between mb-6">
        <div><h2 className="text-2xl font-bold flex items-center gap-2"><Activity className="text-[#E50914]" />NEXUS Mesh Operations</h2><p className="text-sm text-gray-400 mt-1">Live facilitator health, routing, and failover evidence.</p></div>
        <div className="flex gap-3"><button onClick={load} className="p-2 rounded bg-gray-700 hover:bg-gray-600" aria-label="Refresh"><RefreshCw size={18}/></button><button onClick={onClose} className="p-2"><X size={20}/></button></div>
      </div>
      {error && <p className="text-amber-300 bg-amber-950/40 p-3 rounded">{error}</p>}
      {status && <>
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 mb-5">
          <Metric label="Mesh status" value={status.status} />
          <Metric label="Live facilitator" value={status.has_live_facilitator ? 'available' : 'fallback only'} />
          <Metric label="Registered nodes" value={status.facilitators.length} />
        </div>
        <div className="space-y-3">{status.facilitators.map(node => <article key={node.name} className="border border-gray-700 rounded-lg p-4 flex flex-wrap gap-3 justify-between">
          <div><strong>{node.name}</strong><p className="text-sm text-gray-400">{node.network} · {node.is_simulator ? 'simulator' : 'live'} · priority {node.priority}</p></div>
          <div className="text-right"><span className={`text-xs px-2 py-1 rounded ${node.circuit_state === 'CLOSED' ? 'bg-green-900 text-green-200' : 'bg-red-900 text-red-200'}`}>{node.circuit_state}</span><p className="text-xs text-gray-400 mt-2">p50 {node.metrics.latency_p50_ms}ms · errors {Math.round(node.metrics.error_rate * 100)}%</p></div>
        </article>)}</div>
        <p className="text-xs text-gray-500 mt-5 flex gap-1"><ShieldCheck size={14}/> Audit export: GET /api/v1/audit/export</p>
      </>}
    </div>
  </section>;
}

function Metric({ label, value }) { return <div className="bg-[#242424] rounded-lg p-3"><p className="text-xs uppercase text-gray-400">{label}</p><p className="font-bold mt-1 capitalize">{value}</p></div>; }
