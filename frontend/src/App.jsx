import { useState, useRef, useEffect } from 'react'
import { api } from './lib/api'
import VotePage from './pages/VotePage'
import ResultsPage from './pages/ResultsPage'
import AdminPage from './pages/AdminPage'

// ─── SVG icon components ──────────────────────────────────────────────────────
const IconEmail = () => (
  <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75"/>
  </svg>
)
const IconPhone = () => (
  <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 1.5H8.25A2.25 2.25 0 006 3.75v16.5a2.25 2.25 0 002.25 2.25h7.5A2.25 2.25 0 0018 20.25V3.75a2.25 2.25 0 00-2.25-2.25H13.5m-3 0V3h3V1.5m-3 0h3m-3 20.25h3"/>
  </svg>
)
const IconWarn = () => (
  <svg className="w-4 h-4 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" strokeWidth={1.5} viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z"/>
  </svg>
)

// ─── Spinner ──────────────────────────────────────────────────────────────────
const Spinner = () => (
  <svg className="animate-spin w-5 h-5" viewBox="0 0 24 24" fill="none">
    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"/>
  </svg>
)

// ─── OTP 6-box input ──────────────────────────────────────────────────────────
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
  const [channel, setChannel] = useState('email')
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
      onSuccess({ voterInfo: verify.voter_info, voterId: verify.voter_id, channel })
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
          {ok && <span className="absolute right-3 top-1/2 -translate-y-1/2 text-emerald-500 text-lg">✓</span>}
        </div>
      </div>

      {/* Channel selector */}
      <div>
        <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">รับ OTP ทาง</p>
        <div className="grid grid-cols-2 gap-2">
          {[{ value: 'email', icon: <IconEmail/>, label: 'Email' }, { value: 'sms', icon: <IconPhone/>, label: 'SMS' }].map(opt => (
            <button key={opt.value} type="button" onClick={() => setChannel(opt.value)}
              className={`flex items-center gap-2.5 px-4 py-3 rounded-xl border-2 text-sm font-semibold transition-all
                ${channel === opt.value
                  ? 'border-primary-600 bg-primary-50 text-primary-800'
                  : 'border-gray-200 text-gray-600 hover:border-gray-300 hover:bg-gray-50'}`}>
              {opt.icon}
              <span>{opt.label}</span>
              <div className={`ml-auto w-4 h-4 rounded-full border-2 flex items-center justify-center transition-all
                ${channel === opt.value ? 'border-primary-600 bg-primary-600' : 'border-gray-300'}`}>
                {channel === opt.value && <div className="w-1.5 h-1.5 rounded-full bg-white"/>}
              </div>
            </button>
          ))}
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-600 flex gap-2 animate-fade-in">
          <IconWarn/><span>{error}</span>
        </div>
      )}

      <button type="submit" disabled={!ok || loading}
        className="btn-primary flex items-center justify-center gap-2.5">
        {loading ? <><Spinner /><span>กำลังตรวจสอบ…</span></> : <>
          <span>ยืนยันตัวตน</span>
          <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M3 10a.75.75 0 01.75-.75h10.638L10.23 5.29a.75.75 0 111.04-1.08l5.5 5.25a.75.75 0 010 1.08l-5.5 5.25a.75.75 0 11-1.04-1.08l4.158-3.96H3.75A.75.75 0 013 10z" clipRule="evenodd"/>
          </svg>
        </>}
      </button>
    </form>
  )
}

// ─── Step 2 — Enter/confirm delivery address, then send OTP ──────────────────
function StepDelivery({ voterInfo, channel, voterId, onSuccess, onBack }) {
  const defaultContact = channel === 'sms'
    ? (voterInfo?.masked_phone || '')
    : (voterInfo?.masked_email || '')

  const [address, setAddress] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError]     = useState('')

  const isEmail = channel === 'email'
  const label   = isEmail ? 'Email' : 'เบอร์โทรศัพท์'
  const ChannelIcon = isEmail ? IconEmail : IconPhone

  const submit = async (e) => {
    e?.preventDefault()
    setLoading(true); setError('')
    try {
      const otp = await api.otpRequest(voterId, channel, address.trim() || undefined)
      onSuccess({ refCode: otp.ref_code, sentTo: otp.masked_contact || address.trim() || defaultContact })
    } catch (e) { setError(e.message) }
    finally { setLoading(false) }
  }

  return (
    <div className="space-y-4 animate-slide-up">
      {/* Voter badge */}
      <div className="bg-primary-50 border border-primary-100 rounded-xl p-3 flex items-center gap-3">
        <div className="w-9 h-9 bg-primary-700 rounded-full flex items-center justify-center text-white text-sm font-bold flex-shrink-0">
          {voterInfo?.area_id}
        </div>
        <div className="min-w-0">
          <p className="text-xs text-primary-500 font-medium">ผู้มีสิทธิ์เลือกตั้ง</p>
          <p className="text-sm font-semibold text-primary-900 truncate">{voterInfo?.area_name}</p>
        </div>
      </div>

      {/* Registered contact display */}
      {defaultContact && (
        <div className="bg-gray-50 border border-gray-200 rounded-xl p-3">
          <p className="text-xs text-gray-500 font-medium mb-0.5 flex items-center gap-1"><ChannelIcon/>{label} ที่ลงทะเบียนไว้</p>
          <p className="font-mono font-bold text-gray-800 text-sm">{defaultContact}</p>
        </div>
      )}

      {/* Custom address input */}
      <div>
        <label className="block text-xs font-semibold text-gray-500 uppercase tracking-wider mb-1.5">
          {defaultContact ? `ส่งไปที่ ${label} อื่น (ถ้าต้องการ)` : `ใส่ ${label} สำหรับรับ OTP`}
        </label>
        <input
          type={isEmail ? 'email' : 'tel'}
          value={address}
          onChange={e => { setAddress(e.target.value); setError('') }}
          placeholder={defaultContact || (isEmail ? 'your@email.com' : '08xxxxxxxx')}
          className="input-field"
          autoFocus={!defaultContact}
        />
        {!defaultContact && (
          <p className="text-xs text-amber-600 mt-1 flex items-center gap-1"><IconWarn/>ไม่พบ {label} ในระบบ กรุณาใส่เอง</p>
        )}
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-xl p-3 text-sm text-red-600 flex gap-2 animate-fade-in">
          <IconWarn/><span>{error}</span>
        </div>
      )}

      <button onClick={submit} disabled={loading || (!defaultContact && !address.trim())}
        className="btn-primary flex items-center justify-center gap-2">
        {loading ? <><Spinner /><span>กำลังส่ง OTP…</span></> : `ส่ง OTP ทาง ${label}`}
      </button>

      <button type="button" onClick={onBack}
        className="w-full text-sm text-gray-400 hover:text-gray-600 transition-colors text-center">
        ← ย้อนกลับ
      </button>
    </div>
  )
}

// ─── Step 3 — Enter OTP ───────────────────────────────────────────────────────
function StepOTP({ voterInfo, channel, refCode, sentTo, onSuccess, onBack }) {
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

  const ChannelIcon = channel === 'sms' ? IconPhone : IconEmail

  return (
    <form onSubmit={submit} className="space-y-5 animate-slide-up">
      {/* Sent-to banner */}
      <div className="flex items-start gap-3 bg-emerald-50 border border-emerald-100 rounded-xl p-3">
        <span className="text-emerald-600 mt-0.5 flex-shrink-0"><ChannelIcon/></span>
        <div>
          <p className="text-sm font-semibold text-emerald-800">ส่ง OTP ไปแล้ว</p>
          <p className="text-xs text-emerald-600 mt-0.5 font-mono font-bold">{sentTo}</p>
          <p className="text-xs text-emerald-500 mt-0.5">
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
          <IconWarn/><span>{error}</span>
        </div>
      )}

      <button type="submit" disabled={otp.length !== 6 || loading}
        className="btn-primary flex items-center justify-center gap-2">
        {loading ? <><Spinner /><span>กำลังยืนยัน…</span></> : 'ยืนยัน OTP'}
      </button>

      <div className="flex justify-between items-center">
        <button type="button" onClick={onBack}
          className="text-sm text-gray-400 hover:text-gray-600 transition-colors">← ย้อนกลับ</button>
        <button type="button" disabled={cooldown > 0} onClick={() => setCooldown(60)}
          className="text-sm text-primary-600 hover:text-primary-800 disabled:text-gray-300 disabled:cursor-not-allowed transition-colors">
          {cooldown > 0 ? `ส่งใหม่ได้ใน ${cooldown}s` : 'ส่ง OTP ใหม่'}
        </button>
      </div>
    </form>
  )
}

// ─── Step indicator (3 steps) ─────────────────────────────────────────────────
function Steps({ current }) {
  const steps = ['ยืนยันตัวตน', 'ที่อยู่รับ OTP', 'รหัส OTP']
  return (
    <div className="flex items-center justify-center gap-1 mb-7">
      {steps.map((l, i) => {
        const n = i + 1, done = current > n, active = current === n
        return (
          <div key={i} className="flex items-center gap-1">
            <div className="flex flex-col items-center gap-1">
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-sm font-bold transition-all duration-300 ${
                done   ? 'bg-emerald-500 text-white' :
                active ? 'bg-primary-700 text-white ring-4 ring-primary-100' :
                         'bg-gray-100 text-gray-400'}`}>
                {done
                  ? <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd"/>
                    </svg>
                  : n}
              </div>
              <span className={`text-xs font-medium whitespace-nowrap ${active ? 'text-primary-700' : done ? 'text-emerald-600' : 'text-gray-400'}`}>{l}</span>
            </div>
            {i < steps.length - 1 && (
              <div className={`w-8 h-0.5 mb-4 rounded-full transition-colors ${current > n ? 'bg-emerald-400' : 'bg-gray-200'}`}/>
            )}
          </div>
        )
      })}
    </div>
  )
}

// ─── Shell wrapper ────────────────────────────────────────────────────────────
function Shell({ title, subtitle, wide, ultrawide, onClose, children }) {
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-primary-950 to-blue-950 flex flex-col items-center justify-center p-4">
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/4 -right-32 w-80 h-80 bg-primary-500/10 rounded-full blur-3xl"/>
        <div className="absolute bottom-1/4 -left-32 w-80 h-80 bg-blue-500/10 rounded-full blur-3xl"/>
      </div>

      <div className={`relative w-full ${ultrawide ? 'max-w-6xl' : wide ? 'max-w-xl' : 'max-w-md'}`}>
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
            <div className="bg-gradient-to-r from-primary-800 to-primary-700 px-7 py-5 flex items-start justify-between">
              <div>
                <h2 className="text-white font-bold text-lg">{title}</h2>
                {subtitle && <p className="text-primary-200 text-sm mt-0.5">{subtitle}</p>}
              </div>
              {onClose && (
                <button onClick={onClose}
                  className="text-primary-300 hover:text-white transition-colors ml-4 mt-0.5 flex-shrink-0 p-0.5 rounded-lg hover:bg-white/10">
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth={2.5} viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12"/>
                  </svg>
                </button>
              )}
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
  const [step, setStep]           = useState(1)   // 1 | 2 | 3
  const [loginData, setLoginData] = useState({})  // { voterInfo, voterId, channel, refCode, sentTo }
  const [auth, setAuth]           = useState(null)
  const [screen, setScreen]       = useState('login')
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
    <Shell title="ผลโหวต Realtime" subtitle="VoteSpher 2026" ultrawide onClose={reset}>
      <ResultsPage onLogout={auth ? reset : null} justVoted={justVoted}/>
    </Shell>
  )

  if (screen === 'admin' && auth) return (
    <Shell title="Admin Dashboard" subtitle="VoteSpher 2026">
      <AdminPage token={auth.token} onLogout={reset}/>
    </Shell>
  )

  // ── Login flow (3 steps) ──
  return (
    <Shell>
      <Steps current={step}/>

      {step === 1 && (
        <>
          <StepVerify onSuccess={({ voterInfo, voterId, channel }) => {
            setLoginData({ voterInfo, voterId, channel })
            setStep(2)
          }}/>
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
        <StepDelivery
          voterInfo={loginData.voterInfo}
          channel={loginData.channel}
          voterId={loginData.voterId}
          onSuccess={({ refCode, sentTo }) => {
            setLoginData(d => ({ ...d, refCode, sentTo }))
            setStep(3)
          }}
          onBack={() => { setStep(1); setLoginData({}) }}
        />
      )}

      {step === 3 && (
        <StepOTP
          voterInfo={loginData.voterInfo}
          channel={loginData.channel}
          refCode={loginData.refCode}
          sentTo={loginData.sentTo}
          onSuccess={afterLogin}
          onBack={() => setStep(2)}
        />
      )}
    </Shell>
  )
}
