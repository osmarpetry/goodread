import { BrowserRouter, NavLink, Route, Routes } from 'react-router-dom'
import { BookOpen, Download, Import, LayoutDashboard } from 'lucide-react'
import Dashboard from './pages/Dashboard'
import ImportPage from './pages/Import'
import Library from './pages/Library'
import Export from './pages/Export'

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard', end: true },
  { to: '/import', icon: Import, label: 'Import' },
  { to: '/library', icon: BookOpen, label: 'Library' },
  { to: '/export', icon: Download, label: 'Export' },
]

function Sidebar() {
  return (
    <aside className="hidden md:flex w-56 shrink-0 bg-white border-r border-slate-200 flex-col">
      <div className="px-5 py-6 border-b border-slate-100">
        <h1 className="text-base font-semibold text-slate-900 leading-tight">
          media<span className="text-indigo-600">2</span>goodreads
        </h1>
        <p className="text-xs text-slate-400 mt-0.5">Reading history importer</p>
      </div>
      <nav className="flex-1 px-3 py-4 space-y-0.5">
        {navItems.map(({ to, icon: Icon, label, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                isActive
                  ? 'bg-indigo-50 text-indigo-700'
                  : 'text-slate-600 hover:bg-slate-50 hover:text-slate-900'
              }`
            }
          >
            <Icon size={16} />
            {label}
          </NavLink>
        ))}
      </nav>
      <div className="px-5 py-4 border-t border-slate-100">
        <p className="text-xs text-slate-400">All processing is local.</p>
      </div>
    </aside>
  )
}

function BottomNav() {
  return (
    <nav className="md:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-slate-200 flex z-50">
      {navItems.map(({ to, icon: Icon, label, end }) => (
        <NavLink
          key={to}
          to={to}
          end={end}
          className={({ isActive }) =>
            `flex-1 flex flex-col items-center justify-center gap-1 py-2 text-xs font-medium transition-colors ${
              isActive ? 'text-indigo-600' : 'text-slate-500'
            }`
          }
        >
          <Icon size={20} />
          {label}
        </NavLink>
      ))}
    </nav>
  )
}

function MobileHeader() {
  return (
    <header className="md:hidden sticky top-0 z-40 bg-white border-b border-slate-200 px-4 py-3">
      <h1 className="text-sm font-semibold text-slate-900">
        media<span className="text-indigo-600">2</span>goodreads
      </h1>
    </header>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <div className="flex h-screen overflow-hidden">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <MobileHeader />
          <main className="flex-1 overflow-y-auto bg-slate-50 pb-16 md:pb-0">
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/import" element={<ImportPage />} />
              <Route path="/library" element={<Library />} />
              <Route path="/export" element={<Export />} />
            </Routes>
          </main>
        </div>
      </div>
      <BottomNav />
    </BrowserRouter>
  )
}
