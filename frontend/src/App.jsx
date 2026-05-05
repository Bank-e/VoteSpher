import { useState, useRef, useEffect } from 'react'
import { api } from './lib/api'
import VotePage from './pages/VotePage'
import ResultsPage from './pages/ResultsPage'
import AdminPage from './pages/AdminPage'

// ─── Spinner icon ─────────────────────────────────────────────────────────────
const Spinner = () => (
  <svg className="animate-spin w-5 h-5" viewBox="0 0 24 24" fill="none">
    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
  </svg>
)

// ─── OTP 6-box input ─────────────────────────────────────────────────────────
function OTPBoxes({ value, onChange }) {
  const refs   = useRef([])
  const digits = value.split('')

  const handleKey = (i, e) => {
    if (e.key === 'Backspace') {
      e.preventDefault()
      const next = [...digits]; next[i] = ''
      onChange(next.join(''))
      if (!digits[i] && i > 0) refs.current[i - 1]?.focus()
    }
  }
  const handleChange = (i, e) => {
    const ch = e.target.value.replace(/\D/g, '').slice(-1)
    const next = [...digits]; next[i] = ch
    onChange(next.join(''))
    if (ch && i < 5) refs.current[i + 1]?.focus()
  }
  const handlePaste = (e) => {
    e.preventDefault()
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
    onChange(pasted.padEnd(6, '').slice(0, 6))
    refs.current[Math.min(pasted.length, 5)]?.focus()
  }

  return (
    <div className="flex gap-2.5 justify-center">
      {[0,1,2,3,4,5].map(i => (
        <input key={i} ref={el => refs.current[i] = el}
          type="text" inputMode="numeric" maxLength={1}
          value={digits[i] || ''} onPaste={handlePaste}
          onChange={e => handleChange(i, e)} onKeyDown={e => handleKey(i, e)}
          onFocus={e => e.target.select()}
          className={`w-11 h-14 text-center text-2xl font-black rounded-xl border-2 transition-all outline-none
            ${digits[i]
              ? 'border-primary-500 bg-primary-50 text-primary-900'
              : 'border-gray-200 bg-white text-gray-800 focus:border-primary-400'
            }`}
        />
      ))}
    </div>
  )
}

// ─── Step 1 — Verify citizen ID ───────────────────────────────────────────────
function StepVerify({ onSuccess }) {
  const [id, setId]           = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError]     = useState('')

  const raw = id.replace(/\D/g, '').slice(0, 13)
  const fmt = raw.replace(/(\d{1})(\d{4})(\d{5})(\d{2})(\d{1})/, '$1-$2-$3-$4-$5')
  const ok  = raw.length === 13

  const submit = async (e) => {
    e.preventDefault(); if (!ok) return
    setLoading(true); setError('')
    try {
      const verify = await api.verify(raw)
      const otp    = await api.otpRequest(verify.voter_id)
      onSuccess({ voterInfo: verify.voter_info, voterId: verify.voter_id, refCode: otp.ref_code })
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  return (
    <form onSubmit={submit} className="space-y-4 animate-slide-up">
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">เลขบัตรประชาชน 13 หลัก</label>
        <div className="relative">
          <input type="text" inputMode="numeric" placeholder="X-XXXX-XXXXX-XX-X"
            value={fmt} onChange={e => setId(e.target.value.replace(/\D/g, ''))}
            className="input-field font-mono text-lg tracking-widest pr-10"
            autoFocus
          />
          {ok && (
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-emerald-500 text-lg">✓</span>
          )}
        </div>
        <p className="text-xs text-gray-400 mt-1.5">ระบบจะส่ง OTP ไปยัง email ที่ลงทะเบียนไว้</p>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-600 flex gap-2 animate-fade-in">
          <span className="flex-shrink-0">⚠</span><span>{error}</span>
        </div>
      )}

      <button type="submit" disabled={!ok || loading}
        className="btn-primary flex items-center justify-center gap-2.5">
        {loading ? <><Spinner /><span>กำลังตรวจสอบ…</span></> : <>
          <span>ยืนยันตัวตนและรับ OTP</span>
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M3 10a.75.75 0 01.75-.75h10.638L10.23 5.29a.75.75 0 111.04-1.08l5.5 5.25a.75.75 0 010 1.08l-5.5 5.25a.75.75 0 11-1.04-1.08l4.158-3.96H3.75A.75.75 0 013 10z" clipRule="evenodd"/>
          </svg>
        </>}
      </button>
    </form>
  )
}

// ─── Step 2 — Enter OTP ───────────────────────────────────────────────────────
function StepOTP({ voterInfo, refCode, onSuccess, onBack }) {
  const [otp, setOtp]           = useState('')
  const [loading, setLoading]   = useState(false)
  const [error, setError]       = useState('')
  const [cooldown, setCooldown] = useState(60)

  useEffect(() => {
    if (cooldown <= 0) return
    const t = setTimeout(() => setCooldown(c => c - 1), 1000)
    return () => clearTimeout(t)
  }, [cooldown])

  const submit = async (e) => {
    e.preventDefault(); if (otp.length !== 6) return
    setLoading(true); setError('')
    try {
      const data = await api.otpConfirm(otp, refCode)
      onSuccess({ token: data.token, role: data.role, voterInfo })
    } catch (e) { setError(e.message); setOtp('') }
    finally { setLoading(false) }
  }

  return (
    <form onSubmit={submit} className="space-y-5 animate-slide-up">
      {/* Voter info badge */}
      <div className="bg-primary-50 border border-primary-100 rounded-xl p-3 flex items-center gap-3">
        <div className="w-9 h-9 bg-primary-700 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0">
          {voterInfo?.area_id}
        </div>
        <div className="min-w-0">
          <p className="text-xs text-primary-500 font-medium">ผู้มีสิทธิ์เลือกตั้ง</p>
          <p className="text-sm font-semibold text-primary-900 truncate">{voterInfo?.area_name}</p>
        </div>
        {voterInfo?.is_voted && (
          <span className="ml-auto text-xs bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full font-semibold flex-shrink-0">
            โหวตแล้ว
          </span>
        )}
      </div>

      {/* Email notice */}
      <div className="flex items-start gap-3 bg-blue-50 border border-blue-100 rounded-xl p-3">
        <svg className="w-5 h-5 text-blue-500 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75"/>
        </svg>
        <div>
          <p className="text-sm font-semibold text-blue-800">ส่ง OTP ไปยัง email แล้ว</p>
          <p className="text-xs text-blue-500 mt-0.5">
            Ref: <span className="font-mono font-bold tracking-wider">{refCode}</span>
            <span className="ml-2 opacity-70">· หมดอายุใน 5 นาที</span>
          </p>
        </div>
      </div>

      {/* OTP boxes */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-3 text-center">กรอกรหัส OTP 6 หลัก</label>
        <OTPBoxes value={otp} onChange={v => { setOtp(v); setError('') }}/>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-600 flex gap-2 animate-fade-in">
          <span className="flex-shrink-0">⚠</span><span>{error}</span>
        </div>
      )}

      <button type="submit" disabled={otp.length !== 6 || loading}
        className="btn-primary flex items-center justify-center gap-2">
        {loading ? <><Spinner /><span>กำลังยืนยัน…</span></> : 'ยืนยัน OTP'}
      </button>

      <div className="flex justify-between items-center">
        <button type="button" onClick={onBack}
          className="text-sm text-gray-400 hover:text-gray-600 transition-colors">← ย้อนกลับ</button>
        <button type="button" disabled={cooldown > 0}
          onClick={() => setCooldown(60)}
          className="text-sm text-primary-600 hover:text-primary-800 disabled:text-gray-300 disabled:cursor-not-allowed transition-colors">
          {cooldown > 0 ? `ส่งใหม่ได้ใน ${cooldown}s` : 'ส่ง OTP ใหม่'}
        </button>
      </div>
    </form>
  )
}

// ─── Step indicator ───────────────────────────────────────────────────────────
function Steps({ current }) {
  const steps = ['ยืนยันตัวตน', 'รหัส OTP']
  return (
    <div className="flex items-center justify-center gap-2 mb-7">
      {steps.map((l, i) => {
        const n = i + 1, done = current > n, active = current === n
        return (
          <div key={i} className="flex items-center gap-2">
            <div className="flex flex-col items-center gap-1">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300 ${
                done   ? 'bg-emerald-500 text-white' :
                active ? 'bg-primary-700 text-white ring-4 ring-primary-100' :
                         'bg-gray-100 text-gray-400'}`}>
                {done
                  ? <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd"/>
                    </svg>
                  : n
                }
              </div>
              <span className={`text-xs font-medium whitespace-nowrap ${active ? 'text-primary-700' : done ? 'text-emerald-600' : 'text-gray-400'}`}>{l}</span>
            </div>
            {i < steps.length - 1 && (
              <div className={`w-10 h-0.5 mb-4 rounded-full transition-colors ${current > n ? 'bg-emerald-400' : 'bg-gray-200'}`}/>
            )}
          </div>
        )
      })}
    </div>
  )
}

// ─── Shell wrapper ────────────────────────────────────────────────────────────
function Shell({ title, subtitle, wide, children }) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-primary-950 to-blue-950 flex flex-col items-center justify-center p-4">
      {/* Background orbs */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 -right-32 w-80 h-80 bg-primary-500/10 rounded-full blur-3xl"/>
        <div className="absolute bottom-1/4 -left-32 w-80 h-80 bg-blue-500/10 rounded-full blur-3xl"/>
      </div>

      <div className={`relative w-full ${wide ? 'max-w-xl' : 'max-w-md'}`}>
        {/* Logo bar */}
        <div className="text-center mb-6">
          <div className="inline-flex items-center gap-3 bg-white/10 backdrop-blur-md px-5 py-3 rounded-2xl border border-white/10">
            <div className="w-9 h-9 bg-primary-500 rounded-xl flex items-center justify-center text-xl">🗳</div>
            <div className="text-left">
              <p className="text-white font-bold text-lg leading-none">VoteSpher</p>
              <p className="text-primary-300 text-xs mt-0.5">การเลือกตั้ง 2026</p>
            </div>
          </div>
        </div>

        {/* Card */}
        <div className="bg-white rounded-3xl shadow-2xl shadow-black/40 overflow-hidden">
          {title && (
            <div className="bg-gradient-to-r from-primary-800 to-primary-700 px-7 py-5">
              <h2 className="text-white font-bold text-lg">{title}</h2>
              {subtitle && <p className="text-primary-200 text-sm mt-0.5">{subtitle}</p>}
            </div>
          )}
          <div className="p-6">{children}</div>
        </div>

        <p className="text-center text-white/30 text-xs mt-5">
          © 2026 VoteSpher · ระบบการเลือกตั้งออนไลน์ที่ปลอดภัย
        </p>
      </div>
    </div>
  )
}

// ─── App root ─────────────────────────────────────────────────────────────────
export default function App() {
  const [step, setStep]           = useState(1)
  const [loginData, setLoginData] = useState({})  // { voterInfo, refCode }
  const [auth, setAuth]           = useState(null) // { token, role, voterInfo } | null
  const [screen, setScreen]       = useState('login') // login | vote | results | admin
  const [justVoted, setJustVoted] = useState(false)

  const reset = () => {
    setStep(1); setLoginData({}); setAuth(null)
    setScreen('login'); setJustVoted(false)
  }

  const afterLogin = ({ token, role, voterInfo }) => {
    setAuth({ token, role, voterInfo })
    if (role === 'admin')    { setScreen('admin');   return }
    if (voterInfo?.is_voted) { setScreen('results'); return }
    setScreen('vote')
  }

  const afterVote = () => { setJustVoted(true); setScreen('results') }

  // ── Authenticated screens ──
  if (screen === 'vote' && auth) return (
    <Shell title="ลงคะแนนเสียง" subtitle={auth.voterInfo?.area_name}>
      <VotePage token={auth.token} voterInfo={auth.voterInfo} onVoted={afterVote} onLogout={reset}/>
    </Shell>
  )

  if (screen === 'results') return (
    <Shell title="ผลโหวต Realtime" subtitle="VoteSpher 2026" wide>
      <ResultsPage onLogout={auth ? reset : null} justVoted={justVoted}/>
    </Shell>
  )

  if (screen === 'admin' && auth) return (
    <Shell title="Admin Dashboard" subtitle="VoteSpher 2026">
      <AdminPage token={auth.token} onLogout={reset}/>
    </Shell>
  )

  // ── Login flow ──
  return (
    <Shell>
      <Steps current={step}/>

      {step === 1 && (
        <>
          <StepVerify onSuccess={({ voterInfo, refCode }) => {
            setLoginData({ voterInfo, refCode })
            setStep(2)
          }}/>

          {/* Public results link */}
          <div className="mt-5 pt-4 border-t border-gray-100 text-center">
            <button onClick={() => setScreen('results')}
              className="text-sm text-primary-600 hover:text-primary-800 font-medium transition-colors flex items-center gap-1.5 mx-auto">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" strokeWidth={2} viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z"/>
              </svg>
              ดูผลโหวต Realtime (ไม่ต้อง login)
            </button>
          </div>
        </>
      )}

      {step === 2 && (
        <StepOTP
          voterInfo={loginData.voterInfo}
          refCode={loginData.refCode}
          onSuccess={afterLogin}
          onBack={() => { setStep(1); setLoginData({}) }}
        />
      )}
    </Shell>
  )
}
