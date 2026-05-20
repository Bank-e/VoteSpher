import { useRef, useEffect } from 'react'
import * as THREE from 'three'
import { OrbitControls } from 'three/addons/controls/OrbitControls.js'

// ─── Geographic grid [row, col] for Bangkok districts ────────────────────────
const DISTRICT_GRID = {
  17:[0,3], 18:[0,4], 19:[0,5],
  15:[1,2], 14:[1,3], 20:[1,4], 22:[1,5], 25:[1,6],
   1:[2,1], 13:[2,2], 12:[2,3], 21:[2,4], 23:[2,5], 24:[2,6],
   2:[3,0],  3:[3,1], 10:[3,2], 11:[3,3], 31:[3,4], 30:[3,5], 26:[3,6],
   4:[4,0],  5:[4,1], 32:[4,2], 29:[4,3], 28:[4,4], 27:[4,5],
   9:[5,1],  8:[5,2], 33:[5,3],  7:[5,4],
   6:[6,2], 16:[6,3],
}

const LANDMARK_BY_NO = {
   1:'TEMPLE',  2:'DOME',    3:'TEMPLE',  4:'TEMPLE',   5:'TOWER',
   6:'MARKET',  7:'HOUSING', 8:'MARKET',  9:'HOUSING',
  10:'TOWER',  11:'TOWER',  12:'TOWER',  13:'MALL',    14:'MARKET',
  15:'DOME',   16:'DOME',   17:'AIRPORT',18:'HOUSING', 19:'HOUSING',
  20:'MALL',   21:'MALL',   22:'HOUSING',23:'BRIDGE',  24:'HOUSING',
  25:'HOUSING',26:'HOUSING',27:'AIRPORT',28:'HOUSING', 29:'MALL',
  30:'MALL',   31:'TOWER',  32:'MARKET', 33:'HOUSING',
}

const PARTY_COLOR = {
  'พรรคประชาชน':   0xf97316,
  'พรรคเพื่อไทย':  0xe11d48,
  'พรรคภูมิใจไทย': 0x10b981,
  'พรรคกล้าธรรม':  0x8b5cf6,
}
const DEFAULT_COLOR = 0x64748b

const BKK_BOUNDARY = [
  [-2.8,-6.0],[-3.8,-5.0],[-4.6,-3.5],[-5.0,-1.5],
  [-4.8, 0.5],[-4.0, 2.5],[-2.8, 4.2],[-1.0, 5.5],
  [ 0.5, 6.3],[ 2.0, 5.4],[ 3.5, 3.8],[ 4.5, 2.0],
  [ 5.0, 0.0],[ 4.8,-2.0],[ 4.2,-3.8],[ 3.0,-5.2],
  [ 1.0,-6.0],[-1.0,-6.2],
]

// ─── Geometry helpers ─────────────────────────────────────────────────────────
function computeVoronoiCells(seeds, boundary) {
  return seeds.map(si => {
    let cell = boundary.map(p => [p[0], p[1]])
    seeds.forEach(sj => {
      if (sj === si || cell.length < 3) return
      const mx = (si.x + sj.x) / 2, my = (si.z + sj.z) / 2
      const nx = si.x - sj.x, ny = si.z - sj.z
      cell = clipHalfPlane(cell, mx, my, nx, ny)
    })
    return cell
  })
}
function clipHalfPlane(poly, mx, my, nx, ny) {
  const inside = p => nx * (p[0] - mx) + ny * (p[1] - my) >= 0
  const out = []
  for (let i = 0; i < poly.length; i++) {
    const p = poly[i], q = poly[(i + 1) % poly.length]
    const pIn = inside(p), qIn = inside(q)
    if (pIn && qIn) out.push(q)
    else if (pIn && !qIn) out.push(intersectHP(p, q, mx, my, nx, ny))
    else if (!pIn && qIn) { out.push(intersectHP(p, q, mx, my, nx, ny)); out.push(q) }
  }
  return out
}
function intersectHP(p, q, mx, my, nx, ny) {
  const dx = q[0]-p[0], dy = q[1]-p[1]
  const den = nx*dx + ny*dy
  if (Math.abs(den) < 1e-10) return [p[0], p[1]]
  const t = (nx*(mx-p[0]) + ny*(my-p[1])) / den
  return [p[0]+t*dx, p[1]+t*dy]
}
function insetPolygon(ring, center, amount) {
  return ring.map(([x, y]) => {
    const dx = center[0]-x, dy = center[1]-y
    const d = Math.hypot(dx, dy)
    if (d < 0.001) return [x, y]
    const t = Math.min(amount/d, 0.25)
    return [x+dx*t, y+dy*t]
  })
}
function pointInPoly(p, poly) {
  let inside = false
  for (let i = 0, j = poly.length-1; i < poly.length; j = i++) {
    const xi = poly[i][0], yi = poly[i][1], xj = poly[j][0], yj = poly[j][1]
    if (((yi > p[1]) !== (yj > p[1])) && (p[0] < (xj-xi)*(p[1]-yi)/(yj-yi+1e-12)+xi))
      inside = !inside
  }
  return inside
}
function mulberry32(seed) {
  return function() {
    seed |= 0; seed = (seed + 0x6D2B79F5) | 0
    let t = seed
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// ─── Landmark builder ─────────────────────────────────────────────────────────
function makeLandmark(type, accent) {
  const g = new THREE.Group()
  const am = (i=0.2) => new THREE.MeshStandardMaterial({ color:accent, metalness:0.45, roughness:0.4, emissive:accent, emissiveIntensity:i })
  switch (type) {
    case 'TEMPLE': {
      const base = new THREE.Mesh(new THREE.BoxGeometry(0.32,0.1,0.32), new THREE.MeshStandardMaterial({color:0xf3e8d6,roughness:0.7}))
      base.position.y = 0.05; g.add(base)
      const cone = new THREE.Mesh(new THREE.ConeGeometry(0.13,0.5,8), new THREE.MeshStandardMaterial({color:0xfacc15,metalness:0.55,roughness:0.35,emissive:0xfacc15,emissiveIntensity:0.18}))
      cone.position.y = 0.35; g.add(cone)
      const ring = new THREE.Mesh(new THREE.TorusGeometry(0.11,0.025,8,16), am(0.35))
      ring.rotation.x = Math.PI/2; ring.position.y = 0.13; g.add(ring)
      const tip = new THREE.Mesh(new THREE.SphereGeometry(0.028,6,6), new THREE.MeshStandardMaterial({color:0xfde047,metalness:0.85,roughness:0.15}))
      tip.position.y = 0.63; g.add(tip); break
    }
    case 'TOWER': {
      const t = new THREE.Mesh(new THREE.BoxGeometry(0.18,0.85,0.18), new THREE.MeshStandardMaterial({color:0xf3f4f6,roughness:0.3,metalness:0.7,emissive:0xbfdbfe,emissiveIntensity:0.08}))
      t.position.y = 0.425; g.add(t)
      const band = new THREE.Mesh(new THREE.BoxGeometry(0.21,0.1,0.21), am(0.4))
      band.position.y = 0.55; g.add(band)
      const ant = new THREE.Mesh(new THREE.CylinderGeometry(0.014,0.014,0.22,4), new THREE.MeshStandardMaterial({color:0xfecaca,emissive:0xef4444,emissiveIntensity:0.4}))
      ant.position.y = 0.94; g.add(ant)
      const s2 = new THREE.Mesh(new THREE.BoxGeometry(0.11,0.5,0.11), new THREE.MeshStandardMaterial({color:0xd1d5db,roughness:0.4,metalness:0.5}))
      s2.position.set(0.18,0.25,0.05); g.add(s2); break
    }
    case 'MALL': {
      const box = new THREE.Mesh(new THREE.BoxGeometry(0.42,0.22,0.32), new THREE.MeshStandardMaterial({color:0xe5e7eb,roughness:0.5,metalness:0.15}))
      box.position.y = 0.11; g.add(box)
      const sign = new THREE.Mesh(new THREE.BoxGeometry(0.24,0.08,0.04), am(0.6))
      sign.position.set(0,0.28,0.16); g.add(sign); break
    }
    case 'AIRPORT': {
      const rw = new THREE.Mesh(new THREE.BoxGeometry(0.72,0.04,0.13), new THREE.MeshStandardMaterial({color:0x2b2f3a,roughness:0.9}))
      rw.position.y = 0.02; g.add(rw)
      const st = new THREE.Mesh(new THREE.BoxGeometry(0.6,0.005,0.018), new THREE.MeshBasicMaterial({color:0xfde047}))
      st.position.y = 0.045; g.add(st)
      const dome = new THREE.Mesh(new THREE.SphereGeometry(0.14,12,6,0,Math.PI*2,0,Math.PI/2), am(0.2))
      dome.position.set(-0.22,0.04,0.17); g.add(dome)
      const ct = new THREE.Mesh(new THREE.CylinderGeometry(0.024,0.03,0.34,8), new THREE.MeshStandardMaterial({color:0xf3f4f6,metalness:0.4}))
      ct.position.set(0.14,0.17,0.17); g.add(ct); break
    }
    case 'DOME': {
      const d = new THREE.Mesh(new THREE.SphereGeometry(0.22,14,8,0,Math.PI*2,0,Math.PI/2), am(0.22))
      d.position.y = 0.08; g.add(d)
      const b = new THREE.Mesh(new THREE.CylinderGeometry(0.23,0.25,0.1,16), new THREE.MeshStandardMaterial({color:0xf3f4f6}))
      b.position.y = 0.05; g.add(b); break
    }
    case 'HOUSING': {
      const mw = new THREE.MeshStandardMaterial({color:0xc7c2b8,roughness:0.7})
      const sz=[0.2,0.26,0.16,0.18], ps=[[-0.10,-0.10],[0.10,-0.10],[-0.10,0.10],[0.10,0.10]]
      sz.forEach((h,i) => { const b=new THREE.Mesh(new THREE.BoxGeometry(0.14,h,0.14),mw); b.position.set(ps[i][0],h/2,ps[i][1]); g.add(b) })
      const roof = new THREE.Mesh(new THREE.ConeGeometry(0.12,0.1,4), am(0.18))
      roof.rotation.y = Math.PI/4; roof.position.set(ps[0][0],sz[0]+0.05,ps[0][1]); g.add(roof); break
    }
    case 'MARKET': {
      const m = new THREE.Mesh(new THREE.CylinderGeometry(0.22,0.24,0.1,6), new THREE.MeshStandardMaterial({color:0xfde68a,roughness:0.65}))
      m.position.y = 0.05; g.add(m)
      const tip = new THREE.Mesh(new THREE.ConeGeometry(0.22,0.22,6), am(0.18))
      tip.position.y = 0.21; g.add(tip); break
    }
    case 'BRIDGE': {
      const path = new THREE.CatmullRomCurve3([new THREE.Vector3(-0.28,0.02,0),new THREE.Vector3(0,0.24,0),new THREE.Vector3(0.28,0.02,0)])
      const m = new THREE.Mesh(new THREE.TubeGeometry(path,16,0.035,6,false), am(0.3)); g.add(m)
      const pm = new THREE.MeshStandardMaterial({color:0xf3f4f6})
      ;[-0.2,0.2].forEach(x => { const p=new THREE.Mesh(new THREE.BoxGeometry(0.035,0.2,0.035),pm); p.position.set(x,0.1,0); g.add(p) }); break
    }
  }
  g.traverse(o => { if (o.isMesh) { o.castShadow = true; o.receiveShadow = true } })
  return g
}

// ─── Filler builder ───────────────────────────────────────────────────────────
function makeFiller(rand, accent) {
  const g = new THREE.Group()
  const r = rand()
  if (r < 0.28) {
    const body = new THREE.Mesh(new THREE.BoxGeometry(0.16,0.11,0.14), new THREE.MeshStandardMaterial({color:0xe7e2d6,roughness:0.75}))
    body.position.y = 0.055; g.add(body)
    const roof = new THREE.Mesh(new THREE.ConeGeometry(0.13,0.09,4), new THREE.MeshStandardMaterial({color:accent,roughness:0.55,emissive:accent,emissiveIntensity:0.1}))
    roof.rotation.y = Math.PI/4; roof.position.y = 0.15; g.add(roof)
  } else if (r < 0.5) {
    const h=0.16+rand()*0.18, w=0.1+rand()*0.06
    const body=new THREE.Mesh(new THREE.BoxGeometry(w,h,w), new THREE.MeshStandardMaterial({color:0xb3b6be,roughness:0.5,metalness:0.35,emissive:accent,emissiveIntensity:0.08}))
    body.position.y = h/2; g.add(body)
  } else if (r < 0.72) {
    const body=new THREE.Mesh(new THREE.BoxGeometry(0.17,0.045,0.075), new THREE.MeshStandardMaterial({color:accent,metalness:0.6,roughness:0.28,emissive:accent,emissiveIntensity:0.18}))
    body.position.y = 0.025; g.add(body)
    const top=new THREE.Mesh(new THREE.BoxGeometry(0.1,0.035,0.065), new THREE.MeshStandardMaterial({color:0x1f2937,metalness:0.5,roughness:0.35}))
    top.position.y = 0.062; g.add(top)
  } else if (r < 0.88) {
    const trunk=new THREE.Mesh(new THREE.CylinderGeometry(0.018,0.022,0.08,5), new THREE.MeshStandardMaterial({color:0x6b4423}))
    trunk.position.y = 0.04; g.add(trunk)
    const leaves=new THREE.Mesh(new THREE.SphereGeometry(0.075,7,5), new THREE.MeshStandardMaterial({color:0x16a34a,roughness:0.85}))
    leaves.position.y = 0.13; g.add(leaves)
  } else {
    const pole=new THREE.Mesh(new THREE.CylinderGeometry(0.01,0.01,0.18,4), new THREE.MeshStandardMaterial({color:0x71717a}))
    pole.position.y = 0.09; g.add(pole)
    const lamp=new THREE.Mesh(new THREE.SphereGeometry(0.018,6,5), new THREE.MeshBasicMaterial({color:0xfef3c7}))
    lamp.position.y = 0.185; g.add(lamp)
  }
  g.rotation.y = rand() * Math.PI * 2
  g.traverse(o => { if (o.isMesh) { o.castShadow = true; o.receiveShadow = true } })
  return g
}

// ─── Main component ───────────────────────────────────────────────────────────
export default function Election3DScene({ areas, onSelectArea }) {
  const containerRef = useRef(null)
  const tooltipRef   = useRef(null)

  useEffect(() => {
    const container = containerRef.current
    if (!container || !areas || areas.length === 0) return

    const W = container.clientWidth || 640
    const H = container.clientHeight || 420

    // ── Renderer ──────────────────────────────────────────────────────────────
    const renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false })
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2))
    renderer.setSize(W, H)
    renderer.shadowMap.enabled = true
    renderer.shadowMap.type = THREE.PCFSoftShadowMap
    renderer.outputColorSpace = THREE.SRGBColorSpace
    renderer.toneMapping = THREE.ACESFilmicToneMapping
    renderer.toneMappingExposure = 1.05
    container.appendChild(renderer.domElement)

    // ── Scene ─────────────────────────────────────────────────────────────────
    const scene = new THREE.Scene()
    scene.fog = new THREE.Fog(0x0a1e3a, 30, 75)

    // ── Camera ────────────────────────────────────────────────────────────────
    const camera = new THREE.PerspectiveCamera(38, W/H, 0.1, 200)
    const HOME_POS = new THREE.Vector3(13, 14, 18)
    const HOME_TAR = new THREE.Vector3(0, 0.5, 0)
    camera.position.copy(HOME_POS)

    // ── Controls ──────────────────────────────────────────────────────────────
    const controls = new OrbitControls(camera, renderer.domElement)
    controls.enableDamping = true
    controls.dampingFactor = 0.06
    controls.minDistance = 8
    controls.maxDistance = 42
    controls.minPolarAngle = 0.15
    controls.maxPolarAngle = 1.35
    controls.target.copy(HOME_TAR)
    controls.autoRotate = true
    controls.autoRotateSpeed = 0.45

    // ── Lighting ──────────────────────────────────────────────────────────────
    scene.add(new THREE.HemisphereLight(0x9ec5ff, 0x1e1b4b, 0.55))
    scene.add(new THREE.AmbientLight(0xffffff, 0.22))
    const sun = new THREE.DirectionalLight(0xfff2d6, 1.6)
    sun.position.set(13, 22, 9)
    sun.castShadow = true
    sun.shadow.mapSize.set(2048, 2048)
    const SH = 16
    Object.assign(sun.shadow.camera, { left:-SH, right:SH, top:SH, bottom:-SH, near:1, far:60 })
    sun.shadow.bias = -0.0006
    sun.shadow.normalBias = 0.04
    scene.add(sun)
    const fill = new THREE.DirectionalLight(0x7e9cff, 0.55)
    fill.position.set(-12, 8, -10); scene.add(fill)

    // ── Ground ────────────────────────────────────────────────────────────────
    const ground = new THREE.Mesh(
      new THREE.CircleGeometry(24, 96),
      new THREE.MeshStandardMaterial({ color:0x081427, roughness:0.94 })
    )
    ground.rotation.x = -Math.PI/2; ground.position.y = -0.02
    ground.receiveShadow = true; scene.add(ground)

    // BKK boundary ring
    const bndPts = BKK_BOUNDARY.map(([x,z]) => new THREE.Vector3(x, 0.01, z))
    bndPts.push(bndPts[0].clone())
    scene.add(new THREE.Line(
      new THREE.BufferGeometry().setFromPoints(bndPts),
      new THREE.LineBasicMaterial({ color:0x60a5fa, transparent:true, opacity:0.55 })
    ))


    // ── Build districts ───────────────────────────────────────────────────────
    const maxVotes = Math.max(...areas.map(a => a.total_votes), 1)
    const seeds = areas.map(area => {
      const [row, col] = DISTRICT_GRID[area.area_id] || [3, 3]
      return {
        area,
        no:  area.area_id,
        x:   (col - 3) * 1.5 + Math.sin(area.area_id * 2.3) * 0.28,
        z:   (row - 3) * 1.5 + Math.cos(area.area_id * 1.7) * 0.28,
        winner:      area.candidates?.[0]?.party_name || null,
        winnerVotes: area.candidates?.[0]?.votes || 0,
        runnerVotes: area.candidates?.[1]?.votes || 0,
      }
    })

    const cells    = computeVoronoiCells(seeds, BKK_BOUNDARY)
    const districts = []

    seeds.forEach((s, i) => {
      let ring = cells[i]
      if (!ring || ring.length < 3) return

      ring = insetPolygon(ring, [s.x, s.z], 0.04)

      const color  = PARTY_COLOR[s.winner] ?? DEFAULT_COLOR
      const margin = s.winnerVotes > 0 ? (s.winnerVotes - s.runnerVotes) / s.winnerVotes : 0
      const height = 0.6 + (s.area.total_votes / maxVotes) * 1.5 + margin * 0.4

      const shape = new THREE.Shape()
      shape.moveTo(ring[0][0], -ring[0][1])
      for (let j = 1; j < ring.length; j++) shape.lineTo(ring[j][0], -ring[j][1])
      shape.closePath()

      const geom = new THREE.ExtrudeGeometry(shape, {
        depth:height, bevelEnabled:true,
        bevelSize:0.03, bevelThickness:0.025,
        bevelSegments:1, curveSegments:1,
      })
      geom.rotateX(-Math.PI/2)

      const tileMat = new THREE.MeshStandardMaterial({
        color, roughness:0.55, metalness:0.1,
        emissive:color, emissiveIntensity:0.07,
      })
      const tile = new THREE.Mesh(geom, tileMat)
      tile.castShadow = true; tile.receiveShadow = true

      const group = new THREE.Group()
      group.add(tile)

      const landmark = makeLandmark(LANDMARK_BY_NO[s.no] || 'HOUSING', color)
      landmark.position.set(s.x, height + 0.02, s.z)
      landmark.scale.setScalar(1.3)
      group.add(landmark)

      // City fillers
      const rand = mulberry32(s.no * 9973)
      const numF = 5 + Math.floor(rand() * 5)
      let minX=Infinity, minZ=Infinity, maxX=-Infinity, maxZ=-Infinity
      ring.forEach(p => { minX=Math.min(minX,p[0]); maxX=Math.max(maxX,p[0]); minZ=Math.min(minZ,p[1]); maxZ=Math.max(maxZ,p[1]) })
      let placed=0, attempts=0
      const fps=[]
      while (placed < numF && attempts < numF*25) {
        attempts++
        const fx=minX+rand()*(maxX-minX), fz=minZ+rand()*(maxZ-minZ)
        if (!pointInPoly([fx,fz],ring)) continue
        if (Math.hypot(fx-s.x,fz-s.z) < 0.42) continue
        if (fps.some(fp => Math.hypot(fx-fp[0],fz-fp[1]) < 0.22)) continue
        fps.push([fx,fz])
        const filler = makeFiller(rand, color)
        filler.position.set(fx, height+0.005, fz)
        filler.scale.setScalar(0.9+rand()*0.3)
        group.add(filler)
        placed++
      }

      group.userData = { area:s.area, tile, landmark, hovered:false, selected:false, dimAmount:0, height }
      tile.userData.group = group
      landmark.traverse(o => { if (o.isMesh) o.userData.group = group })

      districts.push(group)
      scene.add(group)
    })

    // ── Camera fly-to ─────────────────────────────────────────────────────────
    let cameraAnim = null
    function flyTo(toPos, toTar, dur = 900) {
      cameraAnim = {
        fromPos: camera.position.clone(), toPos: toPos.clone(),
        fromTar: controls.target.clone(), toTar: toTar.clone(),
        start: performance.now(), dur,
      }
    }

    // ── Interaction ───────────────────────────────────────────────────────────
    const raycaster = new THREE.Raycaster()
    const pointer   = new THREE.Vector2(-10, -10)
    const ptrLocal  = { x:0, y:0 }
    let hovered   = null
    let dragStart = null
    const tmpC    = new THREE.Color()

    function showTooltip(area, x, y) {
      const el = tooltipRef.current
      if (!el) return
      const winner = area.candidates?.[0]
      el.style.display = 'block'
      el.style.left = `${x}px`
      el.style.top  = `${y - 72}px`
      el.innerHTML = `
        <p style="font-weight:700;font-size:13px;color:#fff;margin:0 0 2px">${area.area_name}</p>
        <p style="font-size:11px;color:#93c5fd;margin:0 0 1px">${winner?.party_name || 'ยังไม่มีข้อมูล'}</p>
        <p style="font-size:11px;color:#9ca3af;margin:0">${area.total_votes.toLocaleString()} คะแนน</p>
      `
    }
    function hideTooltip() {
      const el = tooltipRef.current; if (el) el.style.display = 'none'
    }

    function setHover(g) {
      if (hovered === g) return
      if (hovered) hovered.userData.hovered = false
      hovered = g
      if (hovered) hovered.userData.hovered = true
      renderer.domElement.style.cursor = hovered ? 'pointer' : 'default'
      if (hovered) showTooltip(hovered.userData.area, ptrLocal.x, ptrLocal.y)
      else hideTooltip()
    }

    function onMouseMove(e) {
      const rect = renderer.domElement.getBoundingClientRect()
      ptrLocal.x = e.clientX - rect.left
      ptrLocal.y = e.clientY - rect.top
      pointer.x =  (ptrLocal.x / rect.width)  * 2 - 1
      pointer.y = -(ptrLocal.y / rect.height)  * 2 + 1
    }
    function onMouseLeave() { pointer.set(-10,-10); setHover(null) }
    function onPointerDown(e) { dragStart = { x:e.clientX, y:e.clientY } }
    function onPointerUp(e) {
      if (!dragStart) return
      const moved = Math.hypot(e.clientX-dragStart.x, e.clientY-dragStart.y) > 5
      dragStart = null
      if (!moved) {
        onMouseMove(e)
        raycaster.setFromCamera(pointer, camera)
        const hit = raycaster.intersectObjects(districts, true)[0]
        const g = hit?.object.userData.group || null

        // deselect old
        const wasSelected = g?.userData.selected
        districts.forEach(d => { d.userData.selected = false })

        if (g && !wasSelected) {
          g.userData.selected = true
          // fly camera toward clicked district
          const s = seeds.find(s => s.no === g.userData.area.area_id)
          if (s) {
            const tar = new THREE.Vector3(s.x, 0.5, s.z)
            flyTo(tar.clone().add(new THREE.Vector3(4, 6, 5)), tar, 900)
          }
        } else {
          // deselected — fly back home
          flyTo(HOME_POS, HOME_TAR, 900)
        }
      }
    }

    renderer.domElement.addEventListener('mousemove', onMouseMove)
    renderer.domElement.addEventListener('mouseleave', onMouseLeave)
    renderer.domElement.addEventListener('pointerdown', onPointerDown)
    renderer.domElement.addEventListener('pointerup', onPointerUp)

    // ── Animation loop ────────────────────────────────────────────────────────
    let animId
    function animate() {
      animId = requestAnimationFrame(animate)

      raycaster.setFromCamera(pointer, camera)
      setHover(raycaster.intersectObjects(districts, true)[0]?.object.userData.group || null)

      controls.autoRotate = !hovered && !districts.some(d => d.userData.selected)

      districts.forEach(g => {
        const u = g.userData
        const lift = (u.hovered ? 0.25 : 0) + (u.selected ? 0.4 : 0)
        g.position.y += (lift - g.position.y) * 0.18

        const em = u.selected ? 0.5 : (u.hovered ? 0.28 : 0.07)
        u.tile.material.emissiveIntensity += (em - u.tile.material.emissiveIntensity) * 0.18

        if (!u.tile.material._orig) u.tile.material._orig = u.tile.material.color.clone()
        tmpC.copy(u.tile.material._orig).lerp(new THREE.Color(0x0c1226), u.dimAmount)
        u.tile.material.color.copy(tmpC)
        u.landmark.visible = u.dimAmount < 0.6
      })

      if (cameraAnim) {
        const t = Math.min(1, (performance.now() - cameraAnim.start) / cameraAnim.dur)
        const e = 1 - Math.pow(1 - t, 3)
        camera.position.lerpVectors(cameraAnim.fromPos, cameraAnim.toPos, e)
        controls.target.lerpVectors(cameraAnim.fromTar, cameraAnim.toTar, e)
        if (t >= 1) cameraAnim = null
      }

      controls.update()
      renderer.render(scene, camera)
    }
    animate()

    // ── Responsive resize ─────────────────────────────────────────────────────
    const observer = new ResizeObserver(() => {
      const nw = container.clientWidth
      const nh = container.clientHeight || 420
      camera.aspect = nw/nh
      camera.updateProjectionMatrix()
      renderer.setSize(nw, nh)
    })
    observer.observe(container)

    // ── Cleanup ───────────────────────────────────────────────────────────────
    return () => {
      cancelAnimationFrame(animId)
      observer.disconnect()
      renderer.domElement.removeEventListener('mousemove', onMouseMove)
      renderer.domElement.removeEventListener('mouseleave', onMouseLeave)
      renderer.domElement.removeEventListener('pointerdown', onPointerDown)
      renderer.domElement.removeEventListener('pointerup', onPointerUp)
      controls.dispose()
      districts.forEach(g => g.traverse(o => {
        if (!o.isMesh) return
        o.geometry.dispose()
        ;(Array.isArray(o.material) ? o.material : [o.material]).forEach(m => m.dispose())
      }))
      renderer.dispose()
      if (container.contains(renderer.domElement)) container.removeChild(renderer.domElement)
      hideTooltip()
    }
  }, [areas])

  return (
    <div className="relative rounded-2xl overflow-hidden" style={{ height:'420px', background:'#081427' }}>
      <div ref={containerRef} style={{ width:'100%', height:'100%' }}/>
      {/* Tooltip */}
      <div ref={tooltipRef} className="absolute pointer-events-none z-10"
        style={{
          display: 'none',
          background: 'rgba(8,20,39,0.92)',
          border: '1px solid rgba(96,165,250,0.35)',
          borderRadius: '10px',
          padding: '8px 14px',
          transform: 'translateX(-50%)',
          backdropFilter: 'blur(10px)',
          whiteSpace: 'nowrap',
          boxShadow: '0 4px 20px rgba(0,0,0,0.5)',
        }}
      />
      {/* Corner hint */}
      <div className="absolute bottom-3 right-3 text-white/30 text-xs select-none pointer-events-none">
        drag หมุน · scroll zoom · คลิกดูรายละเอียด
      </div>
    </div>
  )
}
