export interface ProbeItem {
  domain: string
  tld: string
  available: boolean | null
  is_premium: boolean | null
  currency: string | null
  base_register_price: number | null
  premium_register_price: number | null
  icann_fee: number | null
  eap_fee: number | null
  total_price: number | null
  error: string | null
  raw: Record<string, any>
}

export interface ProgressEvent {
  type: 'progress'
  index: number
  total: number
  data: ProbeItem
}

export interface PhaseEvent {
  type: 'phase'
  phase: string
  message: string
}

export interface SummaryEvent {
  type: 'summary'
  word: string
  total: number
  outdir: string
  report_md: string
  results_csv: string
  results_jsonl: string
  finished_at: string
  elapsed_seconds: number
}

export interface ProbeTask {
  id: string
  word: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  tld_mode: TldMode
  rate_per_min: number
  max_batch: number
  cache_ttl_hours: number
  total: number
  completed: number
  error?: string
  created_at: string
  updated_at: string
  finished_at?: string
  outdir?: string
  report_md?: string
  results_csv?: string
  results_jsonl?: string
  elapsed_seconds?: number
}

export type TldMode = 'all' | 'api-registerable-only' | 'mainstream-only'

export interface TLDOption {
  value: string
  label: string
  price?: number
  category: 'popular' | 'cheap' | 'other'
}

export const POPULAR_TLDS: TLDOption[] = [
  { value: 'com', label: '.com', category: 'popular' },
  { value: 'net', label: '.net', category: 'popular' },
  { value: 'org', label: '.org', category: 'popular' },
  { value: 'io', label: '.io', category: 'popular' },
  { value: 'co', label: '.co', category: 'popular' },
  { value: 'app', label: '.app', category: 'popular' },
  { value: 'dev', label: '.dev', category: 'popular' },
  { value: 'ai', label: '.ai', category: 'popular' },
]

export const CHEAP_TLDS: TLDOption[] = [
  { value: 'xyz', label: '.xyz', price: 0.99, category: 'cheap' },
  { value: 'top', label: '.top', price: 1.99, category: 'cheap' },
  { value: 'club', label: '.club', price: 1.99, category: 'cheap' },
  { value: 'site', label: '.site', price: 2.99, category: 'cheap' },
  { value: 'online', label: '.online', price: 3.99, category: 'cheap' },
  { value: 'store', label: '.store', price: 4.99, category: 'cheap' },
  { value: 'space', label: '.space', price: 2.99, category: 'cheap' },
  { value: 'icu', label: '.icu', price: 1.99, category: 'cheap' },
]

export const OTHER_TLDS: TLDOption[] = [
  { value: 'info', label: '.info', category: 'other' },
  { value: 'biz', label: '.biz', category: 'other' },
  { value: 'me', label: '.me', category: 'other' },
  { value: 'cc', label: '.cc', category: 'other' },
  { value: 'tv', label: '.tv', category: 'other' },
  { value: 'ws', label: '.ws', category: 'other' },
  { value: 'us', label: '.us', category: 'other' },
  { value: 'eu', label: '.eu', category: 'other' },
  { value: 'de', label: '.de', category: 'other' },
  { value: 'uk', label: '.uk', category: 'other' },
  { value: 'fr', label: '.fr', category: 'other' },
  { value: 'jp', label: '.jp', category: 'other' },
  { value: 'cn', label: '.cn', category: 'other' },
  { value: 'ru', label: '.ru', category: 'other' },
]

export const ALL_TLDS = [...POPULAR_TLDS, ...CHEAP_TLDS, ...OTHER_TLDS]

export function getTLDsByMode(mode: TldMode): string[] {
  switch (mode) {
    case 'mainstream-only':
      return POPULAR_TLDS.map(t => t.value)
    case 'api-registerable-only':
      return CHEAP_TLDS.map(t => t.value)
    case 'all':
    default:
      return ALL_TLDS.map(t => t.value)
  }
}
