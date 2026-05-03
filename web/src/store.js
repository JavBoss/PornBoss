import { create } from 'zustand'
import {
  fetchTags,
  fetchVideos,
  createTag,
  deleteTag,
  renameTag,
  addTagToVideos,
  removeTagFromVideos,
  fetchDirectories,
  createDirectory,
  updateDirectory,
  deleteDirectory as deleteDirectoryApi,
  fetchJavs,
  fetchJavIdols,
  fetchJavTags,
  fetchConfig,
} from '@/api'
import { normalizeIdolSort, normalizeJavSort } from '@/constants/jav'
import { normalizeVideoSort } from '@/constants/video'
import { zh } from '@/utils/i18n'

const VIDEO_PAGE_SIZE = 25
const JAV_PAGE_SIZE = 24
const JAV_GRID_COLUMNS_AUTO = 0
let videoLoadSeq = 0
let lastVideoFetchKey = null
let lastJavFetchKey = null
let lastIdolFetchKey = null
const RANDOM_SEED_MAX = 2147483646

const normalizeSeed = (seed) => {
  const num = Math.floor(Number(seed))
  if (!Number.isFinite(num) || num <= 0) return null
  return Math.min(num, RANDOM_SEED_MAX)
}

const generateSeed = () => Math.floor(Math.random() * RANDOM_SEED_MAX) + 1

export const useStore = create((set, get) => ({
  // UI state
  page: 1,
  pageSize: VIDEO_PAGE_SIZE,
  setPageSize: (size) => {
    const next = Math.max(1, Math.floor(Number(size) || VIDEO_PAGE_SIZE))
    set({ pageSize: next, videoTempSort: '', page: 1, randomMode: false, randomSeed: null })
  },
  selectedTags: [],
  selectedDirectoryId: 0,
  selectedVideoIds: new Set(),
  selectedVideoMeta: {},
  searchTerm: '',
  sortOrder: 'recent',
  videoTempSort: '',
  javSort: 'recent',
  javTempSort: '',
  randomMode: false,
  randomSeed: null,
  javRandomMode: false,
  javRandomSeed: null,
  viewMode: 'video', // video | jav
  javTab: 'list', // list | idol
  javPage: 1,
  javPageSize: JAV_PAGE_SIZE,
  javGridColumns: JAV_GRID_COLUMNS_AUTO,
  setJavGridColumns: (columns) => {
    const n = Math.floor(Number(columns))
    const next = Number.isFinite(n) && n > 0 ? Math.min(n, 12) : JAV_GRID_COLUMNS_AUTO
    set({ javGridColumns: next })
  },
  setJavPageSize: (size) => {
    const next = Math.max(1, Math.floor(Number(size) || JAV_PAGE_SIZE))
    set({
      javPageSize: next,
      javTempSort: '',
      javRandomMode: false,
      javRandomSeed: null,
      javPage: 1,
    })
  },
  javSearchTerm: '',
  javActors: [],
  javTags: [],
  javItems: [],
  javTotal: 0,
  javLoading: false,
  javError: null,
  idolPage: 1,
  idolPageSize: JAV_PAGE_SIZE,
  idolSort: 'work',
  idolItems: [],
  idolTotal: 0,
  idolLoading: false,
  idolError: null,
  setIdolPageSize: (size) => {
    const next = Math.max(1, Math.floor(Number(size) || JAV_PAGE_SIZE))
    set({ idolPageSize: next, idolPage: 1 })
  },
  setIdolSort: (sort) => {
    const normalized = normalizeIdolSort(sort)
    set({ idolSort: normalized, idolPage: 1 })
  },

  // data
  config: {},
  videos: [],
  tags: [],
  javTagOptions: [],
  directories: [],
  loading: false,
  error: null,
  total: 0,
  hasNext: false,

  // actions
  setPage: (p) => set({ page: p }),
  setSelectedTags: (names, options = {}) => {
    const { resetPage = true, preserveTempSort = false } = options
    const clean = Array.from(new Set((names || []).map((n) => (n || '').trim()).filter(Boolean)))
    const updates = { selectedTags: clean }
    if (!preserveTempSort) {
      updates.videoTempSort = ''
    }
    if (resetPage) {
      updates.page = 1
    }
    set(updates)
  },
  setSelectedDirectoryId: (id) => {
    const parsed = Number.parseInt(String(id || 0), 10)
    const next = Number.isFinite(parsed) && parsed > 0 ? parsed : 0
    set({
      selectedDirectoryId: next,
      page: 1,
      javPage: 1,
      idolPage: 1,
      randomMode: false,
      randomSeed: null,
      javRandomMode: false,
      javRandomSeed: null,
      selectedVideoIds: new Set(),
      selectedVideoMeta: {},
    })
  },
  setSearchTerm: (value, options = {}) => {
    const { resetPage = true } = options
    const trimmed = (value || '').trim()
    const state = get()
    const baseUpdate = { videoTempSort: '', randomMode: false, randomSeed: null }
    if (trimmed === state.searchTerm) {
      // 仅重置分页/随机模式
      const updates = { ...baseUpdate }
      if (resetPage && state.page !== 1) {
        updates.page = 1
      }
      set(updates)
      return
    }
    const next = { searchTerm: trimmed, ...baseUpdate }
    if (resetPage) {
      next.page = 1
    }
    set(next)
  },
  toggleTagFilter: (tagName) => {
    const { selectedTags } = get()
    const exists = selectedTags.includes(tagName)
    const next = exists ? selectedTags.filter((t) => t !== tagName) : [...selectedTags, tagName]
    set({ selectedTags: next, videoTempSort: '', page: 1 })
  },
  clearFilters: () => set({ selectedTags: [], videoTempSort: '', page: 1 }),
  toggleSelectVideo: (video) => {
    if (!video || !video.id) return
    const id = video.id
    const label = video.filename || video.path || `#${id}`
    const setIds = new Set(get().selectedVideoIds)
    const meta = { ...get().selectedVideoMeta }
    if (setIds.has(id)) {
      setIds.delete(id)
      delete meta[id]
    } else {
      setIds.add(id)
      meta[id] = label
    }
    set({ selectedVideoIds: setIds, selectedVideoMeta: meta })
  },
  clearSelection: () => set({ selectedVideoIds: new Set(), selectedVideoMeta: {} }),
  setSortOrder: (order) => {
    const normalized = normalizeVideoSort(order)
    set({ sortOrder: normalized, videoTempSort: '', randomMode: false, randomSeed: null, page: 1 })
  },
  setVideoTempSort: (order) => {
    const normalized = normalizeVideoSort(order, '')
    set({ videoTempSort: normalized, randomMode: false, randomSeed: null })
  },
  setJavSort: (order) => {
    const normalized = normalizeJavSort(order)
    set({
      javSort: normalized,
      javTempSort: '',
      javRandomMode: false,
      javRandomSeed: null,
      javPage: 1,
    })
  },
  setJavTempSort: (order) => {
    const normalized = normalizeJavSort(order, '')
    set({ javTempSort: normalized, javRandomMode: false, javRandomSeed: null })
  },
  clearRandomMode: () => set({ randomMode: false, randomSeed: null }),
  clearJavRandom: () => set({ javTempSort: '', javRandomMode: false, javRandomSeed: null }),
  setViewMode: (mode) => {
    if (mode !== 'video' && mode !== 'jav') return
    set({ viewMode: mode, ...(mode === 'jav' ? { videoTempSort: '' } : { javTempSort: '' }) })
  },
  setJavTab: (tab) => {
    if (tab !== 'list' && tab !== 'idol') return
    set({ javTab: tab, javTempSort: '' })
  },
  setJavActors: (actors) => {
    const clean = Array.from(new Set((actors || []).map((n) => (n || '').trim()).filter(Boolean)))
    set({ javActors: clean, javTempSort: '', javPage: 1 })
  },
  setJavTags: (tags) => {
    const clean = Array.from(
      new Set(
        (tags || [])
          .map((t) => Number.parseInt(String(t), 10))
          .filter((id) => Number.isFinite(id) && id > 0)
      )
    )
    set({ javTags: clean, javTempSort: '', javPage: 1 })
  },
  setJavPage: (p) => {
    const state = get()
    set({ javPage: state.javRandomMode ? 1 : p })
  },
  setIdolPage: (p) => set({ idolPage: p }),
  setJavSearchTerm: (value, options = {}) => {
    const { resetPage = true } = options
    const trimmed = (value || '').trim()
    const state = get()
    if (trimmed === state.javSearchTerm) {
      if (resetPage && state.javPage !== 1) {
        set({ javTempSort: '', javPage: 1, idolPage: 1 })
      }
      return
    }
    const next = { javSearchTerm: trimmed, javTempSort: '' }
    if (resetPage) {
      next.javPage = 1
      next.idolPage = 1
    }
    set(next)
  },

  loadTags: async () => {
    try {
      const tags = await fetchTags({ directoryId: get().selectedDirectoryId })
      set({ tags })
    } catch (e) {
      set({ error: e.message })
    }
  },
  loadJavTags: async () => {
    try {
      const tags = await fetchJavTags({ directoryId: get().selectedDirectoryId })
      set({ javTagOptions: tags })
    } catch (e) {
      set({ javError: e.message || zh('加载 JAV 标签失败', 'Failed to load JAV tags') })
    }
  },
  loadConfig: async () => {
    try {
      const cfg = await fetchConfig()
      const state = get()
      const clamp = (raw) => {
        const n = parseInt(raw, 10)
        if (!Number.isFinite(n) || n <= 0) return null
        return Math.min(n, 500)
      }
      const updates = { config: cfg }
      const videoSize = clamp(cfg?.video_page_size)
      const videoSort = normalizeVideoSort((cfg?.video_sort || '').toLowerCase(), '')
      const javSize = clamp(cfg?.jav_page_size)
      const javGridColumnsRaw = parseInt(cfg?.jav_grid_columns, 10)
      const javGridColumns =
        Number.isFinite(javGridColumnsRaw) && javGridColumnsRaw > 0
          ? Math.min(javGridColumnsRaw, 12)
          : JAV_GRID_COLUMNS_AUTO
      const idolSize = clamp(cfg?.idol_page_size)
      const javSort = normalizeJavSort((cfg?.jav_sort || '').toLowerCase(), '')
      const idolSort = normalizeIdolSort((cfg?.idol_sort || '').toLowerCase(), '')
      if (videoSize && videoSize !== state.pageSize) {
        updates.pageSize = videoSize
      }
      if (videoSort) {
        updates.sortOrder = videoSort
      }
      if (javSort) {
        updates.javSort = javSort
      }
      if (idolSort) {
        updates.idolSort = idolSort
      }
      if (javSize && javSize !== state.javPageSize) {
        updates.javPageSize = javSize
      }
      if (javGridColumns !== state.javGridColumns) {
        updates.javGridColumns = javGridColumns
      }
      if (idolSize && idolSize !== state.idolPageSize) {
        updates.idolPageSize = idolSize
      }
      set(updates)
      return cfg
    } catch (e) {
      console.error('load config failed', e)
      return null
    }
  },
  loadDirectories: async () => {
    try {
      const directories = await fetchDirectories()
      const activeDirectories = directories.filter((d) => !d.is_delete)
      const selectedDirectoryId = get().selectedDirectoryId
      const selectedExists =
        selectedDirectoryId > 0 && activeDirectories.some((d) => d.id === selectedDirectoryId)
      set({
        directories: activeDirectories,
        ...(selectedDirectoryId > 0 && !selectedExists ? { selectedDirectoryId: 0 } : {}),
      })
    } catch (e) {
      console.error(zh('加载目录失败', 'Failed to load directories'), e)
    }
  },
  loadVideos: async (options = {}) => {
    const {
      page: p0,
      pageSize,
      selectedTags,
      searchTerm,
      sortOrder,
      videoTempSort,
      randomMode,
      randomSeed,
      selectedDirectoryId,
    } = get()
    const search = searchTerm ? searchTerm : ''
    const effectiveSort = videoTempSort || sortOrder
    const key = [
      randomMode ? 'r' : 'p',
      randomMode ? 1 : p0,
      pageSize,
      search,
      effectiveSort,
      randomMode ? randomSeed || '' : '',
      (selectedTags || []).join(','),
      selectedDirectoryId || '',
    ].join('|')
    if (!options.force && key === lastVideoFetchKey) {
      return
    }
    lastVideoFetchKey = key
    const reqId = (videoLoadSeq += 1)
    set({ loading: true, error: null })
    try {
      const resp = await fetchVideos({
        limit: pageSize,
        offset: randomMode ? 0 : (p0 - 1) * pageSize,
        tags: selectedTags,
        search,
        sort: randomMode ? 'random' : effectiveSort,
        seed: randomMode ? randomSeed : null,
        directoryId: selectedDirectoryId,
      })
      if (reqId !== videoLoadSeq) return
      const total = resp.total ?? 0
      const items = resp.items ?? []
      const lastPage = Math.max(1, Math.ceil(total / pageSize))
      const hasNext = randomMode ? false : p0 < lastPage
      set({ videos: items, total, hasNext })
    } catch (e) {
      if (reqId !== videoLoadSeq) return
      set({ error: e.message })
    } finally {
      if (reqId === videoLoadSeq) {
        set({ loading: false })
      }
    }
  },
  loadJavs: async (options = {}) => {
    const {
      javPage,
      javPageSize,
      javSearchTerm,
      javActors,
      javTags,
      javSort,
      javTempSort,
      javRandomMode,
      javRandomSeed,
      selectedDirectoryId,
    } = get()
    const search = javSearchTerm || ''
    const effectiveSort = javTempSort || javSort
    const key = [
      javRandomMode ? 'r' : 'p',
      javRandomMode ? 1 : javPage,
      javPageSize,
      search,
      (javActors || []).join(','),
      (javTags || []).join(','),
      effectiveSort,
      javRandomMode ? javRandomSeed || '' : '',
      selectedDirectoryId || '',
    ].join('|')
    if (!options.force && key === lastJavFetchKey) {
      return
    }
    lastJavFetchKey = key
    set({ javLoading: true, javError: null })
    try {
      const resp = await fetchJavs({
        limit: javPageSize,
        offset: javRandomMode ? 0 : (javPage - 1) * javPageSize,
        search,
        actors: javActors,
        tagIds: javTags,
        sort: javRandomMode ? 'random' : effectiveSort,
        seed: javRandomMode ? javRandomSeed : null,
        directoryId: selectedDirectoryId,
      })
      const items = resp.items || []
      set({
        javItems: items,
        javTotal: javRandomMode ? items.length : resp.total || 0,
      })
    } catch (e) {
      set({ javError: e.message || zh('加载 JAV 失败', 'Failed to load JAV') })
    } finally {
      set({ javLoading: false })
    }
  },
  loadJavIdols: async (options = {}) => {
    const { idolPage, idolPageSize, javSearchTerm, idolSort, selectedDirectoryId } = get()
    const search = javSearchTerm || ''
    const key = ['idol', idolPage, idolPageSize, search, idolSort, selectedDirectoryId || ''].join(
      '|'
    )
    if (!options.force && key === lastIdolFetchKey) {
      return
    }
    lastIdolFetchKey = key
    set({ idolLoading: true, idolError: null })
    try {
      const resp = await fetchJavIdols({
        limit: idolPageSize,
        offset: (idolPage - 1) * idolPageSize,
        search,
        sort: idolSort,
        directoryId: selectedDirectoryId,
      })
      set({
        idolItems: resp.items || [],
        idolTotal: resp.total || 0,
      })
    } catch (e) {
      set({ idolError: e.message || zh('加载女优失败', 'Failed to load idols') })
    } finally {
      set({ idolLoading: false })
    }
  },

  createTag: async (name) => {
    const tag = await createTag(name)
    set({ tags: [...get().tags, tag] })
  },
  deleteTag: async (id) => {
    await deleteTag(id)
    set({ tags: get().tags.filter((t) => t.id !== id) })
  },
  renameTag: async (id, name) => {
    await renameTag(id, name)
    set({ tags: get().tags.map((t) => (t.id === id ? { ...t, name } : t)) })
  },
  addTagToSelection: async (tagId) => {
    const ids = Array.from(get().selectedVideoIds)
    if (ids.length === 0) return
    await addTagToVideos(tagId, ids)
    await get().loadVideos()
  },
  removeTagFromSelection: async (tagId) => {
    const ids = Array.from(get().selectedVideoIds)
    if (ids.length === 0) return
    await removeTagFromVideos(tagId, ids)
    await get().loadVideos()
  },
  goToLastPage: async () => {
    set({ loading: true, error: null })
    try {
      const {
        pageSize,
        selectedTags,
        searchTerm,
        sortOrder,
        videoTempSort,
        randomMode,
        randomSeed,
        selectedDirectoryId,
      } = get()
      const effectiveSort = videoTempSort || sortOrder
      // Get total via a cheap fetch (limit=1) or use existing total
      let { total } = get()
      const search = searchTerm ? searchTerm : ''
      if (!total) {
        const res = await fetchVideos({
          limit: 1,
          offset: 0,
          tags: selectedTags,
          search,
          sort: randomMode ? 'random' : effectiveSort,
          seed: randomMode ? randomSeed : null,
          directoryId: selectedDirectoryId,
        })
        total = res.total ?? 0
        set({ total })
      }
      const lastPage = Math.max(1, Math.ceil(total / pageSize))
      const res2 = await fetchVideos({
        limit: pageSize,
        offset: (lastPage - 1) * pageSize,
        tags: selectedTags,
        search,
        sort: randomMode ? 'random' : effectiveSort,
        seed: randomMode ? randomSeed : null,
        directoryId: selectedDirectoryId,
      })
      const items = res2.items ?? []
      set({ page: lastPage, videos: items, hasNext: false })
    } catch (e) {
      set({ error: e.message })
    } finally {
      set({ loading: false })
    }
  },
  loadRandom: async (seed) => {
    const nextSeed = normalizeSeed(seed) ?? generateSeed()
    const nextPage = 1
    set({ videoTempSort: '', randomMode: true, randomSeed: nextSeed, page: nextPage })
  },
  loadJavRandom: async (seed) => {
    const nextSeed = normalizeSeed(seed) ?? generateSeed()
    set({ javTempSort: '', javRandomMode: true, javRandomSeed: nextSeed, javPage: 1 })
  },

  createDirectory: async ({ path }) => {
    const dir = await createDirectory({ path })
    const next = dir && !dir.is_delete ? [...get().directories, dir] : get().directories
    set({ directories: next })
    return dir
  },
  updateDirectory: async (id, payload) => {
    const dir = await updateDirectory(id, payload)
    set({
      directories: get()
        .directories.map((d) => (d.id === id ? dir : d))
        .filter((d) => d && !d.is_delete),
    })
    return dir
  },
  deleteDirectory: async (id) => {
    const dir = await deleteDirectoryApi(id)
    set({
      directories: get()
        .directories.map((d) => (d.id === id ? dir : d))
        .filter((d) => d && !d.is_delete),
    })
    return dir
  },
}))
