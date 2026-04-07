import { useEffect, useState } from 'react'
import { BookOpen, Headphones, Star, CheckCircle } from 'lucide-react'
import { api } from '../api'
import type { Stats } from '../types'

interface StatCardProps {
  icon: React.ElementType
  label: string
  value: number
  color: string
}

function StatCard({ icon: Icon, label, value, color }: StatCardProps) {
  return (
    <div className="bg-white rounded-xl border border-slate-200 p-4 flex items-center gap-3">
      <div className={`p-2 rounded-lg ${color} shrink-0`}>
        <Icon size={18} className="text-white" />
      </div>
      <div>
        <p className="text-xl font-bold text-slate-900">{value}</p>
        <p className="text-xs text-slate-500">{label}</p>
      </div>
    </div>
  )
}

const SOURCE_COLORS: Record<string, string> = {
  audible: 'bg-orange-500',
  kindle: 'bg-sky-500',
  storytel: 'bg-emerald-500',
}

const SOURCE_LABELS: Record<string, string> = {
  audible: 'Audible',
  kindle: 'Kindle',
  storytel: 'Storytel',
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.stats().then(setStats).catch((e: Error) => setError(e.message))
  }, [])

  if (error) {
    return (
      <div className="p-4 md:p-8">
        <p className="text-red-600 text-sm">{error}</p>
      </div>
    )
  }

  return (
    <div className="p-4 sm:p-6 md:p-8 max-w-4xl mx-auto md:mx-0">
      <div className="mb-6">
        <h2 className="text-xl md:text-2xl font-bold text-slate-900">Dashboard</h2>
        <p className="text-slate-500 text-sm mt-0.5">
          Your consolidated reading library overview.
        </p>
      </div>

      <div className="grid grid-cols-2 gap-3 mb-6">
        <StatCard
          icon={BookOpen}
          label="Total books"
          value={stats?.total ?? 0}
          color="bg-indigo-500"
        />
        <StatCard
          icon={CheckCircle}
          label="Read"
          value={stats?.read ?? 0}
          color="bg-green-500"
        />
        <StatCard
          icon={Headphones}
          label="Audiobooks"
          value={stats?.byFormat?.audiobook ?? 0}
          color="bg-amber-500"
        />
        <StatCard
          icon={Star}
          label="Rated"
          value={stats?.withRating ?? 0}
          color="bg-purple-500"
        />
      </div>

      {stats && Object.keys(stats.bySource).length > 0 && (
        <div className="bg-white rounded-xl border border-slate-200 p-4 md:p-5">
          <h3 className="text-sm font-semibold text-slate-700 mb-4">By source</h3>
          <div className="space-y-3">
            {Object.entries(stats.bySource).map(([src, count]) => {
              const pct = stats.total > 0 ? Math.round((count / stats.total) * 100) : 0
              return (
                <div key={src}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="font-medium text-slate-700">
                      {SOURCE_LABELS[src] ?? src}
                    </span>
                    <span className="text-slate-500">
                      {count} ({pct}%)
                    </span>
                  </div>
                  <div className="h-1.5 bg-slate-100 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${SOURCE_COLORS[src] ?? 'bg-slate-400'}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {stats?.total === 0 && (
        <div className="bg-white rounded-xl border border-dashed border-slate-300 p-8 text-center mt-4">
          <BookOpen size={28} className="mx-auto text-slate-300 mb-3" />
          <p className="text-slate-500 text-sm">No books yet.</p>
          <p className="text-slate-400 text-xs mt-1">
            Go to <span className="font-medium">Import</span> to add your reading history.
          </p>
        </div>
      )}
    </div>
  )
}
