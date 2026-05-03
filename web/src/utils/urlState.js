import { normalizeIdolSort, normalizeJavSort } from '@/constants/jav'
import { normalizeVideoSort } from '@/constants/video'

const RANDOM_SEED_MAX = 2147483646

const clampSeed = (seed) => {
  const n = Number(seed)
  if (!Number.isFinite(n) || n <= 0) return null
  return Math.min(Math.floor(n), RANDOM_SEED_MAX)
}

const parseIds = (raw) =>
  (raw || '')
    .split(',')
    .map((s) => Number.parseInt(s.trim(), 10))
    .filter((id) => Number.isFinite(id) && id > 0)

const parseIntSafe = (val, def = 1) => {
  const n = Number.parseInt(val || '', 10)
  return Number.isFinite(n) && n > 0 ? n : def
}

export const parseUrlState = (searchString = window.location.search) => {
  const sp = new URLSearchParams(searchString)
  const view = sp.get('view') === 'jav' ? 'jav' : 'video'
  const directoryId = parseIntSafe(sp.get('directory_id'), 0)

  const videoSortRaw = (sp.get('sort') || '').trim()
  const videoSort = normalizeVideoSort(videoSortRaw)
  const videoTempSort = normalizeVideoSort((sp.get('temp_sort') || '').trim(), '')

  const video = {
    page: parseIntSafe(sp.get('page'), 1),
    search: (sp.get('search') || '').trim(),
    sort: videoSort,
    tempSort: videoTempSort,
    tagIds: parseIds(sp.get('tag_ids')),
    directoryId,
    random: sp.get('random') === '1',
    seed: clampSeed(sp.get('seed')),
  }

  const sortParam = (sp.get('sort') || '').trim().toLowerCase()
  const javSort = normalizeJavSort(sortParam)
  const idolSort = normalizeIdolSort(sortParam)
  const javTempSort = normalizeJavSort((sp.get('temp_sort') || '').trim(), '')

  const jav = {
    tab: sp.get('tab') === 'idol' ? 'idol' : 'list',
    page: parseIntSafe(sp.get('page'), 1),
    search: (sp.get('search') || '').trim(),
    actors: (sp.get('actors') || '')
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean),
    tagIds: parseIds(sp.get('tag_ids')),
    directoryId,
    sort: javSort,
    tempSort: javTempSort,
    idolSort,
    random: sp.get('random') === '1',
    seed: clampSeed(sp.get('seed')),
  }

  return { view, video, jav }
}

export const buildUrlFromState = (state, basePath = window.location.pathname) => {
  const sp = new URLSearchParams()
  if (state.view === 'jav') {
    sp.set('view', 'jav')
    if (state.directoryId > 0) sp.set('directory_id', String(state.directoryId))
    if (state.jav.tab === 'idol') {
      sp.set('tab', 'idol')
    }
    if (state.jav.search) sp.set('search', state.jav.search)
    if (state.jav.actors?.length) sp.set('actors', state.jav.actors.join(','))
    if (state.jav.tab !== 'idol' && state.jav.tagIds?.length) {
      sp.set('tag_ids', state.jav.tagIds.join(','))
    }
    const sortVal = state.jav.tab === 'idol' ? state.jav.idolSort : state.jav.sort
    if (state.jav.tab === 'idol') {
      if (sortVal && sortVal !== 'work') sp.set('sort', sortVal)
    } else if (sortVal && sortVal !== 'recent') {
      sp.set('sort', sortVal)
    }
    if (state.jav.tab !== 'idol' && !state.jav.random && state.jav.tempSort) {
      sp.set('temp_sort', state.jav.tempSort)
    }
    if (state.jav.tab !== 'idol' && state.jav.random) {
      sp.set('random', '1')
      if (state.jav.seed) sp.set('seed', String(state.jav.seed))
    } else {
      sp.set('page', String(state.jav.page || 1))
    }
    const query = sp.toString()
    return `${basePath}${query ? `?${query}` : ''}`
  }

  sp.set('view', 'video')
  if (state.directoryId > 0) sp.set('directory_id', String(state.directoryId))
  if (state.video.search) sp.set('search', state.video.search)
  if (state.video.sort && state.video.sort !== 'recent') sp.set('sort', state.video.sort)
  if (!state.video.random && state.video.tempSort) sp.set('temp_sort', state.video.tempSort)
  if (state.video.tagIds?.length) {
    sp.set('tag_ids', [...state.video.tagIds].sort((a, b) => a - b).join(','))
  }
  if (state.video.random) {
    sp.set('random', '1')
    if (state.video.seed) sp.set('seed', String(state.video.seed))
  } else {
    sp.set('page', String(state.video.page || 1))
  }
  const query = sp.toString()
  return `${basePath}${query ? `?${query}` : ''}`
}

export const normalizeUrlStateFromStore = (store, tagsByName) => {
  const selectedIds = Array.isArray(tagsByName)
    ? []
    : store.selectedTags
        .map((name) => tagsByName.get(name))
        .filter((id) => Number.isFinite(id) && id > 0)

  return {
    view: store.viewMode === 'jav' ? 'jav' : 'video',
    directoryId: store.selectedDirectoryId || 0,
    video: {
      page: store.randomMode ? 1 : store.page,
      search: store.randomMode ? '' : (store.searchTerm || '').trim(),
      sort: store.sortOrder || 'recent',
      tempSort: store.randomMode ? '' : store.videoTempSort || '',
      tagIds: selectedIds,
      directoryId: store.selectedDirectoryId || 0,
      random: store.randomMode,
      seed: store.randomMode ? store.randomSeed : null,
    },
    jav: {
      tab: store.javTab === 'idol' ? 'idol' : 'list',
      page: store.javTab === 'idol' ? store.idolPage : store.javRandomMode ? 1 : store.javPage,
      search: (store.javSearchTerm || '').trim(),
      actors: store.javActors || [],
      tagIds: store.javTags || [],
      directoryId: store.selectedDirectoryId || 0,
      sort: store.javSort || 'recent',
      tempSort: store.javTab !== 'idol' && !store.javRandomMode ? store.javTempSort || '' : '',
      idolSort: store.idolSort || 'work',
      random: store.javTab !== 'idol' && store.javRandomMode,
      seed: store.javTab !== 'idol' && store.javRandomMode ? store.javRandomSeed : null,
    },
  }
}

export const generateRandomSeed = () => Math.floor(Math.random() * RANDOM_SEED_MAX) + 1
