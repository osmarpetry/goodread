import { useEffect, useState } from 'react'
import { Download, BookOpen } from 'lucide-react'
import { api } from '../api'
import type { Stats } from '../types'

const SHELF_OPTIONS = [
  { value: 'read', label: 'Read' },
  { value: 'to-read', label: 'To Read' },
  { value: 'currently-reading', label: 'Currently Reading' },
]

export default function Export() {
  const [shelf, setShelf] = useState('read')
  const [dateAdded, setDateAdded] = useState(
    new Date().toISOString().split('T')[0],
  )
  const [stats, setStats] = useState<Stats | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)

  useEffect(() => {
    api.stats().then(setStats).catch(() => null)
  }, [])

  async function handleExport() {
    setLoading(true)
    setError('')
    setSuccess(false)
    try {
      await api.exportGoodreads({ shelf, dateAdded })
      setSuccess(true)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="p-4 sm:p-6 md:p-8 max-w-2xl mx-auto md:mx-0">
      <div className="mb-6">
        <h2 className="text-xl md:text-2xl font-bold text-slate-900">Export to Goodreads</h2>
        <p className="text-slate-500 text-sm mt-0.5">
          Generate a CSV for{' '}
          <span className="font-medium text-slate-600">goodreads.com/review/import</span>.
        </p>
      </div>

      {stats !== null && stats.total > 0 && (
        <div className="bg-indigo-50 border border-indigo-100 rounded-lg p-4 mb-5 flex items-center gap-3">
          <BookOpen size={18} className="text-indigo-500 shrink-0" />
          <p className="text-sm text-indigo-700">
            <span className="font-semibold">{stats.total}</span> books will be exported.
          </p>
        </div>
      )}

      {stats?.total === 0 && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-4 mb-5">
          <p className="text-sm text-amber-700">
            Your library is empty. Import books first.
          </p>
        </div>
      )}

      <div className="bg-white rounded-xl border border-slate-200 p-4 md:p-6 space-y-5">
        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1.5">
            Default shelf
          </label>
          <select
            value={shelf}
            onChange={(e) => setShelf(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
          >
            {SHELF_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
          <p className="text-xs text-slate-400 mt-1.5">
            Used for books without a finish date.
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-slate-700 mb-1.5">
            Date added
          </label>
          <input
            type="date"
            value={dateAdded}
            onChange={(e) => setDateAdded(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
          />
          <p className="text-xs text-slate-400 mt-1.5">
            The "Date Added" column in the Goodreads CSV.
          </p>
        </div>

        {error && (
          <div className="rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-700">
            {error}
          </div>
        )}

        {success && (
          <div className="rounded-lg bg-green-50 border border-green-200 p-3 text-sm text-green-700">
            CSV downloaded. Import it at goodreads.com/review/import.
          </div>
        )}

        <button
          onClick={handleExport}
          disabled={loading || stats?.total === 0}
          className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-700 active:bg-indigo-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <Download size={16} />
          {loading ? 'Generating...' : 'Download CSV'}
        </button>
      </div>
    </div>
  )
}
