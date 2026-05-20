const BASE = import.meta.env.VITE_API_URL || ''

async function req(path, opts = {}) {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...opts.headers },
    ...opts,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error(data.error || data.message || `HTTP ${res.status}`)
  return data
}

export const api = {
  // Auth
  verify:     (citizen_id)             => req('/voter/verify',      { method:'POST', body: JSON.stringify({ citizen_id }) }),
  otpRequest: (voter_id, delivery_channel, delivery_address) => req('/voter/otp-request', { method:'POST', body: JSON.stringify({ voter_id, delivery_channel, ...(delivery_address ? { delivery_address } : {}) }) }),
  otpConfirm: (otp_code, ref_code)     => req('/voter/otp-confirm', { method:'POST', body: JSON.stringify({ otp_code, ref_code }) }),
  me:         (token)                  => req('/voter/me',           { headers: { Authorization: `Bearer ${token}` } }),

  // Info
  candidates: (area_id)                => req(`/candidates?area_id=${area_id}`),
  parties:    ()                       => req('/parties'),

  // Ballot
  submit:     (body, token)            => req('/ballot/submit',  { method:'POST', headers:{ Authorization:`Bearer ${token}` }, body: JSON.stringify(body) }),
  status:     (token)                  => req('/ballot/status',  { headers: { Authorization: `Bearer ${token}` } }),

  // Results
  allAreas:   ()                       => req('/results/areas'),
  areaResult: (area_id)                => req(`/results/areas/${area_id}`),
  partyResult:(province, area_id)      => req(`/results/provinces/${encodeURIComponent(province)}/areas/${area_id}`),

  // Election
  getConfig:  ()                       => req('/election/config'),
  setConfig:  (body, token)            => req('/election/config', { method:'PATCH', headers:{ Authorization:`Bearer ${token}` }, body: JSON.stringify(body) }),
}
