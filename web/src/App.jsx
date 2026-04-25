import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import {
  buildUrlFromState,
  generateRandomSeed,
  normalizeUrlStateFromStore,
  parseUrlState,
} from '@/utils/urlState'
import {
  addTagToVideos,
  removeTagFromVideos,
  replaceTagsForVideos,
  updateConfig,
  playVideoFile,
  openVideoFile,
  revealVideoLocation,
  createJavTag,
  renameJavTag,
  deleteJavTag,
  replaceJavTagsForItems,
} from '@/api'
import GlobalSettingsModal from '@/components/GlobalSettingsModal'
import JavIdolView from '@/components/JavIdolView'
import JavSettingsModal from '@/components/JavSettingsModal'
import JavTagModal from '@/components/JavTagModal'
import JavVideoPickerModal from '@/components/JavVideoPickerModal'
import JavView from '@/components/JavView'
import SelectionOpsModal from '@/components/SelectionOpsModal'
import SelectionTagsModal from '@/components/SelectionTagsModal'
import TagPickerModal from '@/components/TagPickerModal'
import Toast from '@/components/Toast'
import TopBar from '@/components/TopBar'
import VideoSettingsModal from '@/components/VideoSettingsModal'
import VideoTagModal from '@/components/VideoTagModal'
import VideoView from '@/components/VideoView'
import { isUserJavTag, normalizeJavSort } from '@/constants/jav'
import { isChineseLocale, zh } from '@/utils/i18n'
import { useStore } from '@/store'

export default function App() {
  const isPoppingRef = useRef(false)
  const lastUrlRef = useRef(window.location.pathname + window.location.search)
  const pendingVideoTagIdsRef = useRef(null)
  const {
    page,
    pageSize,
    setPage,
    videos,
    config,
    tags,
    selectedTags,
    selectedVideoIds,
    selectedVideoMeta,
    loadVideos,
    loadTags,
    toggleTagFilter,
    createTag,
    deleteTag,
    renameTag,
    toggleSelectVideo,
    loading,
    error,
    hasNext,
    total,
    setSelectedTags,
    clearSelection,
    searchTerm,
    setSearchTerm,
    sortOrder,
    loadJavRandom,
    randomMode,
    randomSeed,
    viewMode,
    setViewMode,
    javTab,
    javPage,
    setJavPage,
    javPageSize,
    javSearchTerm,
    javActors,
    javTags,
    javSort,
    javPageSort,
    javRandomMode,
    javRandomSeed,
    idolSort,
    setJavPageSort,
    loadJavs,
    javItems,
    javTotal,
    javLoading,
    javError,
    javTagOptions,
    loadJavTags,
    loadConfig,
    idolPage,
    setIdolPage,
    idolPageSize,
    idolItems,
    idolTotal,
    idolLoading,
    idolError,
    loadJavIdols,
    directories,
    loadDirectories,
    createDirectory,
    updateDirectory,
    deleteDirectory,
  } = useStore()

  const [tagModalOpen, setTagModalOpen] = useState(false)
  const [videoSettingsOpen, setVideoSettingsOpen] = useState(false)
  const [javSettingsOpen, setJavSettingsOpen] = useState(false)
  const [globalSettingsOpen, setGlobalSettingsOpen] = useState(false)
  const [javTagModalOpen, setJavTagModalOpen] = useState(false)
  const [javVideoPickerOpen, setJavVideoPickerOpen] = useState(false)
  const [javVideoPickerItem, setJavVideoPickerItem] = useState(null)
  const [javVideoPickerAction, setJavVideoPickerAction] = useState('play')
  const [searchInput, setSearchInput] = useState('')
  const [javSearchInput, setJavSearchInput] = useState('')
  const [hydrated, setHydrated] = useState(false)
  const [configLoaded, setConfigLoaded] = useState(false)
  const isJavMode = viewMode === 'jav'
  const isModifiedClick = (e) =>
    e && (e.metaKey || e.ctrlKey || e.shiftKey || e.altKey || e.button !== 0)
  const selectedTagIds = useMemo(
    () =>
      tags
        .filter((t) => selectedTags.includes(t.name))
        .map((t) => t.id)
        .filter((id) => id > 0),
    [tags, selectedTags]
  )
  const tagsByName = useMemo(() => new Map(tags.map((t) => [t.name, t.id])), [tags])
  const [tagPickerFor, setTagPickerFor] = useState(null)
  const [tagPickerSelected, setTagPickerSelected] = useState([])
  const [javTagPickerItem, setJavTagPickerItem] = useState(null)
  const [javTagPickerSelected, setJavTagPickerSelected] = useState([])
  const [selectionOpsOpen, setSelectionOpsOpen] = useState(false)
  const [selectionTagsOpen, setSelectionTagsOpen] = useState(false)
  const [selectionTagAction, setSelectionTagAction] = useState('add')
  const [selectionTagChoices, setSelectionTagChoices] = useState([])
  const [videoPageSizeInput, setVideoPageSizeInput] = useState(pageSize)
  const [videoSortInput, setVideoSortInput] = useState(sortOrder)
  const [javPageSizeInput, setJavPageSizeInput] = useState(javPageSize)
  const [idolPageSizeInput, setIdolPageSizeInput] = useState(idolPageSize)
  const [javSortInput, setJavSortInput] = useState(javSort)
  const [idolSortInput, setIdolSortInput] = useState(idolSort)
  const [toastMessage, setToastMessage] = useState('')
  const javVideoChoices = javVideoPickerItem?.videos || []
  const showToast = useCallback((message) => {
    setToastMessage(String(message || '').trim())
  }, [])
  const handleOpenTagModal = useCallback(() => {
    loadTags()
    setTagModalOpen(true)
  }, [loadTags])

  const mapTagIdsToNames = useCallback(
    (ids) => {
      if (!Array.isArray(ids) || ids.length === 0) return []
      const idSet = new Set(ids)
      return tags.filter((t) => idSet.has(t.id)).map((t) => t.name)
    },
    [tags]
  )

  const buildVideoFullPath = useCallback((video) => {
    if (!video) return ''
    const rawPath = String(video.path || '').trim()
    const dirPath = String(video.directory?.path || video.directory_path || '').trim()
    if (!dirPath) return rawPath
    if (!rawPath) return dirPath
    const isAbs = rawPath.startsWith('/') || /^[A-Za-z]:[\\/]/.test(rawPath)
    if (isAbs) return rawPath
    const separator = dirPath.includes('\\') ? '\\' : '/'
    const cleanedDir = dirPath.replace(/[\\/]+$/, '')
    const cleanedRel = rawPath.replace(/^[\\/]+/, '')
    return `${cleanedDir}${separator}${cleanedRel}`
  }, [])

  const getVideoDirPath = useCallback(
    (video) => String(video?.directory?.path || video?.directory_path || '').trim(),
    []
  )

  const getVideoRelPath = useCallback((video) => String(video?.path || '').trim(), [])

  const isVideoOpenable = useCallback(
    (video) => Boolean(getVideoDirPath(video) && getVideoRelPath(video)),
    [getVideoDirPath, getVideoRelPath]
  )

  const handleOpenPlayer = useCallback(
    (video) => {
      if (!video || !isVideoOpenable(video)) return
      playVideoFile({ path: getVideoRelPath(video), dirPath: getVideoDirPath(video) }).catch(
        (err) => console.error(zh('播放文件失败', 'Failed to play file'), err)
      )
    },
    [getVideoDirPath, getVideoRelPath, isVideoOpenable]
  )

  const closeJavVideoPicker = useCallback(() => {
    setJavVideoPickerOpen(false)
    setJavVideoPickerItem(null)
    setJavVideoPickerAction('play')
  }, [])

  const handleVideoTagClick = useCallback(
    (name) => {
      if (!name) return
      setSearchTerm('', { resetPage: false, triggerLoad: false })
      setSelectedTags([name])
    },
    [setSearchTerm, setSelectedTags]
  )

  const handleJavPlay = useCallback(
    (video, item) => {
      const videos = item?.videos || []
      if (videos.length > 1) {
        setJavVideoPickerAction('play')
        setJavVideoPickerItem(item)
        setJavVideoPickerOpen(true)
        return
      }
      const target = video || videos[0]
      if (target) {
        handleOpenPlayer(target)
      }
    },
    [handleOpenPlayer]
  )

  const handleJavOpenFile = useCallback(
    (video, item) => {
      const videos = item?.videos || (video ? [video] : [])
      if (videos.length > 1) {
        setJavVideoPickerAction('open')
        setJavVideoPickerItem(item)
        setJavVideoPickerOpen(true)
        return
      }
      const target = video && isVideoOpenable(video) ? video : videos.find(isVideoOpenable)
      if (!target) return
      openVideoFile({ path: getVideoRelPath(target), dirPath: getVideoDirPath(target) }).catch(
        (err) => console.error(zh('打开文件失败', 'Failed to open file'), err)
      )
    },
    [getVideoDirPath, getVideoRelPath, isVideoOpenable]
  )

  const handleJavRevealFile = useCallback(
    (video, item) => {
      const videos = item?.videos || (video ? [video] : [])
      if (videos.length > 1) {
        setJavVideoPickerAction('reveal')
        setJavVideoPickerItem(item)
        setJavVideoPickerOpen(true)
        return
      }
      const target = video && isVideoOpenable(video) ? video : videos.find(isVideoOpenable)
      if (!target) return
      revealVideoLocation({
        path: getVideoRelPath(target),
        dirPath: getVideoDirPath(target),
      }).catch((err) => console.error(zh('打开所在位置失败', 'Failed to reveal file'), err))
    },
    [getVideoDirPath, getVideoRelPath, isVideoOpenable]
  )

  const handleSelectJavVideo = useCallback(
    async (video) => {
      if (!video) return
      if (javVideoPickerAction === 'play') {
        handleOpenPlayer(video)
        closeJavVideoPicker()
        return
      }
      if (!isVideoOpenable(video)) return
      const payload = { path: getVideoRelPath(video), dirPath: getVideoDirPath(video) }
      try {
        if (javVideoPickerAction === 'open') {
          await openVideoFile(payload)
        } else if (javVideoPickerAction === 'reveal') {
          await revealVideoLocation(payload)
        }
      } catch (err) {
        console.error(
          javVideoPickerAction === 'open'
            ? zh('打开文件失败', 'Failed to open file')
            : zh('打开所在位置失败', 'Failed to reveal file'),
          err
        )
      } finally {
        closeJavVideoPicker()
      }
    },
    [
      closeJavVideoPicker,
      getVideoDirPath,
      getVideoRelPath,
      isVideoOpenable,
      javVideoPickerAction,
      handleOpenPlayer,
    ]
  )
  useEffect(() => {
    let mounted = true
    loadConfig().finally(() => {
      if (mounted) setConfigLoaded(true)
    })
    return () => {
      mounted = false
    }
  }, [loadConfig])
  const buildVideoUrl = useCallback(
    ({
      page: pageOverride,
      search: searchOverride,
      random: randomOverride,
      seed: seedOverride,
      tagIds: tagIdsOverride,
    } = {}) => {
      const sp = new URLSearchParams()
      sp.set('view', 'video')
      const searchVal = (searchOverride ?? searchTerm).trim()
      if (searchVal) {
        sp.set('search', searchVal)
      }
      if (sortOrder && sortOrder !== 'recent') {
        sp.set('sort', sortOrder)
      }
      const tagIds = tagIdsOverride ?? selectedTagIds
      if (tagIds.length > 0) {
        sp.set('tag_ids', [...tagIds].sort((a, b) => a - b).join(','))
      }
      const randomFlag = randomOverride ?? randomMode
      if (randomFlag) {
        sp.set('random', '1')
        const seedValue = seedOverride ?? randomSeed
        if (seedValue) {
          sp.set('seed', String(seedValue))
        }
      } else {
        sp.delete('random')
        sp.delete('seed')
        const targetPage = pageOverride ?? page
        sp.set('page', String(targetPage))
      }
      const query = sp.toString()
      return `${window.location.pathname}${query ? `?${query}` : ''}`
    },
    [page, randomMode, randomSeed, searchTerm, selectedTagIds, sortOrder]
  )

  const buildJavUrl = useCallback(
    (options = {}) => {
      const {
        page: pageOverride,
        search: searchOverride,
        tab: tabOverride,
        actors: actorsOverride,
        sort: sortOverride,
        tagIds: tagIdsOverride,
        random: randomOverride,
        seed: seedOverride,
      } = options
      const sp = new URLSearchParams()
      sp.set('view', 'jav')
      const tab = tabOverride ?? javTab
      if (tab === 'idol') {
        sp.set('tab', 'idol')
      }
      const searchVal = (searchOverride ?? javSearchTerm).trim()
      if (searchVal) {
        sp.set('search', searchVal)
      }
      const actorList = actorsOverride ?? javActors
      if (actorList && actorList.length > 0) {
        sp.set('actors', actorList.join(','))
      }
      const tagList = tagIdsOverride ?? javTags
      if (tab !== 'idol' && tagList && tagList.length > 0) {
        sp.set('tag_ids', tagList.join(','))
      }
      const hasSortOverride = Object.prototype.hasOwnProperty.call(options, 'sort')
      const normalizedSortOverride = hasSortOverride ? normalizeJavSort(sortOverride, null) : null
      const sortVal =
        tab === 'idol'
          ? String(normalizedSortOverride ?? idolSort ?? '').trim()
          : String(normalizedSortOverride ?? javSort ?? '').trim()
      if (tab === 'idol') {
        if (sortVal && sortVal !== 'work') {
          sp.set('sort', sortVal)
        }
      } else if (sortVal && sortVal !== 'recent') {
        sp.set('sort', sortVal)
      }
      const randomFlag = randomOverride ?? javRandomMode
      if (tab !== 'idol' && randomFlag) {
        sp.set('random', '1')
        const seedValue = seedOverride ?? javRandomSeed
        if (seedValue) {
          sp.set('seed', String(seedValue))
        }
      } else {
        sp.delete('random')
        sp.delete('seed')
        const targetPage = pageOverride ?? (tab === 'idol' ? idolPage : javPage)
        sp.set('page', String(targetPage))
      }
      const query = sp.toString()
      return `${window.location.pathname}${query ? `?${query}` : ''}`
    },
    [
      idolPage,
      javActors,
      javPage,
      javSearchTerm,
      javTab,
      javSort,
      javTags,
      idolSort,
      javRandomMode,
      javRandomSeed,
    ]
  )

  const applyJavTagFilter = useCallback((tagIds) => {
    const clean = Array.from(
      new Set(
        (tagIds || [])
          .map((id) => Number.parseInt(String(id), 10))
          .filter((value) => Number.isFinite(value) && value > 0)
      )
    )
    useStore.setState({
      viewMode: 'jav',
      javTab: 'list',
      javPageSort: '',
      javRandomMode: false,
      javRandomSeed: null,
      javActors: [],
      javTags: clean,
      javSearchTerm: '',
      javPage: 1,
      idolPage: 1,
    })
  }, [])

  const applyUrlState = useCallback(
    (parsed, { fromPopstate = false } = {}) => {
      isPoppingRef.current = fromPopstate
      lastUrlRef.current = window.location.pathname + window.location.search
      const mapTagIdsToNamesFromStore = (ids) => {
        if (!Array.isArray(ids) || ids.length === 0) return []
        const { tags: storeTags } = useStore.getState()
        const idSet = new Set(ids)
        return (storeTags || []).filter((t) => idSet.has(t.id)).map((t) => t.name)
      }
      if (parsed.view === 'jav') {
        const { jav } = parsed
        const currentIdolSort = useStore.getState().idolSort
        useStore.setState({
          viewMode: 'jav',
          javTab: jav.tab,
          javRandomMode: jav.tab === 'idol' ? false : jav.random,
          javRandomSeed: jav.tab === 'idol' ? null : jav.random ? jav.seed : null,
          javSearchTerm: jav.search,
          javActors: jav.actors,
          javTags: jav.tab === 'idol' ? [] : jav.tagIds,
          javPage: jav.random ? 1 : jav.page,
          idolPage: jav.tab === 'idol' ? jav.page : 1,
          javSort: jav.tab === 'idol' ? 'recent' : jav.sort,
          javPageSort: '',
          idolSort: jav.tab === 'idol' ? jav.idolSort : currentIdolSort,
        })
        setJavSearchInput(jav.search)
        if (jav.tab !== 'idol' && jav.random) {
          useStore.getState().loadJavRandom(jav.seed ?? undefined)
        }
        setHydrated(true)
        return
      }

      const { video } = parsed
      useStore.setState({
        viewMode: 'video',
        javPageSort: '',
        sortOrder: video.sort,
        randomMode: video.random,
        randomSeed: video.random ? video.seed : null,
        searchTerm: video.random ? '' : video.search,
        page: video.random ? 1 : video.page,
      })
      setSearchInput(video.search)
      const names = mapTagIdsToNamesFromStore(video.tagIds)
      if (names.length || video.tagIds.length === 0) {
        useStore.getState().setSelectedTags(names, { resetPage: false })
      } else {
        pendingVideoTagIdsRef.current = video.tagIds
      }
      if (video.random) {
        useStore.getState().loadRandom(video.seed ?? undefined)
      }
      setHydrated(true)
    },
    [setJavSearchInput, setSearchInput]
  )

  useEffect(() => {
    const apply = (fromPopstate = false) => {
      const parsed = parseUrlState()
      if (parsed.view === 'jav') {
        useStore.setState({ viewMode: 'jav' })
      } else {
        useStore.setState({ viewMode: 'video' })
      }
      applyUrlState(parsed, { fromPopstate })
    }
    apply(false)
    const onPop = () => apply(true)
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [applyUrlState])

  useEffect(() => {
    setSearchInput(searchTerm)
  }, [searchTerm])

  useEffect(() => {
    setJavSearchInput(javSearchTerm)
  }, [javSearchTerm])

  useEffect(() => {
    loadTags()
    loadDirectories()
    loadJavTags()
  }, [loadTags, loadDirectories, loadJavTags])

  useEffect(() => {
    if (!pendingVideoTagIdsRef.current || !tags.length) return
    const names = mapTagIdsToNames(pendingVideoTagIdsRef.current)
    setSelectedTags(names, { resetPage: false })
    pendingVideoTagIdsRef.current = null
  }, [mapTagIdsToNames, setSelectedTags, tags])

  useEffect(() => {
    if (hydrated && configLoaded && !isJavMode) {
      loadVideos()
    }
  }, [
    configLoaded,
    hydrated,
    isJavMode,
    loadVideos,
    page,
    pageSize,
    randomMode,
    randomSeed,
    searchTerm,
    selectedTags,
    sortOrder,
  ])

  useEffect(() => {
    if (!hydrated || !configLoaded || !isJavMode) return
    if (javTab === 'idol') {
      loadJavIdols()
    } else {
      loadJavs()
    }
  }, [
    hydrated,
    isJavMode,
    javTab,
    javPage,
    javPageSize,
    javSearchTerm,
    javActors,
    javTags,
    javSort,
    javPageSort,
    javRandomMode,
    javRandomSeed,
    idolSort,
    idolPage,
    idolPageSize,
    loadJavs,
    loadJavIdols,
    configLoaded,
  ])

  const forceReloadVideos = useCallback(() => {
    if (!hydrated || !configLoaded) return
    loadVideos({ force: true })
  }, [configLoaded, hydrated, loadVideos])

  const forceReloadJavByTab = useCallback(
    (tab) => {
      if (!hydrated || !configLoaded) return
      if (tab === 'idol') {
        loadJavIdols({ force: true })
      } else {
        loadJavs({ force: true })
      }
    },
    [configLoaded, hydrated, loadJavIdols, loadJavs]
  )

  const currentUrlState = useMemo(
    () =>
      normalizeUrlStateFromStore(
        {
          viewMode,
          page,
          searchTerm,
          sortOrder,
          selectedTags,
          randomMode,
          randomSeed,
          javTab,
          javPage,
          javSearchTerm,
          javActors,
          javTags,
          javSort,
          javRandomMode,
          javRandomSeed,
          idolSort,
          idolPage,
        },
        tagsByName
      ),
    [
      idolPage,
      idolSort,
      javActors,
      javPage,
      javRandomMode,
      javRandomSeed,
      javSearchTerm,
      javSort,
      javTab,
      javTags,
      page,
      randomMode,
      randomSeed,
      searchTerm,
      selectedTags,
      sortOrder,
      tagsByName,
      viewMode,
    ]
  )

  useEffect(() => {
    if (!hydrated) return
    const nextUrl = buildUrlFromState(currentUrlState)
    const currentUrl = window.location.pathname + window.location.search
    if (nextUrl === currentUrl) {
      lastUrlRef.current = nextUrl
      isPoppingRef.current = false
      return
    }
    if (isPoppingRef.current) {
      lastUrlRef.current = nextUrl
      isPoppingRef.current = false
      return
    }
    window.history.pushState({}, '', nextUrl)
    lastUrlRef.current = nextUrl
  }, [currentUrlState, hydrated])

  const canPrev = page > 1
  const canNext = hasNext
  const lastPage = Math.max(1, Math.ceil((total || 0) / pageSize))

  const navigateVideoPage = useCallback(
    (targetPage) => {
      if (!targetPage || targetPage === page) return
      setPage(targetPage)
    },
    [page, setPage]
  )
  const selectedCount = useMemo(() => selectedVideoIds.size, [selectedVideoIds])
  const selectedList = useMemo(() => {
    const ids = Array.from(selectedVideoIds)
    return ids.map((id) => {
      const v = videos.find((item) => item.id === id)
      const labelFromMeta = selectedVideoMeta?.[id]
      return {
        id,
        label: labelFromMeta || v?.filename || v?.path || `#${id}`,
        video: v,
      }
    })
  }, [selectedVideoIds, videos, selectedVideoMeta])
  const javLastPage = Math.max(1, Math.ceil((javTotal || 0) / javPageSize))
  const javHasPrev = javPage > 1
  const javHasNext = javPage < javLastPage
  const idolLastPage = Math.max(1, Math.ceil((idolTotal || 0) / idolPageSize))
  const idolHasPrev = idolPage > 1
  const idolHasNext = idolPage < idolLastPage
  const javTagNameMap = useMemo(
    () => new Map((javTagOptions || []).map((tag) => [tag.id, tag.name])),
    [javTagOptions]
  )
  const javUserTagOptions = useMemo(
    () => (javTagOptions || []).filter((tag) => isUserJavTag(tag)),
    [javTagOptions]
  )
  const searchHref = buildVideoUrl({ search: searchInput, page: 1, random: false, tagIds: [] })
  const randomHref = buildVideoUrl({ random: true, page: 1, tagIds: [], search: '' })
  const javSearchHref = buildJavUrl({
    search: javSearchInput,
    page: 1,
    tab: javTab,
    actors: [],
    tagIds: [],
    random: false,
  })
  const javRandomHref = buildJavUrl({
    random: true,
    page: 1,
    tab: 'list',
    actors: [],
    tagIds: [],
    search: '',
  })
  const handleJavRandomClick = useCallback(() => {
    const nextSeed = generateRandomSeed()
    useStore.setState({
      viewMode: 'jav',
      javTab: 'list',
      javPageSort: '',
      javActors: [],
      javTags: [],
      javSearchTerm: '',
      javPage: 1,
      idolPage: 1,
    })
    setJavSearchInput('')
    loadJavRandom(nextSeed)
  }, [loadJavRandom])

  const handleVideoRandomClick = useCallback(() => {
    const nextSeed = generateRandomSeed()
    setSearchInput('')
    useStore.setState({
      viewMode: 'video',
      selectedTags: [],
      searchTerm: '',
      page: 1,
      randomMode: true,
      randomSeed: nextSeed,
    })
  }, [])
  const filterSummary = useMemo(() => {
    const formatList = (items) => {
      if (!items || items.length === 0) return ''
      const separator = isChineseLocale() ? '、' : ', '
      return items.join(separator)
    }
    if (isJavMode) {
      const parts = []
      if (javTab === 'list') {
        const actorsLabel = formatList(javActors)
        if (actorsLabel) parts.push(zh(`女优: ${actorsLabel}`, `Idols: ${actorsLabel}`))
        const tagNames = javTags.map((id) => javTagNameMap.get(id)).filter(Boolean)
        const tagsLabel = formatList(tagNames)
        if (tagsLabel) parts.push(zh(`标签: ${tagsLabel}`, `Tags: ${tagsLabel}`))
        if (javRandomMode) parts.push(zh('随机', 'Random'))
      }
      const searchLabel = (javSearchTerm || '').trim()
      if (searchLabel) parts.push(zh(`搜索: ${searchLabel}`, `Search: ${searchLabel}`))
      return parts.length ? parts.join(isChineseLocale() ? '；' : '; ') : ''
    }
    const parts = []
    const tagsLabel = formatList(selectedTags)
    if (tagsLabel) parts.push(zh(`标签: ${tagsLabel}`, `Tags: ${tagsLabel}`))
    const searchLabel = (searchTerm || '').trim()
    if (searchLabel) parts.push(zh(`搜索: ${searchLabel}`, `Search: ${searchLabel}`))
    if (randomMode) {
      parts.push(zh('随机', 'Random'))
    }
    return parts.length ? parts.join(isChineseLocale() ? '；' : '; ') : ''
  }, [
    isJavMode,
    javTab,
    javActors,
    javTags,
    javTagNameMap,
    javSearchTerm,
    javRandomMode,
    selectedTags,
    searchTerm,
    randomMode,
  ])

  const submitSearch = (e) => {
    e?.preventDefault()
    const nextSearch = (searchInput || '').trim()
    useStore.setState({
      viewMode: 'video',
      selectedTags: [],
      searchTerm: nextSearch,
      page: 1,
      randomMode: false,
      randomSeed: null,
    })
  }

  const submitJavSearch = (e) => {
    e?.preventDefault()
    useStore.setState({
      viewMode: 'jav',
      javPageSort: '',
      javRandomMode: false,
      javRandomSeed: null,
      javActors: [],
      javTags: [],
      javSearchTerm: (javSearchInput || '').trim(),
      javPage: 1,
      idolPage: 1,
    })
  }

  const handleSaveVideoSettings = async () => {
    const size = Math.max(1, parseInt(videoPageSizeInput, 10) || pageSize)
    const normalizedSort =
      videoSortInput === 'filename'
        ? 'filename'
        : videoSortInput === 'duration'
          ? 'duration'
          : videoSortInput === 'play_count'
            ? 'play_count'
            : 'recent'
    try {
      await updateConfig({ video_page_size: size, video_sort: normalizedSort })
      const prevPage = page
      // ensure current page does not exceed last page after page size change
      const lastPage = Math.max(1, Math.ceil((total || 0) / size))
      const nextPage = prevPage > lastPage ? lastPage : prevPage

      useStore.setState({
        pageSize: size,
        sortOrder: normalizedSort,
        page: nextPage,
        randomMode: false,
        randomSeed: null,
      })
      setVideoSettingsOpen(false)
    } catch (err) {
      alert(err.message || zh('保存失败', 'Save failed'))
    }
  }

  const handleSaveJavSettings = async () => {
    const javSize = Math.max(1, parseInt(javPageSizeInput, 10) || javPageSize)
    const idolSize = Math.max(1, parseInt(idolPageSizeInput, 10) || idolPageSize)
    const normalizedSort = normalizeJavSort(javSortInput)
    const normalizedIdolSort =
      idolSortInput === 'birth'
        ? 'birth'
        : idolSortInput === 'height'
          ? 'height'
          : idolSortInput === 'bust'
            ? 'bust'
            : idolSortInput === 'hips'
              ? 'hips'
              : idolSortInput === 'waist'
                ? 'waist'
                : idolSortInput === 'cup'
                  ? 'cup'
                  : 'work'
    try {
      await updateConfig({
        jav_page_size: javSize,
        idol_page_size: idolSize,
        jav_sort: normalizedSort,
        idol_sort: normalizedIdolSort,
      })
      const prevJavPage = javPage
      const prevIdolPage = idolPage
      const javLast = Math.max(1, Math.ceil((javTotal || 0) / javSize))
      const idolLast = Math.max(1, Math.ceil((idolTotal || 0) / idolSize))
      useStore.setState({
        javPageSize: javSize,
        idolPageSize: idolSize,
        javSort: normalizedSort,
        javPageSort: '',
        idolSort: normalizedIdolSort,
        javPage: Math.min(prevJavPage, javLast),
        idolPage: Math.min(prevIdolPage, idolLast),
        javRandomMode: false,
        javRandomSeed: null,
      })
      setJavSettingsOpen(false)
    } catch (err) {
      alert(err.message || zh('保存失败', 'Save failed'))
    }
  }

  useEffect(() => {
    if (videoSettingsOpen) {
      setVideoPageSizeInput(pageSize)
      setVideoSortInput(sortOrder)
    }
  }, [videoSettingsOpen, pageSize, sortOrder])

  useEffect(() => {
    if (javSettingsOpen) {
      setJavPageSizeInput(javPageSize)
      setIdolPageSizeInput(idolPageSize)
      setJavSortInput(javSort)
      setIdolSortInput(idolSort)
    }
  }, [javSettingsOpen, javPageSize, idolPageSize, javSort, idolSort])

  useEffect(() => {
    if (selectedCount !== 0) return
    setSelectionOpsOpen(false)
    setSelectionTagsOpen(false)
    setSelectionTagAction('add')
    setSelectionTagChoices([])
  }, [selectedCount])

  const openTagEditor = useCallback(
    (videoId) => {
      setTagPickerFor(videoId)
      const target = videos.find((v) => v.id === videoId)
      const initial = Array.isArray(target?.tags) ? target.tags.map((t) => String(t.id)) : []
      setTagPickerSelected(initial)
    },
    [videos]
  )

  const openJavTagEditor = useCallback(
    (item) => {
      if (!item) return
      const initial = Array.isArray(item?.tags)
        ? item.tags.filter((tag) => isUserJavTag(tag)).map((tag) => String(tag.id))
        : []
      setJavTagPickerItem(item)
      setJavTagPickerSelected(initial)
      loadJavTags()
    },
    [loadJavTags]
  )

  const tagPickerExisting = useMemo(() => {
    if (!tagPickerFor) return []
    const target = videos.find((v) => v.id === tagPickerFor)
    return Array.isArray(target?.tags) ? target.tags.map((t) => String(t.id)) : []
  }, [tagPickerFor, videos])

  const tagPickerDirty = useMemo(() => {
    if (!tagPickerFor) return false
    const current = new Set(tagPickerExisting)
    const selected = new Set(tagPickerSelected)
    if (current.size !== selected.size) return true
    for (const id of current) {
      if (!selected.has(id)) return true
    }
    return false
  }, [tagPickerExisting, tagPickerFor, tagPickerSelected])

  const javTagPickerExisting = useMemo(() => {
    if (!javTagPickerItem) return []
    return Array.isArray(javTagPickerItem?.tags)
      ? javTagPickerItem.tags.filter((tag) => isUserJavTag(tag)).map((tag) => String(tag.id))
      : []
  }, [javTagPickerItem])

  const javTagPickerDirty = useMemo(() => {
    if (!javTagPickerItem) return false
    const current = new Set(javTagPickerExisting)
    const selected = new Set(javTagPickerSelected)
    if (current.size !== selected.size) return true
    for (const id of current) {
      if (!selected.has(id)) return true
    }
    return false
  }, [javTagPickerExisting, javTagPickerItem, javTagPickerSelected])

  const handleApplyTags = async () => {
    if (!tagPickerFor) {
      setTagPickerFor(null)
      setTagPickerSelected([])
      return
    }
    const selectedIds = tagPickerSelected.map((t) => Number(t)).filter(Boolean)
    if (!tagPickerDirty) {
      setTagPickerFor(null)
      setTagPickerSelected([])
      return
    }
    try {
      await replaceTagsForVideos([tagPickerFor], selectedIds)
      useStore.setState((state) => {
        if (!Array.isArray(state.videos)) return {}
        const tagLookup = new Map((tags || []).map((tag) => [tag.id, tag]))
        const nextVideos = state.videos.map((video) => {
          if (video.id !== tagPickerFor) return video
          const nextTags = selectedIds.map((id) => tagLookup.get(id)).filter(Boolean)
          return { ...video, tags: nextTags }
        })
        return { videos: nextVideos }
      })
    } catch (err) {
      console.error('update tags failed', err)
    } finally {
      setTagPickerFor(null)
      setTagPickerSelected([])
    }
  }

  const handleApplyJavTags = async () => {
    if (!javTagPickerItem) {
      setJavTagPickerItem(null)
      setJavTagPickerSelected([])
      return
    }
    const javId = javTagPickerItem.id
    if (!javId) {
      setJavTagPickerItem(null)
      setJavTagPickerSelected([])
      return
    }
    const selectedIds = javTagPickerSelected.map((t) => Number(t)).filter(Boolean)
    if (!javTagPickerDirty) {
      setJavTagPickerItem(null)
      setJavTagPickerSelected([])
      return
    }
    try {
      await replaceJavTagsForItems([javId], selectedIds)
      useStore.setState((state) => {
        if (!Array.isArray(state.javItems)) return {}
        const userTagMap = new Map(javUserTagOptions.map((tag) => [tag.id, tag]))
        const next = state.javItems.map((item) => {
          if (item.id !== javId) return item
          const existingTags = Array.isArray(item.tags) ? item.tags : []
          for (const tag of existingTags) {
            if (isUserJavTag(tag) && !userTagMap.has(tag.id)) {
              userTagMap.set(tag.id, tag)
            }
          }
          const nextUserTags = selectedIds.map((id) => userTagMap.get(id)).filter(Boolean)
          const nonUserTags = existingTags.filter((tag) => !isUserJavTag(tag))
          return { ...item, tags: [...nonUserTags, ...nextUserTags] }
        })
        return { javItems: next }
      })
    } catch (err) {
      console.error('update jav tags failed', err)
    } finally {
      setJavTagPickerItem(null)
      setJavTagPickerSelected([])
    }
  }

  const handleTagPickerClose = () => {
    setTagPickerFor(null)
    setTagPickerSelected([])
  }

  const handleJavTagPickerClose = () => {
    setJavTagPickerItem(null)
    setJavTagPickerSelected([])
  }

  const handleTagPickerToggle = (tagId, checked) => {
    setTagPickerSelected((prev) => {
      const set = new Set(prev)
      if (checked) set.add(String(tagId))
      else set.delete(String(tagId))
      return Array.from(set)
    })
  }

  const handleJavTagPickerToggle = (tagId, checked) => {
    setJavTagPickerSelected((prev) => {
      const set = new Set(prev)
      if (checked) set.add(String(tagId))
      else set.delete(String(tagId))
      return Array.from(set)
    })
  }

  const handleSelectionTagsClose = () => {
    setSelectionTagsOpen(false)
    setSelectionTagAction('add')
    setSelectionTagChoices([])
  }

  const handleSelectionTagChoiceToggle = (tagId, checked) => {
    setSelectionTagChoices((prev) => {
      const set = new Set(prev)
      if (checked) set.add(String(tagId))
      else set.delete(String(tagId))
      return Array.from(set)
    })
  }

  const handleApplySelectionTags = async () => {
    const ids = selectionTagChoices.map((t) => Number(t)).filter(Boolean)
    const vidIds = Array.from(selectedVideoIds)
    try {
      if (selectionTagAction === 'remove') {
        await Promise.all(ids.map((tid) => removeTagFromVideos(tid, vidIds)))
        const removedIds = new Set(ids)
        useStore.setState(({ videos }) => {
          const next = videos.map((v) => {
            if (!vidIds.includes(v.id)) return v
            const existing = Array.isArray(v.tags) ? v.tags : []
            const nextTags = existing.filter((tag) => !removedIds.has(tag.id))
            return nextTags.length === existing.length ? v : { ...v, tags: nextTags }
          })
          return { videos: next }
        })
      } else {
        await Promise.all(ids.map((tid) => addTagToVideos(tid, vidIds)))
        const addedTags = tags.filter((t) => ids.includes(t.id))
        useStore.setState(({ videos }) => {
          const next = videos.map((v) => {
            if (!vidIds.includes(v.id)) return v
            const existing = Array.isArray(v.tags) ? v.tags : []
            const mergedById = new Map()
            for (const tag of existing) mergedById.set(tag.id, tag)
            for (const tag of addedTags) mergedById.set(tag.id, tag)
            return { ...v, tags: Array.from(mergedById.values()) }
          })
          return { videos: next }
        })
      }
    } catch (err) {
      console.error(`${selectionTagAction} tags for selection failed`, err)
    } finally {
      setSelectionTagsOpen(false)
      setSelectionTagAction('add')
      setSelectionTagChoices([])
      setSelectionOpsOpen(false)
      clearSelection()
    }
  }

  const handleHomeClick = () => {
    setTagModalOpen(false)
    setVideoSettingsOpen(false)
    setJavSettingsOpen(false)
    setGlobalSettingsOpen(false)
    if (isJavMode) {
      useStore.setState({
        viewMode: 'jav',
        javPageSort: '',
        javRandomMode: false,
        javRandomSeed: null,
        javActors: [],
        javTags: [],
        javSearchTerm: '',
        javPage: 1,
        idolPage: 1,
      })
      setJavSearchInput('')
      forceReloadJavByTab(javTab)
    } else {
      useStore.setState({
        viewMode: 'video',
        randomMode: false,
        randomSeed: null,
        selectedTags: [],
        searchTerm: '',
        page: 1,
        selectedVideoIds: new Set(),
        selectedVideoMeta: {},
      })
      setSearchInput('')
      forceReloadVideos()
    }
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }

  const handleSwitchToJav = () => {
    const targetTab = javTab === 'idol' ? 'idol' : 'list'
    useStore.setState({ viewMode: 'jav', javTab: targetTab, javPageSort: '' })
    forceReloadJavByTab(targetTab)
  }

  const handleSwitchJavTab = (tab) => {
    const nextTab = tab === 'idol' ? 'idol' : 'list'
    const nextRandomMode = nextTab === 'idol' ? false : javRandomMode
    const nextRandomSeed = nextTab === 'idol' ? null : javRandomSeed
    useStore.setState({
      javTab: nextTab,
      javPageSort: '',
      javActors: [],
      javTags: [],
      javRandomMode: nextRandomMode,
      javRandomSeed: nextRandomSeed,
      javPage: 1,
      idolPage: 1,
    })
    forceReloadJavByTab(nextTab)
  }

  const handleToggleMode = () => {
    if (isJavMode) {
      setViewMode('video')
      forceReloadVideos()
    } else {
      handleSwitchToJav()
    }
  }

  const handleSelectIdol = (idol) => {
    if (!idol || !idol.name) return
    useStore.setState({
      viewMode: 'jav',
      javTab: 'list',
      javPageSort: '',
      javRandomMode: false,
      javRandomSeed: null,
      javActors: [idol.name],
      javTags: [],
      javSearchTerm: '',
      javPage: 1,
      idolPage: 1,
    })
  }

  const handleJavActorClick = useCallback((name) => {
    const trimmed = (name || '').trim()
    if (!trimmed) return
    useStore.setState({
      viewMode: 'jav',
      javTab: 'list',
      javPageSort: '',
      javRandomMode: false,
      javRandomSeed: null,
      javActors: [trimmed],
      javTags: [],
      javSearchTerm: '',
      javPage: 1,
      idolPage: 1,
    })
  }, [])

  const handleJavTagClick = useCallback(
    (tag) => {
      const raw = typeof tag === 'object' ? tag?.id : tag
      const parsed = Number.parseInt(String(raw), 10)
      if (!Number.isFinite(parsed) || parsed <= 0) return
      applyJavTagFilter([parsed])
    },
    [applyJavTagFilter]
  )

  const handleOpenJavTagModal = useCallback(() => {
    setJavTagModalOpen(true)
    loadJavTags()
  }, [loadJavTags])

  const handleToggleSelectPage = useCallback(() => {
    if (!Array.isArray(videos) || videos.length === 0) return
    useStore.setState((state) => {
      const pageIds = videos.map((video) => video?.id).filter(Boolean)
      if (pageIds.length === 0) return {}
      const nextIds = new Set(state.selectedVideoIds)
      const nextMeta = { ...state.selectedVideoMeta }
      const allSelected = pageIds.every((id) => nextIds.has(id))
      if (allSelected) {
        pageIds.forEach((id) => {
          nextIds.delete(id)
          delete nextMeta[id]
        })
      } else {
        videos.forEach((video) => {
          if (!video?.id) return
          nextIds.add(video.id)
          nextMeta[video.id] = video.filename || video.path || `#${video.id}`
        })
      }
      return { selectedVideoIds: nextIds, selectedVideoMeta: nextMeta }
    })
  }, [videos])

  const activeError = isJavMode ? (javTab === 'idol' ? idolError : javError) : error
  const showDirectorySetupHint =
    hydrated &&
    configLoaded &&
    !loading &&
    !activeError &&
    Array.isArray(directories) &&
    directories.length === 0 &&
    Array.isArray(videos) &&
    videos.length === 0

  const activeJavLoading = javTab === 'idol' ? idolLoading : javLoading
  const javVideoPickerTitle =
    javVideoPickerAction === 'open'
      ? zh('选择打开文件', 'Choose a file to open')
      : javVideoPickerAction === 'reveal'
        ? zh('选择定位文件', 'Choose a file to reveal')
        : zh('选择播放文件', 'Choose a file to play')
  const javVideoPickerEmptyText =
    javVideoPickerAction === 'play'
      ? zh('暂无可播放文件', 'No playable files')
      : zh('暂无可用文件', 'No available files')

  return (
    <div className="min-h-screen">
      <TopBar
        onHome={handleHomeClick}
        isJavMode={isJavMode}
        onToggleMode={handleToggleMode}
        videoSearchInput={searchInput}
        onVideoSearchInputChange={setSearchInput}
        onSubmitVideoSearch={submitSearch}
        videoSearchHref={searchHref}
        randomMode={randomMode}
        randomHref={randomHref}
        onRandomClick={handleVideoRandomClick}
        onOpenTagModal={handleOpenTagModal}
        onOpenJavTagModal={handleOpenJavTagModal}
        onOpenVideoSettings={() => setVideoSettingsOpen(true)}
        onOpenJavSettings={() => setJavSettingsOpen(true)}
        onOpenGlobalSettings={() => setGlobalSettingsOpen(true)}
        javSearchInput={javSearchInput}
        onJavSearchInputChange={setJavSearchInput}
        onSubmitJavSearch={submitJavSearch}
        javSearchHref={javSearchHref}
        javRandomHref={javRandomHref}
        javRandomMode={javRandomMode}
        onJavRandomClick={handleJavRandomClick}
        isModifiedClick={isModifiedClick}
        javTab={javTab}
        onSwitchJavTab={handleSwitchJavTab}
        filterSummary={filterSummary}
        showDirectorySetupHint={showDirectorySetupHint}
      />

      <main className="mx-auto max-w-screen-2xl px-6 pb-6 pt-0">
        {activeError && (
          <div
            role="alert"
            className="mb-4 rounded border border-red-200 bg-red-50 p-3 text-red-700"
          >
            {String(activeError)}
          </div>
        )}

        {isJavMode ? (
          javTab === 'idol' ? (
            <JavIdolView
              page={idolPage}
              lastPage={idolLastPage}
              hasPrev={idolHasPrev}
              hasNext={idolHasNext}
              loading={idolLoading}
              buildPageUrl={({ page: targetPage }) =>
                buildJavUrl({ page: targetPage, tab: 'idol' })
              }
              buildIdolUrl={(idol) =>
                buildJavUrl({ page: 1, search: '', tab: 'list', actors: [idol.name], tagIds: [] })
              }
              onFirst={() => setIdolPage(1)}
              onPrev={() => idolHasPrev && setIdolPage(idolPage - 1)}
              onGoToPage={(p) => setIdolPage(p)}
              onNext={() => idolHasNext && setIdolPage(idolPage + 1)}
              onLast={() => setIdolPage(idolLastPage)}
              items={idolItems}
              onSelectIdol={handleSelectIdol}
            />
          ) : (
            <JavView
              javPage={javPage}
              javLastPage={javLastPage}
              javHasPrev={javHasPrev}
              javHasNext={javHasNext}
              javLoading={activeJavLoading}
              javRandomMode={javRandomMode}
              javPageSort={javPageSort}
              javGlobalSort={javSort}
              buildJavUrl={buildJavUrl}
              setJavPage={setJavPage}
              setJavPageSort={setJavPageSort}
              javItems={javItems}
              onPlay={handleJavPlay}
              onOpenFile={handleJavOpenFile}
              onRevealFile={handleJavRevealFile}
              onIdolClick={handleJavActorClick}
              onTagClick={handleJavTagClick}
              onEditTags={openJavTagEditor}
            />
          )
        ) : (
          <VideoView
            selectedCount={selectedCount}
            clearSelection={clearSelection}
            setSelectionOpsOpen={setSelectionOpsOpen}
            page={page}
            lastPage={lastPage}
            canPrev={canPrev}
            canNext={canNext}
            loading={loading}
            randomMode={randomMode}
            buildVideoUrl={buildVideoUrl}
            setPage={navigateVideoPage}
            goToLastPage={() => navigateVideoPage(lastPage)}
            videos={videos}
            selectedVideoIds={selectedVideoIds}
            toggleSelectVideo={toggleSelectVideo}
            onToggleSelectPage={handleToggleSelectPage}
            openPlayer={handleOpenPlayer}
            setTagPickerFor={openTagEditor}
            onTagClick={handleVideoTagClick}
          />
        )}
      </main>

      <VideoSettingsModal
        open={videoSettingsOpen}
        onClose={() => setVideoSettingsOpen(false)}
        pageSizeInput={videoPageSizeInput}
        onPageSizeChange={setVideoPageSizeInput}
        sortInput={videoSortInput}
        onSortChange={setVideoSortInput}
        onSave={handleSaveVideoSettings}
      />

      <JavSettingsModal
        open={javSettingsOpen}
        onClose={() => setJavSettingsOpen(false)}
        javPageSizeInput={javPageSizeInput}
        onJavPageSizeChange={setJavPageSizeInput}
        idolPageSizeInput={idolPageSizeInput}
        onIdolPageSizeChange={setIdolPageSizeInput}
        javSortInput={javSortInput}
        onJavSortChange={setJavSortInput}
        idolSortInput={idolSortInput}
        onIdolSortChange={setIdolSortInput}
        onSave={handleSaveJavSettings}
      />

      <JavVideoPickerModal
        open={javVideoPickerOpen}
        title={javVideoPickerTitle}
        onClose={closeJavVideoPicker}
        item={javVideoPickerItem}
        choices={javVideoChoices}
        emptyText={javVideoPickerEmptyText}
        action={javVideoPickerAction}
        buildVideoFullPath={buildVideoFullPath}
        isVideoOpenable={isVideoOpenable}
        onSelectVideo={handleSelectJavVideo}
      />

      <SelectionOpsModal
        open={selectionOpsOpen}
        onClose={() => setSelectionOpsOpen(false)}
        selectedList={selectedList}
        selectedCount={selectedCount}
        onOpenTags={() => {
          loadTags()
          setSelectionTagAction('add')
          setSelectionTagChoices([])
          setSelectionOpsOpen(false)
          setSelectionTagsOpen(true)
        }}
        onOpenRemoveTags={() => {
          loadTags()
          setSelectionTagAction('remove')
          setSelectionTagChoices([])
          setSelectionOpsOpen(false)
          setSelectionTagsOpen(true)
        }}
      />

      <SelectionTagsModal
        open={selectionTagsOpen}
        onClose={handleSelectionTagsClose}
        tags={tags}
        action={selectionTagAction}
        selectedChoices={selectionTagChoices}
        onToggleChoice={handleSelectionTagChoiceToggle}
        onConfirm={handleApplySelectionTags}
        confirmDisabled={!selectionTagChoices.length || selectedVideoIds.size === 0}
      />

      <TagPickerModal
        open={Boolean(tagPickerFor)}
        tags={tags}
        selectedIds={tagPickerSelected}
        onToggleChoice={handleTagPickerToggle}
        onClose={handleTagPickerClose}
        onSave={handleApplyTags}
        saveDisabled={!tagPickerDirty}
      />
      <TagPickerModal
        open={Boolean(javTagPickerItem)}
        tags={javUserTagOptions}
        selectedIds={javTagPickerSelected}
        onToggleChoice={handleJavTagPickerToggle}
        onClose={handleJavTagPickerClose}
        onSave={handleApplyJavTags}
        saveDisabled={!javTagPickerDirty}
      />

      <VideoTagModal
        open={tagModalOpen}
        onClose={() => setTagModalOpen(false)}
        tags={tags}
        onToggleFilter={(name) => toggleTagFilter(name)}
        onCreateTag={async (name) => {
          await createTag(name)
          await loadTags()
        }}
        onRenameTag={async (id, name) => {
          const oldName = tags.find((t) => t.id === id)?.name || ''
          await renameTag(id, name)
          useStore.setState((state) => {
            const nextTags = Array.isArray(state.tags)
              ? state.tags.map((tag) => (tag.id === id ? { ...tag, name } : tag))
              : state.tags
            const nextVideos = Array.isArray(state.videos)
              ? state.videos.map((video) => {
                  if (!Array.isArray(video.tags)) return video
                  const nextVideoTags = video.tags.map((tag) =>
                    tag.id === id ? { ...tag, name } : tag
                  )
                  return nextVideoTags === video.tags ? video : { ...video, tags: nextVideoTags }
                })
              : state.videos
            const nextSelectedTags =
              oldName && Array.isArray(state.selectedTags)
                ? state.selectedTags.map((tagName) => (tagName === oldName ? name : tagName))
                : state.selectedTags
            return { tags: nextTags, videos: nextVideos, selectedTags: nextSelectedTags }
          })
          await loadTags()
        }}
        onDeleteTag={async (tag) => {
          const id = typeof tag === 'object' ? tag?.id : tag
          if (!id) return
          const name =
            typeof tag === 'object' ? tag?.name : tags.find((item) => item.id === id)?.name || ''
          await deleteTag(id)
          if (name) {
            useStore.setState((state) => ({
              selectedTags: state.selectedTags.filter((tagName) => tagName !== name),
            }))
          }
          await loadTags()
        }}
        onApplyTagFilter={(names) => {
          setSearchTerm('', { resetPage: false, triggerLoad: false })
          setSelectedTags(names)
        }}
      />
      <JavTagModal
        open={javTagModalOpen}
        onClose={() => setJavTagModalOpen(false)}
        tags={javTagOptions}
        onApplyTagFilter={applyJavTagFilter}
        onCreateTag={async (name) => {
          await createJavTag(name)
          await loadJavTags()
        }}
        onRenameTag={async (id, name) => {
          await renameJavTag(id, name)
          useStore.setState((state) => {
            const options = Array.isArray(state.javTagOptions) ? state.javTagOptions : []
            const items = Array.isArray(state.javItems) ? state.javItems : []
            const nextOptions = options.map((tag) => (tag.id === id ? { ...tag, name } : tag))
            const nextItems = items.map((item) => {
              if (!Array.isArray(item.tags)) return item
              const nextTags = item.tags.map((tag) => (tag.id === id ? { ...tag, name } : tag))
              return nextTags === item.tags ? item : { ...item, tags: nextTags }
            })
            return { javTagOptions: nextOptions, javItems: nextItems }
          })
          await loadJavTags()
        }}
        onDeleteTag={async (tag) => {
          const id = typeof tag === 'object' ? tag?.id : tag
          if (!id) return
          await deleteJavTag(id)
          useStore.setState((state) => {
            const nextOptions = Array.isArray(state.javTagOptions)
              ? state.javTagOptions.filter((item) => item.id !== id)
              : state.javTagOptions
            const nextItems = Array.isArray(state.javItems)
              ? state.javItems.map((item) => {
                  if (!Array.isArray(item.tags)) return item
                  const nextTags = item.tags.filter((tagItem) => tagItem.id !== id)
                  return nextTags === item.tags ? item : { ...item, tags: nextTags }
                })
              : state.javItems
            const nextFilters = Array.isArray(state.javTags)
              ? state.javTags.filter((tagId) => tagId !== id)
              : state.javTags
            return { javTagOptions: nextOptions, javItems: nextItems, javTags: nextFilters }
          })
          await loadJavTags()
        }}
      />
      <GlobalSettingsModal
        open={globalSettingsOpen}
        onClose={() => setGlobalSettingsOpen(false)}
        directories={directories}
        onCreateDirectory={async (payload) => {
          const created = await createDirectory(payload)
          await loadDirectories()
          showToast(
            zh(
              '目录添加成功，首次扫描目录里的视频需要一定时间，请耐心等待，您可手动刷新页面查看扫描进度',
              'Directory added. The first scan may take some time. You can refresh manually to check progress.'
            )
          )
          return created
        }}
        onUpdateDirectory={async (id, payload) => {
          const updated = await updateDirectory(id, payload)
          await loadDirectories()
          return updated
        }}
        onDeleteDirectory={async (id) => {
          const deleted = await deleteDirectory(id)
          await loadDirectories()
          return deleted
        }}
        proxyPort={Number.parseInt(config?.proxy_port, 10) || 0}
        onSaveProxyPort={async (port) => {
          const cfg = await updateConfig({ proxy_port: port })
          useStore.setState({ config: cfg })
        }}
        playerWindowWidth={
          Number.parseInt(config?.player_window_width, 10) ||
          Number.parseInt(config?.player_window_size, 10) ||
          70
        }
        playerWindowHeight={
          Number.parseInt(config?.player_window_height, 10) ||
          Number.parseInt(config?.player_window_size, 10) ||
          70
        }
        playerWindowUseAutofit={
          config?.player_window_use_autofit == null
            ? false
            : !['0', 'false', 'no', 'off'].includes(
                String(config.player_window_use_autofit).trim().toLowerCase()
              )
        }
        playerOntop={
          config?.player_ontop == null
            ? true
            : !['0', 'false', 'no', 'off'].includes(
                String(config.player_ontop).trim().toLowerCase()
              )
        }
        playerVolume={
          config?.player_volume === '0' ? 0 : Number.parseInt(config?.player_volume, 10) || 70
        }
        onSavePlayerBasicSettings={async (payload) => {
          const cfg = await updateConfig(payload)
          useStore.setState({ config: cfg })
        }}
        playerHotkeys={config?.player_hotkeys}
        onSavePlayerHotkeys={async (hotkeys) => {
          const cfg = await updateConfig({ player_hotkeys: hotkeys })
          useStore.setState({ config: cfg })
        }}
      />
      <Toast
        open={Boolean(toastMessage)}
        message={toastMessage}
        onClose={() => setToastMessage('')}
      />
    </div>
  )
}
