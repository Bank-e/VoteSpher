import { useState, useEffect, useCallback } from 'react'
import { api } from '../lib/api'
import Election3DScene from '../components/Election3DScene'

// ─── Per-area drill-down (party breakdown + candidates) ──────────────────────
function AreaDetail({ area, partyLogoMap, onBack }) {
  const [data, setData]       = useState(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.areaResult(area.area_id)
      .then(setData)
      .catch(() => setData({ candidate_results: [], party_list_results: [] }))
      .finally(() => setLoading(false))
  }, [area.area_id])

  const partyRows = data?.party_list_results || []
  const candidateRows = data?.candidate_results || []
  const totalParty = partyRows.reduce((s, r) => s + (r.votes || 0), 0)
  const maxParty   = partyRows.reduce((m, r) => Math.max(m, r.votes || 0), 1)

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
        <p className="text-primary-300 text-sm mt-1">{area.total_votes.toLocaleString()} คะแนนรวม</p>
      </div>

      {loading ? (
        <div className="space-y-3">
          {[1,2,3,4].map(i => <div key={i} className="h-16 bg-gray-100 rounded-xl animate-pulse"/>)}
        </div>
      ) : (
        <>
          {/* Party list results */}
          {partyRows.length > 0 && (
            <div>
              <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">บัญชีรายชื่อพรรค</p>
              <div className="space-y-2.5">
                {[...partyRows].sort((a,b) => b.votes - a.votes).map((r, i) => {
                  const logoUrl = partyLogoMap?.[r.party_name]
                  const pct = totalParty > 0 ? ((r.votes / totalParty) * 100).toFixed(1) : 0
                  const bar = Math.round((r.votes / maxParty) * 100)
                  return (
                    <div key={i} className="card p-4">
                      <div className="flex items-center gap-3 mb-2.5">
                        {i === 0 && <span className="text-lg">🥇</span>}
                        {i === 1 && <span className="text-lg">🥈</span>}
                        {i === 2 && <span className="text-lg">🥉</span>}
                        {i > 2  && <span className="w-6 h-6 rounded-full bg-gray-100 flex items-center justify-center text-xs font-bold text-gray-500">{i+1}</span>}
                        <div className="w-10 h-10 rounded-xl overflow-hidden bg-gray-50 border border-gray-100 flex-shrink-0">
                          {logoUrl
                            ? <img src={logoUrl} className="w-full h-full object-contain p-1" alt=""/>
                            : <div className="w-full h-full flex items-center justify-center text-gray-300 text-xs font-bold">?</div>}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="font-semibold text-gray-900 text-sm truncate">{r.party_name}</p>
                        </div>
                        <div className="text-right">
                          <p className="font-black text-primary-700">{r.votes.toLocaleString()}</p>
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
            </div>
          )}

          {/* Candidate results */}
          {candidateRows.length > 0 && (
            <div>
              <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 mt-1">ผู้สมัครแบบแบ่งเขต</p>
              <div className="space-y-2">
                {[...candidateRows].sort((a,b) => b.votes - a.votes).slice(0, 5).map((c, i) => (
                  <div key={i} className="card px-4 py-3 flex items-center gap-3">
                    <span className="text-xs font-bold text-gray-400 w-5">#{i+1}</span>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-semibold text-gray-900 truncate">{c.name}</p>
                      <p className="text-xs text-gray-400 truncate">{c.party_name}</p>
                    </div>
                    <span className="text-sm font-black text-primary-700">{c.votes.toLocaleString()}</span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {partyRows.length === 0 && candidateRows.length === 0 && (
            <div className="text-center py-10 text-gray-400">
              <p className="text-2xl mb-2">📊</p>
              <p className="text-sm">ยังไม่มีคะแนนในเขตนี้</p>
            </div>
          )}
        </>
      )}
    </div>
  )
}

// ─── Party overview section ───────────────────────────────────────────────────
function PartyOverview({ parties, total }) {
  if (!parties || parties.length === 0) return null
  const sorted = [...parties].sort((a,b) => b.votes - a.votes)
  const max = sorted[0]?.votes || 1
  return (
    <div className="card p-4">
      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">
        คะแนนบัญชีรายชื่อ · รวมทั้งประเทศ
      </p>
      <div className="space-y-3">
        {sorted.map((p, i) => {
          const pct = total > 0 ? ((p.votes / total) * 100).toFixed(1) : 0
          const bar = Math.round((p.votes / max) * 100)
          return (
            <div key={p.party_no}>
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  {i === 0 && <span className="text-sm">🥇</span>}
                  {i === 1 && <span className="text-sm">🥈</span>}
                  {i === 2 && <span className="text-sm">🥉</span>}
                  {i > 2 && <span className="w-5 h-5 rounded-full bg-gray-100 flex items-center justify-center text-xs font-bold text-gray-500">{i+1}</span>}
                  <span className="text-sm font-semibold text-gray-800">{p.party_name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="font-bold text-primary-700 text-sm">{p.votes.toLocaleString()}</span>
                  <span className="text-xs text-gray-400 w-10 text-right">{pct}%</span>
                </div>
              </div>
              <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                <div className="h-full rounded-full bg-gradient-to-r from-primary-400 to-primary-700 transition-all duration-700"
                  style={{ width: `${bar}%` }}/>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}

// ─── Overview ─────────────────────────────────────────────────────────────────
export default function ResultsPage({ onLogout, justVoted }) {
  const [overview, setOverview]     = useState(null)
  const [partyLogoMap, setPartyLogoMap] = useState({})
  const [drillArea, setDrillArea]   = useState(null)
  const [loading, setLoading]       = useState(true)
  const [countdown, setCountdown]   = useState(10)

  const load = useCallback(async () => {
    try {
      const [ov, pts] = await Promise.all([api.allAreas(), api.parties()])
      setOverview(ov)
      // build party name → logo_url map for drill-down
      const map = {}
      if (Array.isArray(pts)) pts.forEach(p => { map[p.party_name] = p.logo_url })
      setPartyLogoMap(map)
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
  const partyTotals = overview?.party || []
  const total = overview?.total_votes || 0
  const max   = areas.reduce((m, a) => Math.max(m, a.total_votes), 1)

  if (drillArea) return (
    <div className="animate-fade-in">
      {onLogout && (
        <div className="flex justify-end mb-4">
          <button onClick={onLogout} className="btn-ghost text-xs">ออกจากระบบ</button>
        </div>
      )}
      <AreaDetail area={drillArea} partyLogoMap={partyLogoMap} onBack={() => setDrillArea(null)}/>
    </div>
  )

  return (
    <div className="animate-slide-up">
      {/* Success banner */}
      {justVoted && (
        <div className="bg-emerald-50 border border-emerald-200 rounded-2xl p-4 text-center animate-pop mb-4">
          <p className="font-bold text-emerald-800">ลงคะแนนเรียบร้อยแล้ว!</p>
          <p className="text-emerald-600 text-sm mt-0.5">ขอบคุณที่ร่วมใช้สิทธิ์เลือกตั้ง VoteSpher 2026</p>
        </div>
      )}

      {loading ? (
        <div className="space-y-3">
          <div className="h-96 bg-gray-100 rounded-2xl animate-pulse"/>
          <div className="h-20 bg-gray-100 rounded-2xl animate-pulse"/>
        </div>
      ) : (
        <div className="flex flex-col lg:flex-row gap-4">

          {/* ── LEFT: 3D scene (sticky) ─────────────────────── */}
          <div className="lg:flex-1 lg:sticky lg:top-0 lg:self-start space-y-3">
            <Election3DScene areas={areas}/>

            {/* Total votes hero */}
            <div className="bg-gradient-to-r from-primary-800 to-primary-700 rounded-2xl px-5 py-3 text-white flex items-center justify-between">
              <div>
                <p className="text-primary-300 text-xs">คะแนนรวมทั้งหมด</p>
                <p className="text-3xl font-black tracking-tight">{total.toLocaleString()}</p>
              </div>
              <div className="text-right">
                <p className="text-primary-300 text-sm">{areas.length} เขต</p>
                <button onClick={() => { load(); setCountdown(10) }}
                  className="text-primary-300 hover:text-white text-xs flex items-center gap-1 mt-0.5 ml-auto transition-colors">
                  <svg className="w-3 h-3" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"/>
                  </svg>
                  {countdown}s
                </button>
              </div>
            </div>

            <PartyOverview parties={partyTotals} total={total}/>
          </div>

          {/* ── RIGHT: area list ────────────────────────────── */}
          <div className="lg:w-72 lg:max-h-[calc(100vh-160px)] lg:overflow-y-auto space-y-2 pr-0.5">
            <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider px-1">
              รายชื่อเขต · กดดูรายละเอียด
            </p>
            {[...areas].sort((a,b) => a.area_id - b.area_id).map((a) => {
              const pct = total > 0 ? ((a.total_votes / total) * 100).toFixed(1) : 0
              const bar = Math.round((a.total_votes / max) * 100)
              const winner = a.candidates?.[0]
              return (
                <button key={a.area_id} onClick={() => setDrillArea(a)}
                  className="card w-full p-3 text-left hover:shadow-card-hover hover:-translate-y-0.5 transition-all duration-200 group">
                  <div className="flex items-center justify-between mb-1.5">
                    <div className="min-w-0">
                      <p className="font-semibold text-gray-800 text-sm truncate">{a.area_name}</p>
                      {winner && <p className="text-xs text-gray-400 truncate">{winner.party_name}</p>}
                    </div>
                    <div className="flex items-center gap-1.5 flex-shrink-0 ml-2">
                      <span className="font-bold text-primary-700 text-sm">{a.total_votes.toLocaleString()}</span>
                      <svg className="w-3.5 h-3.5 text-gray-300 group-hover:text-primary-500 transition-colors" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5"/>
                      </svg>
                    </div>
                  </div>
                  <div className="h-1.5 bg-gray-100 rounded-full overflow-hidden">
                    <div className="h-full rounded-full bg-gradient-to-r from-primary-400 to-primary-700 transition-all duration-700"
                      style={{ width: `${bar}%` }}/>
                  </div>
                </button>
              )
            })}

            {onLogout && (
              <button onClick={onLogout}
                className="w-full py-2.5 text-sm text-gray-400 hover:text-gray-600 border border-gray-200 rounded-xl transition-colors hover:bg-gray-50 mt-2">
                ออกจากระบบ
              </button>
            )}
          </div>

        </div>
      )}
    </div>
  )
}
