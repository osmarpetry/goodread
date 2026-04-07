import type { Library, Stats, ImportResponse, ExportRequest } from './types'

const BASE = '/api'

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, init)
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(body.error ?? res.statusText)
  }
  return res.json() as Promise<T>
}

export const api = {
  stats(): Promise<Stats> {
    return request('/stats')
  },

  library(): Promise<Library> {
    return request('/library')
  },

  importAudible(file: File, format: string, region: string): Promise<ImportResponse> {
    const form = new FormData()
    form.append('file', file)
    form.append('format', format)
    form.append('region', region)
    return request('/import/audible', { method: 'POST', body: form })
  },

  importKindle(file: File): Promise<ImportResponse> {
    const form = new FormData()
    form.append('file', file)
    return request('/import/kindle', { method: 'POST', body: form })
  },

  importStoritel(file: File, format: string): Promise<ImportResponse> {
    const form = new FormData()
    form.append('file', file)
    form.append('format', format)
    return request('/import/storytel', { method: 'POST', body: form })
  },

  async exportGoodreads(req: ExportRequest): Promise<void> {
    const res = await fetch(`${BASE}/export/goodreads`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    })
    if (!res.ok) {
      const body = await res.json().catch(() => ({ error: res.statusText }))
      throw new Error(body.error ?? res.statusText)
    }
    const blob = await res.blob()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'goodreads_import.csv'
    a.click()
    URL.revokeObjectURL(url)
  },
}
