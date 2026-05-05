import { useState, useEffect, useCallback } from 'react'
import { api } from '../lib/api'

// ─── Party breakdown drill-down ───────────────────────────────────────────────
function AreaDetail({ area, parties, onBack }) {
  const [data, setData]     = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.partyResult('all', area.area_id)
      .then(setData)
      .catch(() => setData([]))
      .finally(() => setLoading(false))
  }, [area.area_id])

  const partyMap = Object.fromEntries(parties.map(p => [p.party_name, p]))
  const rows = Array.isArray(data) ? data : []
  const total = rows.reduce((s, r) => s + (r.total || 0), 0)
  const max   = rows.reduce((m, r) => Math.max(m, r.total || 0), 1)

  return (
    <div className="animate-slide-up space-y-4">
      <button onClick={onBack} className="flex items-center gap-2 text-sm font-semibold text-primary-600 hover:text-primary-800 transition-colors">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18"/>
        </svg>
        กลับไปภาพรวม
      </button>

      <div className="bg-gradient-to-br from-primary-800 to-primary-900 rounded-2xl p-5 text-white">
        <p className="text-primary-300 text-xs mb-0.5">ผลโหวต</p>
        <h3 className="font-bold text-lg">{area.area_name}</h3>
        <p className="text-primary-300 text-sm mt-1">{total.toLocaleString()} คะแนนรวม</p>
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1,2,3,4].map(i => <div key={i} className="h-16 bg-gray-100 rounded-xl animate-pulse"/>)}
        </div>
      ) : rows.length === 0 ? (
        <div className="text-center py-10 text-gray-400">
          <p className="text-2xl mb-2">📊</p>
          <p className="text-sm">ยังไม่มีคะแนนในเขตนี้</p>
        </div>
      ) : (
        <div className="space-y-3">
          {[...rows].sort((a,b) => b.total - a.total).map((r, i) => {
            const p   = partyMap[r.party_name] || {}
            const pct = total > 0 ? ((r.total / total) * 100).toFixed(1) : 0
            const bar = Math.round((r.total / max) * 100)
            return (
              <div key={i} className="card p-4">
                <div className="flex items-center gap-3 mb-3">
                  {i === 0 && <span className="text-lg">🥇</span>}
                  {i === 1 && <span className="text-lg">🥈</span>}
                  {i === 2 && <span className="text-lg">🥉</span>}
                  {i > 2  && <span className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-xs font-bold text-gray-500">{i+1}</span>}
                  <div className="w-10 h-10 rounded-xl overflow-hidden bg-gray-50 border border-gray-100 flex-shrink-0">
                    {p.logo_url
                      ? <img src={p.logo_url} className="w-full h-full object-contain p-1" alt=""/>
                      : <div className="w-full h-full flex items-center justify-center text-gray-300 text-xs font-bold">?</div>
                    }
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="font-semibold text-gray-900 text-sm truncate">{r.party_name}</p>
                  </div>
                  <div className="text-right">
                    <p className="font-black text-primary-700">{r.total.toLocaleString()}</p>
                    <p className="text-xs text-gray-400">{pct}%</p>
                  </div>
                </div>
                <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                  <div className="h-full rounded-full bg-gradient-to-r from-primary-500 to-primary-700 transition-all duration-700"
                    style={{ width: `${bar}%` }}/>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

// ─── Overview ─────────────────────────────────────────────────────────────────
export default function ResultsPage({ onLogout, justVoted }) {
  const [overview, setOverview]   = useState(null)
  const [parties, setParties]     = useState([])
  const [drillArea, setDrillArea] = useState(null)
  const [loading, setLoading]     = useState(true)
  const [countdown, setCountdown] = useState(10)

  const load = useCallback(async () => {
    try {
      const [ov, pt] = await Promise.all([api.allAreas(), api.parties()])
      setOverview(ov)
      setParties(Array.isArray(pt) ? pt : [])
    } catch {}
    finally { setLoading(false) }
  }, [])

  useEffect(() => { load() }, [])

  useEffect(() => {
    const t = setInterval(() => {
      setCountdown(c => {
        if (c <= 1) { load(); return 10 }
        return c - 1
      })
    }, 1000)
    return () => clearInterval(t)
  }, [load])

  const areas = overview?.areas || []
  const total = overview?.total_votes || 0
  const max   = areas.reduce((m, a) => Math.max(m, a.total_votes), 1)

  if (drillArea) return (
    <div className="animate-fade-in">
      {onLogout && (
        <div className="flex justify-end mb-4">
          <button onClick={onLogout} className="btn-ghost text-xs">ออกจากระบบ</button>
        </div>
      )}
      <AreaDetail area={drillArea} parties={parties} onBack={() => setDrillArea(null)}/>
    </div>
  )

  return (
    <div className="space-y-4 animate-slide-up">
      {/* Success banner */}
      {justVoted && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-2xl p-4 text-center animate-pop">
          <p className="text-2xl mb-1">🎉</p>
          <p className="font-bold text-emerald-800">ลงคะแนนเรียบร้อยแล้ว!</p>
          <p className="text-emerald-600 text-sm mt-0.5">ขอบคุณที่ร่วมใช้สิทธิ์เลือกตั้ง VoteSpher 2026</p>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-bold text-gray-900">ผลโหวต Realtime</h2>
          <p className="text-xs text-gray-400">รีเฟรชใน {countdown}s · กดเขตเพื่อดูรายละเอียด</p>
        </div>
        <button onClick={() => { load(); setCountdown(10) }}
          className="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 text-gray-500 hover:text-gray-700 transition-all">
          <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"/>
          </svg>
        </button>
      </div>

      {loading ? (
        <div className="space-y-3">
          <div className="h-28 bg-gray-100 rounded-2xl animate-pulse"/>
          {[1,2,3].map(i => <div key={i} className="h-20 bg-gray-100 rounded-2xl animate-pulse"/>)}
        </div>
      ) : (
        <>
          {/* Total */}
          <div className="bg-gradient-to-br from-primary-700 via-primary-800 to-slate-900 rounded-2xl p-6 text-white text-center relative overflow-hidden">
            <div className="absolute top-0 right-0 w-32 h-32 bg-white/5 rounded-full -translate-y-1/2 translate-x-1/2"/>
            <p className="text-primary-300 text-sm font-medium mb-1">คะแนนรวมทั้งหมด</p>
            <p className="text-6xl font-black tracking-tight">{total.toLocaleString()}</p>
            <p className="text-primary-300 text-xs mt-2">จาก {areas.length} เขตเลือกตั้ง</p>
          </div>

          {/* Areas */}
          {areas.length === 0 ? (
            <div className="text-center py-10 text-gray-400">
              <p className="text-3xl mb-2">📊</p>
              <p className="text-sm">ยังไม่มีคะแนน</p>
            </div>
          ) : (
            <div className="space-y-2.5">
              {[...areas].sort((a,b) => b.total_votes - a.total_votes).map((a, i) => {
                const pct = total > 0 ? ((a.total_votes / total) * 100).toFixed(1) : 0
                const bar = Math.round((a.total_votes / max) * 100)
                return (
                  <button key={a.area_id} onClick={() => setDrillArea(a)}
                    className="card w-full p-4 text-left hover:shadow-card-hover hover:-translate-y-0.5 transition-all duration-200 group">
                    <div className="flex items-center justify-between mb-2.5">
                      <div className="flex items-center gap-2">
                        {i === 0 && <span className="text-sm">🥇</span>}
                        <span className="font-semibold text-gray-800 text-sm">{a.area_name}</span>
                      </div>
                      <div className="flex items-center gap-2">
                        <div className="text-right">
                          <span className="font-bold text-primary-700 text-sm">{a.total_votes.toLocaleString()}</span>
                          <span className="text-gray-400 text-xs ml-1.5">{pct}%</span>
                        </div>
                        <svg className="w-4 h-4 text-gray-300 group-hover:text-primary-500 transition-colors flex-shrink-0" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5"/>
                        </svg>
                      </div>
                    </div>
                    <div className="h-2 bg-gray-100 rounded-full overflow-hidden">
                      <div className="h-full rounded-full bg-gradient-to-r from-primary-400 to-primary-700 transition-all duration-700"
                        style={{ width: `${bar}%` }}/>
                    </div>
                  </button>
                )
              })}
            </div>
          )}
        </>
      )}

      {onLogout && (
        <button onClick={onLogout}
          className="w-full py-3 text-sm text-gray-400 hover:text-gray-600 border border-gray-200 rounded-xl transition-colors hover:bg-gray-50">
          ออกจากระบบ
        </button>
      )}
    </div>
  )
}
