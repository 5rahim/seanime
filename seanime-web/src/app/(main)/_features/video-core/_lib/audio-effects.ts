import { AudioEffect } from "../video-core-audio-effects"
// File that will compute some filters and also provide settings for the audio effects
// settings
export const C_MAX_GAIN: number = 12
export const C_MIN_GAIN: number = -12

// Maximum value of the frequency for each windows
const SUB_BASS : number = 60
const BASS     : number = 250
const LOW_MIDS : number = 500
const MIDS     : number = 2000
const HIGH_MIDS: number = 4000
const TREBLE   : number = 6000
const BRILLANCE: number = 20000

// EQ_BAND type
type EQBand = {
  freq_hz: number
  sub: string
}

// Returns an array of gains filled with 0s
export function newBand(N_BANDS: number): Array<number> {
  return new Array(N_BANDS).fill(0)
}

// Calculate manually the gains of a bassboost
export function getBassBoostGains(EQ_BANDS: EQBand[]): number[] {
  let gains: number[] = []
  EQ_BANDS.forEach((component: EQBand) => {
    let freq: number = component.freq_hz
    let value: number = 0

    if (freq <= SUB_BASS) {
      value = 0.8 * C_MAX_GAIN
    } else if (freq <= BASS) {
      value = 0.6 * C_MAX_GAIN
    } else if (freq <= LOW_MIDS) {
      value = 0.2 * C_MAX_GAIN
    } else {
      value = 0
    }

    value = clamp(Math.round(10 * value) / 10)
    gains.push(value)
  })

  return gains
}

export function getNightModeGains(EQ_BANDS: EQBand[]): number[] {
  let gains: number[] = []
  EQ_BANDS.forEach((component: EQBand) => {
    let freq = component.freq_hz
    let value = 0

    if (freq <= SUB_BASS) {
      value = 0.3  * C_MIN_GAIN
    } else if (freq <= BASS) {
      value = 0.15 * C_MIN_GAIN
    } else if (freq <= LOW_MIDS) {
      value = 0.3  * C_MAX_GAIN
    } else if (freq <= MIDS) {
      value = 0.15 * C_MAX_GAIN
    } else if (freq <= HIGH_MIDS) {
      value = 0.1  * C_MAX_GAIN
    } else {
      value = 0.2 * C_MIN_GAIN
    }

    value = clamp(Math.round(10 * value) / 10)
    gains.push(value)
  })

  return gains
}

export function labelizeFrequency(band: number): string {
  const label = band < 1000 ? `${band}Hz` : `${Math.round(band/100)/10}kHz`
  return label
}

// Check if two configurations are the same
export function areGainsEqual(a: number[], b: number[]) {
  if (!a || !b || a.length !== b.length) return false
  return a.every((val, index) => val === b[index])
}

// Utils function for gain calculation
export function clamp(value: number): number {
  return Math.min(Math.max(value, C_MIN_GAIN), C_MAX_GAIN)
}

// Variables export
export const EQ_BANDS: EQBand[] = [
  { freq_hz: 60, sub: "Deep bass" },
  { freq_hz: 230, sub: "Bass / Medium" },
  { freq_hz: 910, sub: "Voices / Medium" },
  { freq_hz: 3600, sub: "High pitch" },
  { freq_hz: 5000, sub: "Treble"},
  { freq_hz: 7000, sub: "dummy"},
  { freq_hz: 14000, sub: "Brilliance" },
]

export const defaultGains = newBand(EQ_BANDS.length)
const bassBoostGains = getBassBoostGains(EQ_BANDS)
const nightModeGains = getNightModeGains(EQ_BANDS)

export const AUDIO_EFFECTS_REGISTRY: Record<string, AudioEffect> = {
  bassBoost: {
    id: "bassBoost",
    name: "Bass Boost",
    description: "Bass booster",
    gains: bassBoostGains,
    apply: (context, source) => {
      const filter = context.createBiquadFilter()
      filter.type = "lowshelf"
      filter.frequency.value = 200
      filter.gain.value = C_MAX_GAIN
      source.connect(filter)
      return filter
    },
  },
  nightMode: {
    id: "nightMode",
    name: "Night Mode",
    description: "Compress the dynamics (clear voices, reduced explosions)",
    gains: nightModeGains,
    apply: (context, source) => {
      const compressor = context.createDynamicsCompressor()
      compressor.threshold.value = -30
      compressor.knee.value = 12
      compressor.ratio.value = 8
      compressor.attack.value = 0.003
      compressor.release.value = 0.25
      source.connect(compressor)
      return compressor
    },
  },
  custom: {
    id: "custom",
    name: "Custom",
    description: "Personnalized",
    gains: defaultGains,
    apply: (context, source) => source,
  },
}