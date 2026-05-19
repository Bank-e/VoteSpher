import { useState, useEffect, useCallback } from 'react'
import { api } from '../lib/api'

// ─── Election timer ───────────────────────────────────────────────────────────
function useTimer(endTime) {
  const [rem, setRem] = useState(0)
  useEffect(() => {
    if (!endTime) return
    const calc = () => Math.max(0, new Date(endTime) - Date.now())
    setRem(calc())
    const t = setInterval(() => setRem(calc()), 1000)
    return () => clearInterval(t)
  }, [endTime])
  const h = Math.floor(rem / 3600000)
  const m = Math.floor((rem % 3600000) / 60000)
  const s = Math.floor((rem % 60000) / 1000)
  return { h, m, s, expired: rem === 0 }
}

// ─── Candidate card ───────────────────────────────────────────────────────────
function CandidateCard({ c, selected, onSelect }) {
  const [open, setOpen] = useState(false)
  const isSelected = selected?.candidate_id === c.candidate_id
  return (
    <div className={`rounded-2xl border-2 transition-all duration-200 overflow-hidden
      ${isSelected
        ? 'border-blue-600 ring-2 ring-blue-300 shadow-lg'
        : 'border-gray-200 hover:border-blue-200'}`}>
      <button onClick={() => onSelect(c)}
        className={`w-full flex items-center gap-4 p-4 text-left transition-colors
          ${isSelected ? 'bg-blue-50' : 'bg-white hover:bg-gray-50'}`}>
        {/* Logo with selected overlay */}
        <div className="relative w-14 h-14 rounded-xl overflow-hidden flex-shrink-0 border border-gray-100 bg-gray-50">
          {c.logo_url
            ? <img src={c.logo_url} alt="" className="w-full h-full object-contain p-1"/>
            : <div className="w-full h-full flex items-center justify-center text-xl font-black text-gray-300">{c.candidate_no}</div>
          }
          {isSelected && (
            <div className="absolute inset-0 bg-blue-600/20 flex items-center justify-center">
              <div className="w-6 h-6 rounded-full bg-blue-600 flex items-center justify-center">
                <svg className="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" strokeWidth={3} viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7"/>
                </svg>
              </div>
            </div>
          )}
        </div>
        {/* Info */}
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-1.5 mb-0.5">
            <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${isSelected ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-600'}`}>
              #{c.candidate_no}
            </span>
            <span className="text-xs text-gray-400 truncate">{c.party_name}</span>
          </div>
          <p className={`font-bold text-sm leading-snug ${isSelected ? 'text-blue-900' : 'text-gray-900'}`}>{c.name}</p>
          {isSelected && <p className="text-xs text-blue-600 font-semibold mt-0.5">✓ เลือกแล้ว</p>}
        </div>
        {/* Radio */}
        <div className={`w-6 h-6 rounded-full border-2 flex-shrink-0 flex items-center justify-center transition-all
          ${isSelected ? 'border-blue-600 bg-blue-600' : 'border-gray-300'}`}>
          {isSelected && <div className="w-2.5 h-2.5 rounded-full bg-white"/>}
        </div>
      </button>
      {/* Bio toggle */}
      {c.biography && (
        <div className={`${isSelected ? 'bg-primary-50' : 'bg-gray-50'} border-t border-gray-100`}>
          <button onClick={() => setOpen(o => !o)}
            className="w-full flex items-center justify-between px-4 py-2 text-xs text-gray-500 hover:text-gray-700">
            <span>ประวัติย่อ</span>
            <svg className={`w-3.5 h-3.5 transition-transform ${open ? 'rotate-180' : ''}`} fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5"/>
            </svg>
          </button>
          {open && <p className="px-4 pb-3 text-xs text-gray-600 leading-relaxed animate-fade-in">{c.biography}</p>}
        </div>
      )}
    </div>
  )
}

// ─── Confirm modal ────────────────────────────────────────────────────────────
function ConfirmModal({ candidate, onConfirm, onCancel, loading }) {
  return (
    <div className="fixed inset-0 z-50 flex items-end sm:items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="bg-white rounded-3xl w-full max-w-sm shadow-2xl animate-pop p-6 space-y-5">
        <div className="text-center">
          <div className="text-4xl mb-3">🗳</div>
          <h3 className="font-bold text-gray-900 text-lg">ยืนยันการลงคะแนน</h3>
          <p className="text-sm text-gray-500 mt-1">การกระทำนี้ไม่สามารถย้อนกลับได้</p>
        </div>
        <div className="bg-primary-50 border border-primary-100 rounded-2xl p-4 flex items-center gap-3">
          <div className="w-12 h-12 rounded-xl overflow-hidden flex-shrink-0 bg-white border border-primary-100">
            {candidate.logo_url
              ? <img src={candidate.logo_url} className="w-full h-full object-contain p-1" alt=""/>
              : <div className="w-full h-full flex items-center justify-center text-primary-300 font-black">{candidate.candidate_no}</div>
            }
          </div>
          <div>
            <p className="font-bold text-primary-900">{candidate.name}</p>
            <p className="text-xs text-primary-500">{candidate.party_name}</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-3">
          <button onClick={onCancel} className="py-3 rounded-xl border-2 border-gray-200 text-sm font-semibold text-gray-600 hover:bg-gray-50 transition-colors">
            ยกเลิก
          </button>
          <button onClick={onConfirm} disabled={loading}
            className="py-3 rounded-xl bg-primary-700 text-white text-sm font-semibold hover:bg-primary-800 disabled:opacity-50 transition-colors flex items-center justify-center gap-2">
            {loading ? <><div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"/> กำลังส่ง…</> : 'ยืนยัน ✓'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ─── Main page ────────────────────────────────────────────────────────────────
export default function VotePage({ token, voterInfo, onVoted, onLogout }) {
  const [candidates, setCandidates] = useState([])
  const [cfg, setCfg]               = useState(null)
  const [selected, setSelected]     = useState(null)
  const [loading, setLoading]       = useState(true)
  const [confirming, setConfirming] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError]           = useState('')
  const timer = useTimer(cfg?.end_time)
  const isOpen = cfg?.status === 'OPEN'

  useEffect(() => {
    Promise.all([api.candidates(voterInfo.area_id), api.getConfig()])
      .then(([cands, elec]) => {
        setCandidates(Array.isArray(cands) ? cands : [])
        setCfg(elec?.data)
      })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [voterInfo.area_id])

  const handleSubmit = async () => {
    setSubmitting(true); setError('')
    try {
      await api.submit({ candidate_id: selected.candidate_id, party_id: selected.party_id }, token)
      onVoted()
    } catch (e) { setError(e.message); setConfirming(false) }
    finally { setSubmitting(false) }
  }

  const STATUS_STYLE = {
    emerald: 'bg-emerald-50 text-emerald-700 border-emerald-200',
    amber:   'bg-amber-50 text-amber-700 border-amber-200',
    red:     'bg-red-50 text-red-700 border-red-200',
    gray:    'bg-gray-50 text-gray-700 border-gray-200',
    purple:  'bg-purple-50 text-purple-700 border-purple-200',
  }
  const StatusBadge = () => {
    const map = {
      OPEN:     ['emerald', 'เปิดโหวต',
        <svg key="o" className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="12" r="4"/></svg>],
      PAUSED:   ['amber',   'หยุดชั่วคราว',
        <svg key="p" className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5"/></svg>],
      CLOSED:   ['red',     'ปิดแล้ว',
        <svg key="c" className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>],
      PREPARE:  ['gray',    'เตรียมการ',
        <svg key="pr" className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>],
      COUNTING: ['purple',  'นับคะแนน',
        <svg key="cn" className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zm9.75-9.75c0-.621.504-1.125 1.125-1.125h2.25C17.496 2.25 18 2.754 18 3.375v16.5C18 20.496 17.496 21 16.875 21h-2.25A1.125 1.125 0 0113.5 19.875V3.375zm-9.75 9.75c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75z"/></svg>],
    }
    const [c,l,i] = map[cfg?.status] || map.PREPARE
    return (
      <div className={`flex items-center gap-2 px-4 py-2.5 rounded-xl text-sm font-semibold border ${STATUS_STYLE[c]}`}>
        {i}<span>{l}</span>
        {isOpen && !timer.expired && (
          <span className="ml-auto font-mono text-xs bg-emerald-100 text-emerald-700 px-2 py-0.5 rounded-lg">
            {String(timer.h).padStart(2,'0')}:{String(timer.m).padStart(2,'0')}:{String(timer.s).padStart(2,'0')}
          </span>
        )}
      </div>
    )
  }

  return (
    <>
      {confirming && selected && (
        <ConfirmModal candidate={selected} onConfirm={handleSubmit} onCancel={() => setConfirming(false)} loading={submitting}/>
      )}
      <div className="space-y-4 animate-slide-up">
        {/* Area info */}
        <div className="flex items-center justify-between bg-gradient-to-r from-primary-800 to-primary-700 rounded-2xl px-5 py-3.5 text-white">
          <div>
            <p className="text-primary-200 text-xs mb-0.5">เขตเลือกตั้ง</p>
            <p className="font-bold">{voterInfo.area_name}</p>
          </div>
          <button onClick={onLogout} className="text-primary-300 hover:text-white text-xs transition-colors">ออกจากระบบ →</button>
        </div>

        {cfg && <StatusBadge/>}

        {loading ? (
          <div className="space-y-3">
            {[1,2,3].map(i => <div key={i} className="h-20 bg-gray-100 rounded-2xl animate-pulse"/>)}
          </div>
        ) : (
          <>
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider">
              ผู้สมัคร {candidates.length} คน · เลือก 1 คน
            </p>
            <div className="space-y-2.5">
              {candidates.map(c => (
                <CandidateCard key={c.candidate_no} c={c} selected={selected}
                  onSelect={s => { setSelected(s); setError('') }}/>
              ))}
            </div>

            {error && (
              <div className="flex gap-2 items-start bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-600 animate-fade-in">
                <span>⚠</span><span>{error}</span>
              </div>
            )}

            <button
              onClick={() => setConfirming(true)}
              disabled={!selected || !isOpen}
              className="btn-primary"
            >
              {!isOpen ? `การโหวตยังไม่เปิด (${cfg?.status || '...'})` : selected ? `ลงคะแนนให้ ${selected.name}` : 'เลือกผู้สมัครก่อน'}
            </button>
          </>
        )}
      </div>
    </>
  )
}
