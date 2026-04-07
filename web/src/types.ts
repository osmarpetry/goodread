export interface BookItem {
  id: string
  title: string
  authors?: string[]
  asin?: string
  isbn10?: string
  isbn13?: string
  format?: string
  sources?: string[]
  startedAt?: string
  finishedAt?: string
  rating?: number
  review?: string
  publisher?: string
  pageCount?: number
  yearPublished?: number
  notes?: string[]
}

export interface Library {
  items: BookItem[]
}

export interface Stats {
  total: number
  byFormat: Record<string, number>
  bySource: Record<string, number>
  withRating: number
  read: number
}

export interface ImportResponse {
  total: number
  added: number
  merged: number
  message: string
}

export interface ExportRequest {
  shelf: string
  dateAdded: string
}
