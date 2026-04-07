import { useEffect, useMemo, useState } from 'react'
import { Search, BookOpen, Headphones, Star } from 'lucide-react'
import { api } from '../api'
import type { BookItem } from '../types'

const SOURCE_BADGE: Record<string, string> = {
  audible: 'bg-orange-100 text-orange-700',
  kindle: 'bg-sky-100 text-sky-700',
  storytel: 'bg-emerald-100 text-emerald-700',
}

const FORMAT_ICONS: Record<string, React.ElementType> = {
  audiobook: Headphones,
  ebook: BookOpen,
}

function Stars({ rating }: { rating?: number }) {
  if (!rating) return null
  return (
    <div className="flex items-center gap-0.5 mt-1">
      {Array.from({ length: 5 }).map((_, i) => (
        <Star
          key={i}
          size={11}
          className={i < rating ? 'text-amber-400 fill-amber-400' : 'text-slate-200 fill-slate-200'}
        />
      ))}
    </div>
  )
}

function BookCard({ book }: { book: BookItem }) {
  const FormatIcon = book.format ? (FORMAT_ICONS[book.format] ?? BookOpen) : BookOpen
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4 hover:border-slate-300 transition-colors">
      <div className="flex items-start gap-3">
        <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center shrink-0">
          <FormatIcon size={15} className="text-slate-500" />
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-semibold text-slate-900 leading-snug line-clamp-2">
            {book.title}
          </p>
          {book.authors && book.authors.length > 0 && (
            <p className="text-xs text-slate-500 mt-0.5 truncate">
              {book.authors.join(', ')}
            </p>
          )}
          <Stars rating={book.rating} />
        </div>
      </div>
      <div className="flex flex-wrap gap-1.5 mt-3">
        {book.sources?.map((src) => (
          <span
            key={src}
            className={`text-xs px-2 py-0.5 rounded-full font-medium ${SOURCE_BADGE[src] ?? 'bg-slate-100 text-slate-600'}`}
          >
            {src}
          </span>
        ))}
        {book.finishedAt && (
          <span className="text-xs px-2 py-0.5 rounded-full font-medium bg-green-100 text-green-700">
            read
          </span>
        )}
      </div>
    </div>
  )
}

const ALL = 'all'

export default function Library() {
  const [books, setBooks] = useState<BookItem[]>([])
  const [query, setQuery] = useState('')
  const [sourceFilter, setSourceFilter] = useState(ALL)
  const [formatFilter, setFormatFilter] = useState(ALL)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api
      .library()
      .then((lib) => setBooks(lib.items ?? []))
      .catch((e: Error) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  const sources = useMemo(
    () => [...new Set(books.flatMap((b) => b.sources ?? []))],
    [books],
  )
  const formats = useMemo(
    () => [...new Set(books.map((b) => b.format).filter(Boolean))] as string[],
    [books],
  )

  const filtered = useMemo(() => {
    const q = query.toLowerCase()
    return books.filter((b) => {
      if (q && !b.title.toLowerCase().includes(q) && !b.authors?.join(' ').toLowerCase().includes(q))
        return false
      if (sourceFilter !== ALL && !b.sources?.includes(sourceFilter)) return false
      if (formatFilter !== ALL && b.format !== formatFilter) return false
      return true
    })
  }, [books, query, sourceFilter, formatFilter])

  if (loading) {
    return <div className="p-4 md:p-8 text-sm text-slate-400">Loading library...</div>
  }

  return (
    <div className="p-4 sm:p-6 md:p-8">
      <div className="mb-5">
        <h2 className="text-xl md:text-2xl font-bold text-slate-900">Library</h2>
        <p className="text-slate-500 text-sm mt-0.5">
          {books.length} {books.length === 1 ? 'book' : 'books'} in your collection.
        </p>
      </div>

      <div className="flex flex-col gap-2 mb-5">
        <div className="relative">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
          <input
            type="text"
            placeholder="Search by title or author..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-2.5 rounded-lg border border-slate-300 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400"
          />
        </div>

        {(sources.length > 0 || formats.length > 0) && (
          <div className="flex gap-2 overflow-x-auto pb-1">
            {sources.length > 0 && (
              <select
                value={sourceFilter}
                onChange={(e) => setSourceFilter(e.target.value)}
                className="shrink-0 rounded-lg border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 bg-white"
              >
                <option value={ALL}>All sources</option>
                {sources.map((s) => (
                  <option key={s} value={s}>{s}</option>
                ))}
              </select>
            )}

            {formats.length > 0 && (
              <select
                value={formatFilter}
                onChange={(e) => setFormatFilter(e.target.value)}
                className="shrink-0 rounded-lg border border-slate-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 bg-white"
              >
                <option value={ALL}>All formats</option>
                {formats.map((f) => (
                  <option key={f} value={f}>{f}</option>
                ))}
              </select>
            )}
          </div>
        )}
      </div>

      {error && <p className="text-sm text-red-600 mb-4">{error}</p>}

      {filtered.length === 0 ? (
        <div className="bg-white rounded-xl border border-dashed border-slate-300 p-10 text-center">
          <BookOpen size={28} className="mx-auto text-slate-300 mb-3" />
          <p className="text-slate-500 text-sm">
            {books.length === 0 ? 'No books imported yet.' : 'No books match your filters.'}
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {filtered.map((book) => (
            <BookCard key={book.id} book={book} />
          ))}
        </div>
      )}
    </div>
  )
}
