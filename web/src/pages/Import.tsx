import { useRef, useState } from 'react'
import { Upload, CheckCircle2, AlertCircle } from 'lucide-react'
import { api } from '../api'
import type { ImportResponse } from '../types'

type Source = 'audible' | 'kindle' | 'storytel'

const TABS: { id: Source; label: string }[] = [
  { id: 'audible', label: 'Audible' },
  { id: 'kindle', label: 'Kindle' },
  { id: 'storytel', label: 'Storytel' },
]

const SOURCE_INFO: Record<Source, { description: string; accept: string; hint: string }> = {
  audible: {
    description:
      'Export your library from OpenAudible (File → Export → JSON or CSV), then upload the file here.',
    accept: '.json,.csv',
    hint: 'books.json or books.csv from OpenAudible',
  },
  kindle: {
    description:
      'Connect your Kindle via USB and copy documents/My Clippings.txt, then upload it here.',
    accept: '.txt',
    hint: 'My Clippings.txt from your Kindle device',
  },
  storytel: {
    description:
      'Export your Storytel history as CSV or JSON from your account, then upload it here.',
    accept: '.csv,.json',
    hint: 'storytel_history.csv or .json',
  },
}

interface FileDropzoneProps {
  accept: string
  hint: string
  file: File | null
  onFile: (f: File) => void
}

function FileDropzone({ accept, hint, file, onFile }: FileDropzoneProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [dragging, setDragging] = useState(false)

  function handleDrop(e: React.DragEvent) {
    e.preventDefault()
    setDragging(false)
    const f = e.dataTransfer.files[0]
    if (f) onFile(f)
  }

  return (
    <div
      className={`border-2 border-dashed rounded-xl p-6 text-center cursor-pointer transition-colors ${
        dragging
          ? 'border-indigo-400 bg-indigo-50'
          : file
          ? 'border-green-400 bg-green-50'
          : 'border-slate-300 hover:border-slate-400 bg-white'
      }`}
      onClick={() => inputRef.current?.click()}
      onDragOver={(e) => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
    >
      <input
        ref={inputRef}
        type="file"
        accept={accept}
        className="hidden"
        onChange={(e) => { const f = e.target.files?.[0]; if (f) onFile(f) }}
      />
      <Upload size={22} className={`mx-auto mb-2 ${file ? 'text-green-500' : 'text-slate-400'}`} />
      {file ? (
        <p className="text-sm font-medium text-green-700 break-all">{file.name}</p>
      ) : (
        <>
          <p className="text-sm font-medium text-slate-600">
            Tap to choose a file
          </p>
          <p className="text-xs text-slate-400 mt-1">{hint}</p>
        </>
      )}
    </div>
  )
}

function ImportPanel({ source }: { source: Source }) {
  const info = SOURCE_INFO[source]
  const [file, setFile] = useState<File | null>(null)
  const [format, setFormat] = useState('json')
  const [region, setRegion] = useState('us')
  const [loading, setLoading] = useState(false)
  const [result, setResult] = useState<ImportResponse | null>(null)
  const [error, setError] = useState('')

  async function handleImport() {
    if (!file) return
    setLoading(true)
    setError('')
    setResult(null)
    try {
      let res: ImportResponse
      if (source === 'audible') res = await api.importAudible(file, format, region)
      else if (source === 'kindle') res = await api.importKindle(file)
      else res = await api.importStoritel(file, format)
      setResult(res)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-600">{info.description}</p>

      <FileDropzone accept={info.accept} hint={info.hint} file={file} onFile={setFile} />

      {source === 'audible' && (
        <div className="grid grid-cols-2 gap-3">
          <div>
            <label className="block text-xs font-medium text-slate-600 mb-1">Format</label>
            <select
              value={format}
              onChange={(e) => setFormat(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
            >
              <option value="json">JSON</option>
              <option value="csv">CSV</option>
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-600 mb-1">Region</label>
            <select
              onChange={(e) => setRegion(e.target.value)}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
            >
              {['us', 'uk', 'de', 'br'].map((r) => (
                <option key={r} value={r}>{r.toUpperCase()}</option>
              ))}
            </select>
          </div>
        </div>
      )}

      {source === 'storytel' && (
        <div>
          <label className="block text-xs font-medium text-slate-600 mb-1">Format</label>
          <select
            value={format}
            onChange={(e) => setFormat(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
          >
            <option value="csv">CSV</option>
            <option value="json">JSON</option>
          </select>
        </div>
      )}

      {error && (
        <div className="flex items-start gap-2 rounded-lg bg-red-50 border border-red-200 p-3 text-sm text-red-700">
          <AlertCircle size={16} className="shrink-0 mt-0.5" />
          {error}
        </div>
      )}

      {result && (
        <div className="flex items-start gap-2 rounded-lg bg-green-50 border border-green-200 p-3 text-sm text-green-700">
          <CheckCircle2 size={16} className="shrink-0 mt-0.5" />
          <span>
            {result.message} &mdash;{' '}
            <span className="font-medium">{result.added} new</span>,{' '}
            <span className="font-medium">{result.merged} merged</span>
          </span>
        </div>
      )}

      <button
        onClick={handleImport}
        disabled={!file || loading}
        className="w-full flex items-center justify-center gap-2 px-4 py-3 rounded-lg bg-indigo-600 text-white text-sm font-medium hover:bg-indigo-700 active:bg-indigo-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        <Upload size={16} />
        {loading ? 'Importing...' : 'Import'}
      </button>
    </div>
  )
}

export default function ImportPage() {
  const [active, setActive] = useState<Source>('audible')

  return (
    <div className="p-4 sm:p-6 md:p-8 max-w-2xl mx-auto md:mx-0">
      <div className="mb-6">
        <h2 className="text-xl md:text-2xl font-bold text-slate-900">Import</h2>
        <p className="text-slate-500 text-sm mt-0.5">
          Add books from your reading platforms.
        </p>
      </div>

      <div className="flex gap-1 p-1 bg-slate-100 rounded-lg mb-5">
        {TABS.map((t) => (
          <button
            key={t.id}
            onClick={() => setActive(t.id)}
            className={`flex-1 py-2 rounded-md text-sm font-medium transition-colors ${
              active === t.id
                ? 'bg-white text-slate-900 shadow-sm'
                : 'text-slate-500 hover:text-slate-700'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="bg-white rounded-xl border border-slate-200 p-4 md:p-6">
        <ImportPanel key={active} source={active} />
      </div>
    </div>
  )
}
