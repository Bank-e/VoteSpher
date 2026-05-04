import { useState, useRef, useEffect } from 'react'

// ─── Icons ───────────────────────────────────────────────────────────────────

const BallotIcon = () => (
  <svg viewBox="0 0 24 24" fill="none" className="w-8 h-8 text-white" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 002.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 00-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 00.75-.75 2.25 2.25 0 00-.1-.664m-5.8 0A2.251 2.251 0 0113.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25zM6.75 12h.008v.008H6.75V12zm0 3h.008v.008H6.75V15zm0 3h.008v.008H6.75V18z" />
  </svg>
)

const CheckIcon = () => (
  <svg viewBox="0 0 24 24" fill="none" className="w-16 h-16 text-emerald-500" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
  </svg>
)

const MailIcon = () => (
  <svg viewBox="0 0 24 24" fill="none" className="w-5 h-5" stroke="currentColor" strokeWidth={1.5}>
    <path strokeLinecap="round" strokeLinejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" />
  </svg>
)

const SpinnerIcon = () => (
  <svg className="animate-spin w-5 h-5" viewBox="0 0 24 24" fill="none">
    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
  </svg>
)

// ─── API ─────────────────────────────────────────────────────────────────────

const API_BASE = import.meta.env.VITE_API_URL || ''

async function apiPost(path, body) {
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  const data = await res.json()
  if (!res.ok) throw new Error(data.error || 'เกิดข้อผิดพลาด กรุณาลองใหม่')
  return data
}

// ─── Step Indicator ──────────────────────────────────────────────────────────

function StepIndicator({ current }) {
  const steps = ['ยืนยันตัวตน', 'รหัส OTP', 'สำเร็จ']
  return (
    <div className="flex items-center justify-center gap-2 mb-8">
      {steps.map((label, i) => {
        const idx = i + 1
        const done = current > idx
        const active = current === idx
        return (
          <div key={i} className="flex items-center gap-2">
            <div className="flex flex-col items-center gap-1">
              <div className={`step-dot ${
                done    ? 'bg-emerald-500 text-white shadow-emerald-200 shadow-md' :
                active  ? 'bg-primary-700 text-white shadow-primary-200 shadow-md ring-4 ring-primary-100' :
                          'bg-gray-100 text-gray-400'
              }`}>
                {done ? (
                  <svg viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                ) : idx}
              </div>
              <span className={`text-xs font-medium whitespace-nowrap ${active ? 'text-primary-700' : done ? 'text-emerald-600' : 'text-gray-400'}`}>
                {label}
              </span>
            </div>
            {i < steps.length - 1 && (
              <div className={`w-12 h-0.5 mb-4 rounded-full transition-colors duration-300 ${current > idx ? 'bg-emerald-400' : 'bg-gray-200'}`} />
            )}
          </div>
        )
      })}
    </div>
  )
}

// ─── OTP Input ───────────────────────────────────────────────────────────────

function OTPInput({ value, onChange }) {
  const inputs = useRef([])
  const digits = value.split('')

  const handleKey = (i, e) => {
    if (e.key === 'Backspace') {
      e.preventDefault()
      const next = [...digits]
      if (next[i]) {
        next[i] = ''
        onChange(next.join(''))
      } else if (i > 0) {
        next[i - 1] = ''
        onChange(next.join(''))
        inputs.current[i - 1]?.focus()
      }
    }
  }

  const handleChange = (i, e) => {
    const char = e.target.value.replace(/\D/g, '').slice(-1)
    const next = [...digits]
    next[i] = char
    onChange(next.join(''))
    if (char && i < 5) inputs.current[i + 1]?.focus()
  }

  const handlePaste = (e) => {
    e.preventDefault()
    const pasted = e.clipboardData.getData('text').replace(/\D/g, '').slice(0, 6)
    onChange(pasted.padEnd(6, '').slice(0, 6))
    if (pasted.length > 0) inputs.current[Math.min(pasted.length, 5)]?.focus()
  }

  return (
    <div className="flex gap-2 justify-center">
      {[0, 1, 2, 3, 4, 5].map(i => (
        <input
          key={i}
          ref={el => (inputs.current[i] = el)}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={digits[i] || ''}
          onChange={e => handleChange(i, e)}
          onKeyDown={e => handleKey(i, e)}
          onPaste={handlePaste}
          onFocus={e => e.target.select()}
          className={`otp-input ${digits[i] ? 'border-primary-500 bg-primary-50' : ''}`}
        />
      ))}
    </div>
  )
}

// ─── Step 1: Verify ───────────────────────────────────────────────────────────

function StepVerify({ onSuccess }) {
  const [citizenId, setCitizenId] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const isValid = /^\d{13}$/.test(citizenId)

  const handleFormat = (val) => {
    setCitizenId(val.replace(/\D/g, '').slice(0, 13))
    setError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (!isValid) return
    setLoading(true)
    setError('')
    try {
      const verifyData = await apiPost('/voter/verify', { citizen_id: citizenId })
      const otpData = await apiPost('/voter/otp-request', { voter_id: verifyData.voter_id })
      onSuccess({ voterInfo: verifyData.voter_info, refCode: otpData.ref_code })
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  const formatted = citizenId.replace(/(\d{1})(\d{4})(\d{5})(\d{2})(\d{1})/, '$1-$2-$3-$4-$5')

  return (
    <form onSubmit={handleSubmit} className="animate-slide-up space-y-5">
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-2">
          เลขบัตรประชาชน
        </label>
        <div className="relative">
          <input
            type="text"
            inputMode="numeric"
            placeholder="X-XXXX-XXXXX-XX-X"
            value={formatted}
            onChange={e => handleFormat(e.target.value.replace(/\D/g, ''))}
            className="input-field font-mono text-lg tracking-widest pr-12"
            autoFocus
          />
          {citizenId.length > 0 && (
            <div className="absolute right-3 top-1/2 -translate-y-1/2">
              {isValid ? (
                <span className="text-emerald-500 text-xl">✓</span>
              ) : (
                <span className="text-xs text-gray-400">{citizenId.length}/13</span>
              )}
            </div>
          )}
        </div>
        <p className="mt-1.5 text-xs text-gray-400">กรอกเลข 13 หลักโดยไม่ต้องใส่ขีด</p>
      </div>

      {error && (
        <div className="flex items-start gap-2.5 bg-red-50 border border-red-200 rounded-xl px-4 py-3 animate-fade-in">
          <span className="text-red-500 mt-0.5 text-base flex-shrink-0">⚠</span>
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      <button type="submit" disabled={!isValid || loading} className="btn-primary flex items-center justify-center gap-2.5">
        {loading ? (
          <>
            <SpinnerIcon />
            <span>กำลังตรวจสอบ…</span>
          </>
        ) : (
          <>
            <span>ยืนยันตัวตน</span>
            <svg viewBox="0 0 20 20" fill="currentColor" className="w-4 h-4">
              <path fillRule="evenodd" d="M3 10a.75.75 0 01.75-.75h10.638L10.23 5.29a.75.75 0 111.04-1.08l5.5 5.25a.75.75 0 010 1.08l-5.5 5.25a.75.75 0 11-1.04-1.08l4.158-3.96H3.75A.75.75 0 013 10z" clipRule="evenodd" />
            </svg>
          </>
        )}
      </button>
    </form>
  )
}

// ─── Step 2: OTP ──────────────────────────────────────────────────────────────

function StepOTP({ voterInfo, refCode, onSuccess, onBack }) {
  const [otp, setOtp] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [resending, setResending] = useState(false)
  const [resendCooldown, setResendCooldown] = useState(60)

  useEffect(() => {
    if (resendCooldown > 0) {
      const t = setTimeout(() => setResendCooldown(c => c - 1), 1000)
      return () => clearTimeout(t)
    }
  }, [resendCooldown])

  const handleSubmit = async (e) => {
    e.preventDefault()
    if (otp.length !== 6) return
    setLoading(true)
    setError('')
    try {
      const data = await apiPost('/voter/otp-confirm', { otp_code: otp, ref_code: refCode })
      onSuccess({ token: data.token, voterInfo })
    } catch (err) {
      setError(err.message)
      setOtp('')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="animate-slide-up space-y-5">

      {/* Voter info badge */}
      <div className="bg-primary-50 border border-primary-100 rounded-xl px-4 py-3 flex items-center gap-3">
        <div className="w-9 h-9 rounded-full bg-primary-700 flex items-center justify-center flex-shrink-0">
          <span className="text-white text-sm font-bold">
            {voterInfo?.area_id ?? '?'}
          </span>
        </div>
        <div>
          <p className="text-xs text-primary-600 font-medium">ผู้มีสิทธิ์เลือกตั้ง</p>
          <p className="text-sm font-semibold text-primary-900">{voterInfo?.area_name || 'เขตที่ ' + voterInfo?.area_id}</p>
        </div>
        {voterInfo?.is_voted && (
          <span className="ml-auto text-xs bg-amber-100 text-amber-700 px-2 py-0.5 rounded-full font-medium">
            โหวตแล้ว
          </span>
        )}
      </div>

      {/* Email notice */}
      <div className="flex items-start gap-2.5 bg-blue-50 border border-blue-100 rounded-xl px-4 py-3">
        <span className="text-blue-500 mt-0.5 flex-shrink-0"><MailIcon /></span>
        <div>
          <p className="text-sm text-blue-700 font-medium">ส่ง OTP ไปยัง email ของคุณแล้ว</p>
          <p className="text-xs text-blue-500 mt-0.5">Ref Code: <span className="font-mono font-bold tracking-wider">{refCode}</span></p>
        </div>
      </div>

      {/* OTP input */}
      <div>
        <label className="block text-sm font-semibold text-gray-700 mb-3 text-center">
          กรอกรหัส OTP 6 หลัก
        </label>
        <OTPInput value={otp} onChange={val => { setOtp(val); setError('') }} />
        <p className="text-xs text-gray-400 text-center mt-2.5">รหัสมีอายุ 5 นาที</p>
      </div>

      {error && (
        <div className="flex items-start gap-2.5 bg-red-50 border border-red-200 rounded-xl px-4 py-3 animate-fade-in">
          <span className="text-red-500 mt-0.5 flex-shrink-0">⚠</span>
          <p className="text-sm text-red-600">{error}</p>
        </div>
      )}

      <button type="submit" disabled={otp.length !== 6 || loading} className="btn-primary flex items-center justify-center gap-2.5">
        {loading ? (
          <>
            <SpinnerIcon />
            <span>กำลังยืนยัน…</span>
          </>
        ) : (
          'ยืนยัน OTP'
        )}
      </button>

      {/* Resend + back */}
      <div className="flex items-center justify-between pt-1">
        <button type="button" onClick={onBack} className="btn-ghost">
          ← ย้อนกลับ
        </button>
        <button
          type="button"
          disabled={resendCooldown > 0 || resending}
          className="btn-ghost disabled:opacity-40"
          onClick={async () => {
            setResending(true)
            setResendCooldown(60)
            setResending(false)
          }}
        >
          {resendCooldown > 0 ? `ส่งใหม่ (${resendCooldown}s)` : 'ส่ง OTP ใหม่'}
        </button>
      </div>
    </form>
  )
}

// ─── Step 3: Success ──────────────────────────────────────────────────────────

function StepSuccess({ voterInfo, token, onReset }) {
  const [copied, setCopied] = useState(false)

  const copy = () => {
    navigator.clipboard.writeText(token)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="animate-slide-up text-center space-y-5">
      {/* Success icon */}
      <div className="flex justify-center">
        <div className="relative">
          <div className="absolute inset-0 bg-emerald-100 rounded-full animate-pulse-slow" />
          <CheckIcon />
        </div>
      </div>

      <div>
        <h2 className="text-xl font-bold text-gray-800">เข้าสู่ระบบสำเร็จ!</h2>
        <p className="text-sm text-gray-500 mt-1">ยืนยันตัวตนเรียบร้อยแล้ว</p>
      </div>

      {/* Voter info */}
      <div className="bg-gray-50 rounded-xl p-4 text-left space-y-2.5">
        <Row label="เขตเลือกตั้ง" value={voterInfo?.area_name || 'เขตที่ ' + voterInfo?.area_id} />
        <Row label="สถานะ" value={voterInfo?.is_voted ? '✓ ลงคะแนนแล้ว' : 'ยังไม่ได้ลงคะแนน'} highlight={!voterInfo?.is_voted} />
      </div>

      {/* Token */}
      <div className="bg-gray-900 rounded-xl p-3 text-left">
        <div className="flex items-center justify-between mb-1.5">
          <span className="text-xs text-gray-400 font-mono">access_token</span>
          <button onClick={copy} className="text-xs text-primary-400 hover:text-primary-300 transition-colors">
            {copied ? '✓ คัดลอกแล้ว' : 'คัดลอก'}
          </button>
        </div>
        <p className="text-xs text-emerald-400 font-mono break-all leading-relaxed line-clamp-3">
          {token}
        </p>
      </div>

      <button onClick={onReset} className="btn-primary">
        เริ่มใหม่
      </button>
    </div>
  )
}

function Row({ label, value, highlight }) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm text-gray-500">{label}</span>
      <span className={`text-sm font-semibold ${highlight ? 'text-primary-700' : 'text-gray-800'}`}>{value}</span>
    </div>
  )
}

// ─── App ──────────────────────────────────────────────────────────────────────

export default function App() {
  const [step, setStep] = useState(1)
  const [state, setState] = useState({})

  const reset = () => { setStep(1); setState({}) }

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-900 via-primary-800 to-blue-900 flex flex-col items-center justify-center p-4">

      {/* Background decoration */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute -top-40 -right-40 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl" />
        <div className="absolute -bottom-40 -left-40 w-96 h-96 bg-primary-500/10 rounded-full blur-3xl" />
      </div>

      {/* Card */}
      <div className="relative w-full max-w-md">

        {/* Header */}
        <div className="text-center mb-6">
          <div className="inline-flex items-center gap-3 bg-white/10 backdrop-blur-sm px-5 py-2.5 rounded-2xl mb-4">
            <BallotIcon />
            <div className="text-left">
              <h1 className="text-white font-bold text-xl tracking-wide">VoteSpher</h1>
              <p className="text-blue-200 text-xs">ระบบการเลือกตั้งออนไลน์</p>
            </div>
          </div>
        </div>

        {/* Main card */}
        <div className="bg-white rounded-2xl shadow-2xl shadow-black/30 p-7">
          <StepIndicator current={step} />

          {step === 1 && (
            <>
              <div className="mb-6">
                <h2 className="text-lg font-bold text-gray-800">ยืนยันตัวตน</h2>
                <p className="text-sm text-gray-500 mt-1">กรอกเลขบัตรประชาชนเพื่อรับรหัส OTP</p>
              </div>
              <StepVerify onSuccess={({ voterInfo, refCode }) => {
                setState({ voterInfo, refCode })
                setStep(2)
              }} />
            </>
          )}

          {step === 2 && (
            <>
              <div className="mb-6">
                <h2 className="text-lg font-bold text-gray-800">ยืนยันรหัส OTP</h2>
                <p className="text-sm text-gray-500 mt-1">ตรวจสอบ email ของคุณแล้วกรอกรหัส</p>
              </div>
              <StepOTP
                voterInfo={state.voterInfo}
                refCode={state.refCode}
                onSuccess={({ token, voterInfo }) => {
                  setState(s => ({ ...s, token, voterInfo }))
                  setStep(3)
                }}
                onBack={() => setStep(1)}
              />
            </>
          )}

          {step === 3 && (
            <StepSuccess
              voterInfo={state.voterInfo}
              token={state.token}
              onReset={reset}
            />
          )}
        </div>

        {/* Footer */}
        <p className="text-center text-blue-300/60 text-xs mt-5">
          © 2026 VoteSpher — ข้อมูลทุกอย่างถูกเข้ารหัสและปลอดภัย
        </p>
      </div>
    </div>
  )
}
