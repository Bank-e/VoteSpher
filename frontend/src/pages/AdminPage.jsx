import { useState, useEffect, useCallback } from 'react'
import { api } from '../lib/api'

const STATUS_LIST = ['PREPARE', 'OPEN', 'PAUSED', 'CLOSED', 'COUNTING']
const STATUS_META = {
  PREPARE:  { label: 'เตรียมการ',    color: 'gray',    icon: '📋', bg: 'bg-gray-50',    border: 'border-gray-300',    text: 'text-gray-700',    ring: 'ring-gray-200' },
  OPEN:     { label: 'เปิดโหวต',     color: 'emerald', icon: '🟢', bg: 'bg-emerald-50', border: 'border-emerald-400', text: 'text-emerald-700', ring: 'ring-emerald-200' },
  PAUSED:   { label: 'หยุดชั่วคราว', color: 'amber',   icon: '⏸', bg: 'bg-amber-50',   border: 'border-amber-400',   text: 'text-amber-700',   ring: 'ring-amber-200' },
  CLOSED:   { label: 'ปิดโหวต',      color: 'red',     icon: '🔴', bg: 'bg-red-50',     border: 'border-red-400',     text: 'text-red-700',     ring: 'ring-red-200' },
  COUNTING: { label: 'นับคะแนน',     color: 'purple',  icon: '📊', bg: 'bg-purple-50',  border: 'border-purple-400',  text: 'text-purple-700',  ring: 'ring-purple-200' },
}

// Valid state machine transitions
const TRANSITIONS = {
  PREPARE: ['OPEN'],
  OPEN:    ['PAUSED', 'CLOSED'],
  PAUSED:  ['OPEN', 'CLOSED'],
  CLOSED:  ['COUNTING'],
  COUNTING:[],
}

function StatCard({ icon, label, value, sub, accent }) {
  return (
    <div className={`rounded-2xl p-4 border ${accent || 'bg-white border-gray-100 shadow-sm'}`}>
      <div className="flex items-start justify-between">
        <span className="text-xl">{icon}</span>
        {sub && <span className="text-xs text-gray-400 font-medium">{sub}</span>}
      </div>
      <p className="mt-2 text-2xl font-black text-gray-900">{value}</p>
      <p className="text-xs text-gray-500 mt-0.5">{label}</p>
    </div>
  )
}

export default function AdminPage({ token, onLogout }) {
  const [cfg, setCfg]         = useState(null)
  const [areas, setAreas]     = useState([])
  const [form, setForm]       = useState({ status: 'PREPARE', start_time: '', end_time: '' })
  const [loading, setLoading] = useState(true)
  const [saving, setSaving]   = useState(false)
  const [msg, setMsg]         = useState({ type: '', text: '' })

  const load = useCallback(async () => {
    try {
      const [conf, ov] = await Promise.all([api.getConfig(), api.allAreas().catch(() => null)])
      const c = conf?.data
      setCfg(c)
      setAreas(ov?.areas || [])
      if (c) setForm({
        status:     c.status,
        start_time: c.start_time?.slice(0, 16) || '',
        end_time:   c.end_time?.slice(0, 16) || '',
      })
    } catch (e) {
      setMsg({ type: 'error', text: e.message })
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  const save = async () => {
    setSaving(true); setMsg({ type: '', text: '' })
    try {
      const body = { status: form.status }
      if (form.start_time) body.start_time = new Date(form.start_time).toISOString()
      if (form.end_time)   body.end_time   = new Date(form.end_time).toISOString()
      const data = await api.setConfig(body, token)
      setCfg(data.data)
      setMsg({ type: 'success', text: 'บันทึกการตั้งค่าเรียบร้อยแล้ว' })
      setTimeout(() => setMsg({ type: '', text: '' }), 3000)
    } catch (e) {
      setMsg({ type: 'error', text: e.message })
    } finally { setSaving(false) }
  }

  const meta     = STATUS_META[cfg?.status] || STATUS_META.PREPARE
  const totalVotes = areas.reduce((s, a) => s + (a.total_votes || 0), 0)
  const areasWithVotes = areas.filter(a => a.total_votes > 0).length
  const allowed  = TRANSITIONS[form.status] || []

  return (
    <div className="space-y-5 animate-slide-up">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-primary-700 rounded-xl flex items-center justify-center text-lg">🛡</div>
          <div>
            <h2 className="font-bold text-gray-900 text-base">Admin Dashboard</h2>
            <p className="text-xs text-gray-400">VoteSpher 2026</p>
          </div>
        </div>
        <button onClick={onLogout}
          className="flex items-center gap-1 text-xs text-gray-400 hover:text-red-500 transition-colors py-1.5 px-3 rounded-lg hover:bg-red-50">
          <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15M12 9l-3 3m0 0l3 3m-3-3h12.75"/>
          </svg>
          ออกจากระบบ
        </button>
      </div>

      {/* Live status */}
      {cfg && (
        <div className={`rounded-2xl p-4 ${meta.bg} border ${meta.border}`}>
          <div className="flex items-center gap-3">
            <span className="text-2xl">{meta.icon}</span>
            <div className="flex-1">
              <p className={`font-bold ${meta.text}`}>{meta.label}</p>
              <p className="text-xs text-gray-500">
                {cfg.start_time ? new Date(cfg.start_time).toLocaleString('th-TH') : 'ยังไม่ตั้งเวลา'}
                {cfg.end_time ? ` → ${new Date(cfg.end_time).toLocaleString('th-TH')}` : ''}
              </p>
            </div>
            <button onClick={load}
              className="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-black/5 text-gray-500 transition-all">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0013.803-3.7M4.031 9.865a8.25 8.25 0 0113.803-3.7l3.181 3.182m0-4.991v4.99"/>
              </svg>
            </button>
          </div>
        </div>
      )}

      {/* Stats row */}
      {!loading && (
        <div className="grid grid-cols-3 gap-2.5">
          <StatCard icon="🗳" label="คะแนนรวม" value={totalVotes.toLocaleString()}/>
          <StatCard icon="📍" label="เขตที่มีโหวต" value={`${areasWithVotes}/${areas.length}`}/>
          <StatCard icon="📊" label="สถานะ" value={cfg?.status || '…'}/>
        </div>
      )}

      {/* Top areas mini table */}
      {areas.length > 0 && (
        <div className="bg-white rounded-2xl border border-gray-100 shadow-sm overflow-hidden">
          <div className="px-4 py-3 border-b border-gray-50 flex items-center justify-between">
            <p className="text-sm font-semibold text-gray-700">คะแนนแต่ละเขต</p>
            <span className="text-xs text-gray-400">{areas.length} เขต</span>
          </div>
          <div className="divide-y divide-gray-50">
            {[...areas].sort((a,b)=>b.total_votes-a.total_votes).slice(0,5).map((a,i) => {
              const pct = totalVotes > 0 ? Math.round((a.total_votes/totalVotes)*100) : 0
              return (
                <div key={a.area_id} className="flex items-center gap-3 px-4 py-2.5">
                  <span className="text-xs font-bold text-gray-400 w-4">{i+1}</span>
                  <span className="text-sm text-gray-800 flex-1 truncate">{a.area_name}</span>
                  <span className="text-xs font-bold text-primary-700">{a.total_votes.toLocaleString()}</span>
                  <span className="text-xs text-gray-400 w-8 text-right">{pct}%</span>
                </div>
              )
            })}
          </div>
          {areas.length > 5 && (
            <p className="text-center text-xs text-gray-400 py-2">แสดง 5 จาก {areas.length} เขต</p>
          )}
        </div>
      )}

      {/* Settings form */}
      {loading ? (
        <div className="flex justify-center py-8">
          <div className="w-8 h-8 border-3 border-primary-200 border-t-primary-600 rounded-full animate-spin"/>
        </div>
      ) : (
        <div className="bg-white border border-gray-100 rounded-2xl p-5 shadow-sm space-y-5">
          <p className="font-semibold text-gray-700 text-sm">ตั้งค่าการเลือกตั้ง</p>

          {/* Status selector */}
          <div>
            <label className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2 block">สถานะใหม่</label>
            <div className="grid grid-cols-3 gap-2">
              {STATUS_LIST.map(s => {
                const m = STATUS_META[s]
                const isSelected = form.status === s
                const isAllowed  = s === cfg?.status || (cfg && TRANSITIONS[cfg.status]?.includes(s))
                return (
                  <button key={s}
                    onClick={() => isAllowed && setForm(f => ({ ...f, status: s }))}
                    disabled={!isAllowed}
                    title={!isAllowed ? `ไม่สามารถเปลี่ยนจาก ${cfg?.status} → ${s}` : ''}
                    className={`py-2.5 px-2 rounded-xl border-2 text-xs font-bold transition-all flex flex-col items-center gap-1
                      ${isSelected
                        ? `${m.border} ${m.bg} ${m.text} ring-4 ${m.ring}`
                        : isAllowed
                          ? 'border-gray-200 text-gray-600 hover:border-gray-300 hover:bg-gray-50 cursor-pointer'
                          : 'border-gray-100 text-gray-300 cursor-not-allowed opacity-50'
                      }`}
                  >
                    <span>{m.icon}</span>
                    <span>{m.label}</span>
                    {s === cfg?.status && <span className="text-[10px] opacity-60">(ปัจจุบัน)</span>}
                  </button>
                )
              })}
            </div>

            {/* State machine hint */}
            <div className="mt-2 bg-gray-50 rounded-xl px-3 py-2 text-xs text-gray-500 flex items-center gap-1.5">
              <span>จาก <strong className={STATUS_META[cfg?.status]?.text}>{cfg?.status}</strong> → ไปได้:</span>
              {allowed.length === 0
                ? <span className="text-gray-400">ไม่มี (สถานะสุดท้าย)</span>
                : allowed.map(s => (
                  <span key={s} className={`font-semibold ${STATUS_META[s].text}`}>{STATUS_META[s].label}</span>
                ))
              }
            </div>
          </div>

          {/* Date pickers */}
          {[['start_time','วันเวลาเริ่มโหวต'],['end_time','วันเวลาสิ้นสุดโหวต']].map(([key, label]) => (
            <div key={key}>
              <label className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1.5 block">{label}</label>
              <input type="datetime-local" value={form[key]}
                onChange={e => setForm(f => ({ ...f, [key]: e.target.value }))}
                className="input-field text-sm"
              />
            </div>
          ))}

          {/* Messages */}
          {msg.text && (
            <div className={`rounded-xl px-4 py-3 text-sm font-medium flex items-center gap-2 animate-fade-in ${
              msg.type === 'success'
                ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                : 'bg-red-50 text-red-600 border border-red-200'
            }`}>
              <span>{msg.type === 'success' ? '✅' : '⚠'}</span>
              <span>{msg.text}</span>
            </div>
          )}

          <button onClick={save} disabled={saving}
            className="btn-primary flex items-center justify-center gap-2">
            {saving
              ? <><div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"/> กำลังบันทึก…</>
              : <>
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                  บันทึกการตั้งค่า
                </>
            }
          </button>
        </div>
      )}
    </div>
  )
}
