import React, { useEffect, useState } from 'react';
import { CheckCircle2, ClipboardCheck, FileSearch, Hash, RefreshCw, ShieldAlert, X } from 'lucide-react';
import { getAuditExport } from '../services/api';

const stateStyle = {
  PENDING: 'bg-yellow-900/50 text-yellow-200', VERIFIED: 'bg-blue-900/50 text-blue-200',
  SETTLING: 'bg-purple-900/50 text-purple-200', SETTLED: 'bg-green-900/50 text-green-200',
  FAILED: 'bg-red-900/50 text-red-200',
};

const short = (value, size = 14) => value ? `${value.slice(0, size)}…${value.slice(-6)}` : '—';
const time = value => value ? new Date(value * 1000).toLocaleTimeString() : '—';

export default function TransactionActivityDashboard({ onClose }) {
  const [audit, setAudit] = useState({ records: [], generated_at: null });
  const [error, setError] = useState(null);
  const load = async () => {
    try { setError(null); setAudit(await getAuditExport()); }
    catch { setError('The audit API is unavailable. Start NEXUS Mesh on port 8001.'); }
  };
  useEffect(() => { load(); const id = setInterval(load, 3000); return () => clearInterval(id); }, []);
  const records = [...audit.records].sort((a, b) => b.created_at - a.created_at);

  return <section className="fixed inset-0 z-50 bg-black/85 p-4 sm:p-6 overflow-auto">
    <div className="max-w-6xl mx-auto bg-[#181818] border border-gray-700 rounded-xl p-5 sm:p-6 shadow-2xl">
      <header className="flex items-start justify-between gap-3 mb-5">
        <div><h2 className="text-2xl font-bold flex items-center gap-2"><FileSearch className="text-[#E50914]" />Payment Activity Audit</h2><p className="text-sm text-gray-400 mt-1">Live trace of parser checks, payload fingerprints, settlement states, and transaction evidence. Refreshes every 3 seconds.</p></div>
        <div className="flex gap-2"><button onClick={load} className="p-2 rounded bg-gray-700 hover:bg-gray-600" aria-label="Refresh"><RefreshCw size={18}/></button><button onClick={onClose} className="p-2" aria-label="Close"><X size={20}/></button></div>
      </header>
      {error && <p className="text-amber-300 bg-amber-950/40 p-3 rounded">{error}</p>}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-5">
        <Metric icon={<ClipboardCheck size={17}/>} label="Traces" value={records.length}/>
        <Metric icon={<CheckCircle2 size={17}/>} label="Settled" value={records.filter(r => r.state === 'SETTLED').length}/>
        <Metric icon={<ShieldAlert size={17}/>} label="Failed" value={records.filter(r => r.state === 'FAILED').length}/>
        <Metric icon={<Hash size={17}/>} label="Last audit" value={audit.generated_at ? time(audit.generated_at) : '—'}/>
      </div>
      {!error && records.length === 0 && <div className="border border-dashed border-gray-600 rounded-lg p-10 text-center text-gray-400">No payment traces yet. Click Play and complete a payment flow; parser, verification, and settlement events will appear here.</div>}
      <div className="space-y-4">{records.map(record => <article key={record.trace_id} className="border border-gray-700 rounded-lg overflow-hidden">
        <div className="p-4 flex flex-wrap gap-3 justify-between items-start bg-[#202020]"><div><strong>{record.trace_id}</strong><p className="text-xs text-gray-400 mt-1">{record.network} · created {time(record.created_at)}</p></div><span className={`text-xs px-2 py-1 rounded font-semibold ${stateStyle[record.state] || 'bg-gray-700'}`}>{record.state}</span></div>
        <div className="grid md:grid-cols-3 gap-4 p-4 text-sm"><Info label="Parser fingerprint" value={short(record.payload_hash)} title={record.payload_hash}/><Info label="Transaction hash" value={short(record.tx_hash)} title={record.tx_hash}/><Info label="Payment" value={`${record.amount ?? '—'} → ${short(record.pay_to, 10)}`}/></div>
        <div className="border-t border-gray-700 p-4"><p className="text-xs uppercase tracking-wider text-gray-500 mb-3">Verification & settlement timeline</p><ol className="flex flex-wrap gap-2">{record.timeline?.map((entry, index) => <li key={`${entry.event}-${index}`} className="bg-[#2a2a2a] rounded px-2.5 py-2 text-xs"><span className="text-gray-200">{entry.event.replace('STATE_CHANGED_TO_', '')}</span><span className="block text-gray-500 mt-1">{time(entry.timestamp)}</span></li>)}</ol>{record.error && <p className="text-red-300 text-xs mt-3">Error: {record.error}</p>}</div>
      </article>)}</div>
    </div>
  </section>;
}

function Metric({ icon, label, value }) { return <div className="bg-[#242424] rounded-lg p-3"><div className="text-gray-400 flex gap-1 items-center text-xs uppercase">{icon}{label}</div><p className="font-bold mt-2">{value}</p></div>; }
function Info({ label, value, title }) { return <div><p className="text-xs uppercase text-gray-500 mb-1">{label}</p><p className="font-mono text-xs text-gray-200 truncate" title={title}>{value}</p></div>; }
