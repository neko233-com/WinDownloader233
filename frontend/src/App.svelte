<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import type { Tool, Progress, LogEntry, Category } from './lib/types';
  import { ALL_CATEGORIES, CATEGORY_ICONS } from './lib/types';
  import type { Driver } from '../bindings/github.com/neko233/WinDownloader233/drivers/models';
  import type { PackageAction, PackageInfo, PackageManagerInfo } from '../bindings/github.com/neko233/WinDownloader233/installer/models';
  import Header from './components/Header.svelte';
  import ToolList from './components/ToolList.svelte';
  import SettingsPanel from './components/SettingsPanel.svelte';
  import { AppService } from '../bindings/github.com/neko233/WinDownloader233';

  type View = 'toolkit' | 'discover' | 'installed' | 'updates' | 'bundle' | 'drivers';
  const FALLBACK_UI: Record<string, string> = {
    search: '搜索工具...', install: '安装', installed: '已安装', installing: '安装中...', download: '下载',
    downloading: '下载中...', uninstall: '卸载', settings: '设置', refresh: '刷新列表', free: '免费',
    paid: '付费', version: '版本', size: '大小', homepage: '官网', no_results: '没有找到匹配的工具',
    app_subtitle: '软件、驱动、工具集、备份迁移一体化管理',
    view_discover: '搜索软件', view_installed: '已安装', view_updates: '可更新', view_drivers: '驱动下载',
    view_bundle: '备份迁移', view_toolkit: '工具集', title_toolkit: '游戏开发工具集',
    title_drivers: '驱动下载', title_packages: '软件包管理', desc_toolkit: '按分类安装常用开发工具',
    desc_drivers: '主板、显卡、鼠标、键盘、音频、网卡等官方驱动入口',
    desc_packages: '搜索、安装、更新、卸载本机软件包', loading: '加载中', all_sources: '全部来源',
    all: '全部', manager: '来源', name: '名称', available: '可用版本', action: '操作',
    running: '运行中', update: '更新', search_btn: '搜索', install_selected: '安装选中',
    update_selected: '更新选中', uninstall_selected: '卸载选中', force_refresh: '强制刷新',
    check_updates: '检查更新', export_backup: '导出备份', export_installed: '导出已安装',
    parse_import: '解析导入', install_imported: '安装导入项',
    package_placeholder: '输入软件名，例如 vscode、git、python', package_options: '可选安装参数',
    package_empty: '输入关键词后搜索包', no_package_results: '没有结果',
    bundle_placeholder: '粘贴 WinDownloader233 备份 JSON',
    driver_placeholder: '搜索厂商 / 型号 / 别名，例如 RTX、ROG、G502、Realtek、声卡',
    driver_empty: '没有匹配驱动。试试厂商名、型号、设备类型或中文别名。',
    official_download: '官方下载', support_page: '支持页', tasks: '任务日志',
    background_tasks: '后台任务', task_count: '个任务', running_count: '个运行中',
    no_tasks: '暂无后台任务', logs: '日志', refresh_logs: '刷新', copy_logs: '复制日志',
    export_logs: '导出日志', no_logs: '暂无日志；任务启动后会实时追加。', logs_copied: '日志已复制',
    global_search_hint: '按 Ctrl+K 全局搜索',
  };

  let category: Category = $state('programming');
  let view = $state<View>('discover');
  let query = $state('');
  let activeTags = $state<string[]>([]);
  let tools: Tool[] = $state([]);
  let allTags: string[] = $state([]);
  let categoryNames: Record<string, string> = $state({});
  let ui: Record<string, string> = $state({});
  let lang = $state('zh');
  let loading = $state(false);
  let settingsVisible = $state(false);
  let statusMessage = $state('');
  let progressMap = $state<Map<string, Progress>>(new Map());
  let tasks: Progress[] = $state([]);
  let taskPanelOpen = $state(false);
  let selectedTaskId = $state('');
  let taskLogs: LogEntry[] = $state([]);
  let taskRefreshTimer: ReturnType<typeof setInterval>;
  let managers: PackageManagerInfo[] = $state([]);
  let selectedManager = $state('all');
  let packageQuery = $state('');
  let packages: PackageInfo[] = $state([]);
  let packageCache = $state<Map<string, PackageInfo[]>>(new Map());
  let selectedPackages = $state<Set<string>>(new Set());
  let packageLoading = $state(false);
  let installOptions = $state('');
  let bundleText = $state('');
  let importedItems: PackageAction[] = $state([]);
  let packageSearchTimeout: ReturnType<typeof setTimeout>;
  let driverQuery = $state('');
  let selectedDriverCategory = $state('all');
  let driverCategories: string[] = $state([]);
  let drivers: Driver[] = $state([]);
  let driverCache = $state<Map<string, Driver[]>>(new Map());
  let driverLoading = $state(false);
  let driverSearchTimeout: ReturnType<typeof setTimeout>;
  let packageSearchInput = $state<HTMLInputElement>();
  let driverSearchInput = $state<HTMLInputElement>();

  let filteredTools = $derived.by(() => {
    if (query.length > 0) return tools;
    if (activeTags.length > 0) {
      return tools.filter(t => activeTags.every(tag => t.tags.includes(tag)));
    }
    return tools;
  });

  let activeTaskCount = $derived(tasks.filter(task => task.status === 'installing' || task.status === 'downloading').length);
  let viewTitle = $derived(view === 'toolkit' ? tr('title_toolkit') : view === 'drivers' ? tr('title_drivers') : tr('title_packages'));
  let viewDesc = $derived(view === 'toolkit' ? tr('desc_toolkit') : view === 'drivers' ? tr('desc_drivers') : tr('desc_packages'));

  function normalizeTools(rows: unknown): Tool[] {
    return ((rows as Tool[] | null) ?? []).map(tool => ({
      ...tool,
      tags: tool.tags ?? [],
    }));
  }

  async function loadStrings() {
    try {
      categoryNames = (await AppService.GetCategoryNames()) as Record<string, string>;
      ui = (await AppService.GetUIStrings()) as Record<string, string>;
    } catch {
      categoryNames = {
        programming: '程序开发', art: '美术设计', planning: '策划文档',
        audio: '音频制作', qa: '测试 QA', pm: '项目管理', ai: 'AI 环境',
      };
      ui = FALLBACK_UI;
    }
  }

  function tr(key: string, fallback = key) {
    return ui[key] || FALLBACK_UI[key] || fallback;
  }

  async function loadTools() {
    loading = true;
    try {
      tools = query.length > 0
        ? normalizeTools(await AppService.SearchTools(query))
        : normalizeTools(await AppService.GetToolsByCategory(category));
      allTags = (await AppService.GetAllTags()) ?? [];
    } catch { tools = []; }
    finally { loading = false; }
  }

  async function loadManagers() {
    try { managers = (await AppService.GetPackageManagers()) ?? []; }
    catch { managers = []; }
  }

  async function loadPackages(nextView: View = view, force = false) {
    if (nextView === 'toolkit' || nextView === 'bundle') return;
    const queryKey = packageQuery.trim();
    const cacheKey = `${nextView}:${selectedManager}:${queryKey}`;
    if (packageCache.has(cacheKey)) {
      packages = packageCache.get(cacheKey) ?? [];
      packageLoading = false;
    } else {
      packageLoading = true;
    }
    selectedPackages = new Set();
    try {
      let rows: PackageInfo[] = [];
      if (nextView === 'discover') {
        rows = queryKey
          ? ((await AppService.SearchPackages(selectedManager, queryKey, 80)) ?? [])
          : [];
      } else if (nextView === 'installed') {
        rows = force
          ? ((await AppService.RefreshInstalledPackages(selectedManager)) ?? [])
          : ((await AppService.GetInstalledPackages(selectedManager)) ?? []);
      } else if (nextView === 'updates') {
        rows = force
          ? ((await AppService.RefreshPackageUpdates(selectedManager)) ?? [])
          : ((await AppService.GetPackageUpdates(selectedManager)) ?? []);
      }
      packages = rows;
      const nextCache = new Map(packageCache);
      nextCache.set(cacheKey, rows);
      packageCache = nextCache;
    } catch (e) {
      if (!packageCache.has(cacheKey)) packages = [];
      statusMessage = '';
    } finally {
      packageLoading = false;
    }
  }

  function normalizeDrivers(rows: unknown): Driver[] {
    return ((rows as Driver[] | null) ?? []).map(driver => ({
      ...driver,
      aliases: driver.aliases ?? [],
      deviceTypes: driver.deviceTypes ?? [],
    }));
  }

  async function loadDrivers() {
    const cacheKey = `${selectedDriverCategory}:${driverQuery.trim()}`;
    if (driverCache.has(cacheKey)) {
      drivers = driverCache.get(cacheKey) ?? [];
      driverLoading = false;
    } else {
      driverLoading = true;
    }
    try {
      const rows = normalizeDrivers(await AppService.SearchDrivers(driverQuery.trim(), selectedDriverCategory));
      drivers = rows;
      const nextCache = new Map(driverCache);
      nextCache.set(cacheKey, rows);
      driverCache = nextCache;
      driverCategories = (await AppService.GetDriverCategories()) ?? [];
    } catch (e) {
      if (!driverCache.has(cacheKey)) drivers = [];
      statusMessage = '';
    } finally {
      driverLoading = false;
    }
  }

  function onDriverSearchInput() {
    clearTimeout(driverSearchTimeout);
    driverSearchTimeout = setTimeout(loadDrivers, 250);
  }

  function onPackageSearchInput() {
    clearTimeout(packageSearchTimeout);
    if (packageQuery.trim().length < 2) {
      packages = [];
      return;
    }
    packageSearchTimeout = setTimeout(() => loadPackages('discover'), 250);
  }

  async function switchView(nextView: View) {
    view = nextView;
    if (nextView === 'toolkit') {
      await loadTools();
    } else if (nextView === 'drivers') {
      await loadDrivers();
    } else {
      await loadPackages(nextView);
    }
  }

  async function switchManager(id: string) {
    selectedManager = id;
    await loadPackages();
  }

  async function switchDriverCategory(category: string) {
    selectedDriverCategory = category;
    await loadDrivers();
  }

  function packageKey(pkg: PackageInfo) {
    return `${pkg.manager}:${pkg.id}`;
  }

  function progressKey(manager: string, id: string) {
    return `${manager}:${id}`;
  }

  function isActiveTask(task: Progress) {
    return task.status === 'installing' || task.status === 'downloading';
  }

  function taskTitle(id: string) {
    return id.includes(':') ? id : id.replaceAll('-', ' ');
  }

  function logsToText(rows: LogEntry[]) {
    return rows.map(row => `${row.time} [${row.level}] ${row.message}`).join('\n');
  }

  function upsertTask(task: Progress) {
    const next = tasks.filter(row => row.toolId !== task.toolId);
    tasks = [task, ...next].sort((a, b) => (b.updatedAt || '').localeCompare(a.updatedAt || ''));
    progressMap = new Map(progressMap);
    progressMap.set(task.toolId, task);
  }

  function setOptimisticTask(toolId: string, status: string, message: string) {
    const now = new Date().toISOString();
    upsertTask({ toolId, status, percent: 1, message, startedAt: now, updatedAt: now, logLines: 0, exitCode: 0 });
    selectedTaskId = toolId;
    taskPanelOpen = true;
  }

  async function refreshTasks(refreshLogs = false) {
    try {
      const rows = ((await AppService.GetAllInstallProgress()) ?? []) as Progress[];
      tasks = rows
        .filter(row => row.status && row.status !== 'idle')
        .sort((a, b) => (b.updatedAt || '').localeCompare(a.updatedAt || ''));
      const next = new Map(progressMap);
      for (const row of tasks) next.set(row.toolId, row);
      progressMap = next;
      if (!selectedTaskId && tasks.length > 0) selectedTaskId = tasks[0].toolId;
      if (refreshLogs && selectedTaskId) await loadTaskLogs(selectedTaskId, false);
    } catch {}
  }

  async function loadTaskLogs(toolId: string, open = true) {
    selectedTaskId = toolId;
    if (open) taskPanelOpen = true;
    try {
      taskLogs = ((await AppService.GetInstallLogs(toolId)) ?? []) as LogEntry[];
    } catch {
      taskLogs = [];
    }
  }

  async function copySelectedLog() {
    if (!selectedTaskId) return;
    let logText = logsToText(taskLogs);
    try {
      logText = await AppService.ExportInstallLog(selectedTaskId);
    } catch {}
    await navigator.clipboard.writeText(logText);
    statusMessage = tr('logs_copied', '日志已复制');
  }

  async function downloadSelectedLog() {
    if (!selectedTaskId) return;
    let logText = logsToText(taskLogs);
    try {
      logText = await AppService.ExportInstallLog(selectedTaskId);
    } catch {}
    const blob = new Blob([logText], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${selectedTaskId.replaceAll(':', '_')}.log`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function togglePackage(pkg: PackageInfo) {
    const next = new Set(selectedPackages);
    const key = packageKey(pkg);
    if (next.has(key)) next.delete(key);
    else next.add(key);
    selectedPackages = next;
  }

  function selectedPackageActions(): PackageAction[] {
    return packages
      .filter(pkg => selectedPackages.has(packageKey(pkg)))
      .map(pkg => ({ manager: pkg.manager, id: pkg.id, name: pkg.name, options: installOptions }));
  }

  async function runPackageAction(action: 'install' | 'update' | 'uninstall', pkg: PackageInfo) {
    try {
      const key = progressKey(pkg.manager, pkg.id);
      setOptimisticTask(key, 'installing', `${pkg.manager} ${action} ${pkg.id}`);
      statusMessage = await AppService.RunPackageAction(action, pkg.manager, pkg.id, installOptions);
      pollProgress(key);
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  async function runBulk(action: 'install' | 'update' | 'uninstall') {
    const items = selectedPackageActions();
    if (items.length === 0) return;
    try {
      for (const item of items) setOptimisticTask(progressKey(item.manager, item.id), 'installing', `${item.manager} ${action} ${item.id}`);
      statusMessage = await AppService.RunBulkPackageAction(action, items);
      for (const item of items) pollProgress(progressKey(item.manager, item.id));
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  async function exportInstalled() {
    try {
      bundleText = await AppService.ExportInstalledPackages(selectedManager);
      const blob = new Blob([bundleText], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `windownloader233-packages-${selectedManager}.json`;
      a.click();
      URL.revokeObjectURL(url);
      statusMessage = lang === 'zh' ? '备份已生成' : 'Backup generated';
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  async function importBundle() {
    try {
      importedItems = (await AppService.ImportPackageBundle(bundleText)) ?? [];
      statusMessage = `${importedItems.length} packages ready`;
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  async function installImported() {
    if (importedItems.length === 0) return;
    try {
      for (const item of importedItems) setOptimisticTask(progressKey(item.manager, item.id), 'installing', `${item.manager} install ${item.id}`);
      statusMessage = await AppService.RunBulkPackageAction('install', importedItems);
      for (const item of importedItems) pollProgress(progressKey(item.manager, item.id));
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  function openExternal(url: string) {
    if (url) window.open(url, '_blank');
  }

  async function selectCategory(cat: Category) {
    view = 'toolkit';
    category = cat; query = ''; activeTags = [];
    await loadTools();
  }

  async function onSearch(q: string) {
    query = q;
    if (q.length > 0) {
      loading = true;
      try { tools = normalizeTools(await AppService.SearchTools(q)); }
      catch {} finally { loading = false; }
    } else { await loadTools(); }
  }

  async function toggleTag(tag: string) {
    const idx = activeTags.indexOf(tag);
    activeTags = idx >= 0 ? activeTags.filter(t => t !== tag) : [...activeTags, tag];
  }

  async function installTool(tool: Tool) {
    try {
      setOptimisticTask(tool.id, tool.installType === 'download' ? 'downloading' : 'installing', `${tool.installType === 'download' ? 'download' : 'install'} ${getToolName(tool)}`);
      statusMessage = await AppService.InstallTool(tool.id);
      pollProgress(tool.id);
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  function pollProgress(toolId: string) {
    const interval = setInterval(async () => {
      try {
        const p = await AppService.GetInstallProgress(toolId);
        upsertTask(p as Progress);
        if (taskPanelOpen && selectedTaskId === toolId) await loadTaskLogs(toolId, false);
        if (p.status === 'done' || p.status === 'error') {
          clearInterval(interval);
          statusMessage = p.message;
          await refreshTasks(taskPanelOpen);
          if (view === 'toolkit') await loadTools();
          else if (view === 'drivers') await loadDrivers();
          else await loadPackages();
        }
      } catch { clearInterval(interval); }
    }, 500);
  }

  async function uninstallTool(tool: Tool) {
    try {
      setOptimisticTask(tool.id, 'installing', `uninstall ${getToolName(tool)}`);
      statusMessage = await AppService.UninstallTool(tool.id);
      pollProgress(tool.id);
    } catch {}
  }

  async function refreshRegistry() {
    loading = true;
    try {
      statusMessage = await AppService.RefreshRegistry();
      await loadTools();
    } catch {} finally { loading = false; }
  }

  async function forceRefreshCurrentPackages() {
    const target: View = view === 'updates' ? 'updates' : 'installed';
    await loadPackages(target, true);
  }

  function focusGlobalSearch() {
    view = 'discover';
    setTimeout(() => packageSearchInput?.focus(), 0);
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'k') {
      e.preventDefault();
      focusGlobalSearch();
    }
  }

  async function switchLanguage(newLang: string) {
    lang = newLang;
    await AppService.SetLanguage(newLang);
    await loadStrings();
    await loadTools();
  }

  function getToolName(tool: Tool) { return lang === 'zh' ? tool.nameZh : tool.nameEn; }
  function getToolDesc(tool: Tool) { return lang === 'zh' ? tool.descZh : tool.descEn; }

  onMount(async () => {
    try { statusMessage = await AppService.InitRegistry(); } catch {}
    await loadStrings();
    await loadManagers();
    await loadDrivers();
    await refreshTasks(false);
    window.addEventListener('keydown', handleKeydown);
    taskRefreshTimer = setInterval(() => refreshTasks(taskPanelOpen), 1200);
    setTimeout(() => { statusMessage = ''; }, 5000);
  });

  onDestroy(() => {
    clearInterval(taskRefreshTimer);
    clearTimeout(packageSearchTimeout);
    clearTimeout(driverSearchTimeout);
    window.removeEventListener('keydown', handleKeydown);
  });
</script>

<div class="bg-mesh">
</div>

<div class="app">
  <div class="main-area">
    <div class="workspace-head">
      <div class="brand-titlebar">
        <div class="brand-mark" aria-hidden="true">
          <span>WD</span>
        </div>
        <div>
          <div class="brand-name">WinDownloader233</div>
          <h1>{viewTitle}</h1>
          <p>{viewDesc} · {tr('global_search_hint', '按 Ctrl+K 全局搜索')}</p>
        </div>
      </div>
      <div class="tabs" aria-label="Views">
        <button class:active={view === 'discover'} onclick={() => switchView('discover')}>{tr('view_discover')}</button>
        <button class:active={view === 'installed'} onclick={() => switchView('installed')}>{tr('view_installed')}</button>
        <button class:active={view === 'updates'} onclick={() => switchView('updates')}>{tr('view_updates')}</button>
        <button class:active={view === 'drivers'} onclick={() => switchView('drivers')}>{tr('view_drivers')}</button>
        <button class:active={view === 'bundle'} onclick={() => switchView('bundle')}>{tr('view_bundle')}</button>
        <button class:active={view === 'toolkit'} onclick={() => switchView('toolkit')}>{tr('view_toolkit')}</button>
        <button class:active={taskPanelOpen} class="task-toggle" onclick={() => { taskPanelOpen = !taskPanelOpen; refreshTasks(true); }}>
          {tr('tasks')}{activeTaskCount > 0 ? ` ${activeTaskCount}` : ''}
        </button>
        <button onclick={() => settingsVisible = true}>{tr('settings')}</button>
      </div>
    </div>

    {#if view === 'toolkit'}
      <div class="manager-strip category-strip">
        {#each ALL_CATEGORIES as cat}
          <button
            class:active={category === cat}
            onclick={() => selectCategory(cat)}
          >
            <span>{CATEGORY_ICONS[cat]}</span>
            {categoryNames[cat] || cat}
          </button>
        {/each}
      </div>
    {:else if view !== 'drivers'}
      <div class="manager-strip">
        <button class:active={selectedManager === 'all'} onclick={() => switchManager('all')}>{tr('all_sources')}</button>
        {#each managers as manager}
          <button
            class:active={selectedManager === manager.id}
            class:missing={!manager.available}
            title={manager.available ? `${manager.name} ${manager.version}` : manager.installHint}
            onclick={() => switchManager(manager.id)}
          >
            <span class="status-led" class:on={manager.available}></span>
            {manager.name}
          </button>
        {/each}
      </div>
    {/if}

    {#if taskPanelOpen}
      <aside class="task-panel" aria-label="Task logs">
        <div class="task-panel-head">
          <div>
            <strong>{tr('background_tasks')}</strong>
            <span>{tasks.length} {tr('task_count')} · {activeTaskCount} {tr('running_count')}</span>
          </div>
          <button class="icon-btn" onclick={() => taskPanelOpen = false}>×</button>
        </div>

        {#if tasks.length === 0}
          <div class="task-empty">{tr('no_tasks')}</div>
        {:else}
          <div class="task-body">
            <div class="task-list">
              {#each tasks as task (task.toolId)}
                <button
                  class="task-item"
                  class:active={selectedTaskId === task.toolId}
                  class:error={task.status === 'error'}
                  onclick={() => loadTaskLogs(task.toolId)}
                >
                  <span class="task-name">{taskTitle(task.toolId)}</span>
                  <span class="task-meta">
                    <i class:running={isActiveTask(task)}></i>
                    {task.status}{task.logLines ? ` · ${task.logLines} lines` : ''}
                  </span>
                  <span class="task-message">{task.message}</span>
                  {#if task.percent >= 0 && isActiveTask(task)}
                    <span class="mini-progress"><b style="width: {Math.min(100, Math.max(0, task.percent))}%"></b></span>
                  {/if}
                </button>
              {/each}
            </div>

            <div class="task-log-pane">
              <div class="task-log-head">
                <strong>{selectedTaskId ? taskTitle(selectedTaskId) : tr('logs')}</strong>
                <div>
                  <button onclick={() => selectedTaskId && loadTaskLogs(selectedTaskId, false)}>{tr('refresh_logs')}</button>
                  <button onclick={copySelectedLog} disabled={!selectedTaskId}>{tr('copy_logs')}</button>
                  <button onclick={downloadSelectedLog} disabled={!selectedTaskId}>{tr('export_logs')}</button>
                </div>
              </div>
              {#if taskLogs.length === 0}
                <div class="task-empty compact">{tr('no_logs')}</div>
              {:else}
                <pre class="task-log">{logsToText(taskLogs)}</pre>
              {/if}
            </div>
          </div>
        {/if}
      </aside>
    {/if}

    {#if view === 'toolkit'}
      <Header
        {query}
        {allTags}
        {activeTags}
        {ui}
        {loading}
        {statusMessage}
        onSearch={onSearch}
        onToggleTag={toggleTag}
        onRefresh={refreshRegistry}
      />

      <ToolList
        tools={filteredTools}
        {progressMap}
        {lang}
        {loading}
        {ui}
        getToolName={getToolName}
        getToolDesc={getToolDesc}
        onInstall={installTool}
        onUninstall={uninstallTool}
      />
    {:else if view === 'drivers'}
      <section class="package-shell">
        {#if statusMessage}
          <div class="status-line">{statusMessage}</div>
        {/if}

        <div class="driver-toolbar">
          <div class="package-search">
            <input
              bind:this={driverSearchInput}
              bind:value={driverQuery}
              placeholder={tr('driver_placeholder')}
              oninput={onDriverSearchInput}
              onkeydown={(e) => { if (e.key === 'Enter') loadDrivers(); }}
            />
            <button onclick={loadDrivers} disabled={driverLoading}>{tr('search_btn')}</button>
          </div>
          <div class="driver-cats">
            <button class:active={selectedDriverCategory === 'all'} onclick={() => switchDriverCategory('all')}>{tr('all')}</button>
            {#each driverCategories as cat}
              <button class:active={selectedDriverCategory === cat} onclick={() => switchDriverCategory(cat)}>
                {cat}
              </button>
            {/each}
          </div>
        </div>

        {#if driverLoading}
          <div class="package-empty">{tr('loading')}</div>
        {:else if drivers.length === 0}
          <div class="package-empty">{tr('driver_empty')}</div>
        {:else}
          <div class="driver-grid">
            {#each drivers as driver (driver.id)}
              <article class="driver-card">
                <div class="driver-card-head">
                  <div>
                    <span class="vendor">{driver.vendor}</span>
                    <h3>{driver.name}</h3>
                  </div>
                  <span class="driver-category">{driver.category}</span>
                </div>
                <p>{driver.notes}</p>
                <div class="driver-types">
                  {#each driver.deviceTypes ?? [] as type}
                    <span>{type}</span>
                  {/each}
                </div>
                <div class="driver-aliases">
                  {#each (driver.aliases ?? []).slice(0, 9) as alias}
                    <span>{alias}</span>
                  {/each}
                </div>
                <div class="driver-actions">
                  <button class="primary-link" onclick={() => openExternal(driver.downloadUrl)}>{tr('official_download')}</button>
                  <button onclick={() => openExternal(driver.supportUrl)}>{tr('support_page')}</button>
                </div>
              </article>
            {/each}
          </div>
        {/if}
      </section>
    {:else}
      <section class="package-shell">
        {#if statusMessage}
          <div class="status-line">{statusMessage}</div>
        {/if}

        {#if view === 'discover'}
          <div class="package-toolbar">
            <div class="package-search">
              <input
                bind:this={packageSearchInput}
                bind:value={packageQuery}
                placeholder={tr('package_placeholder')}
                oninput={onPackageSearchInput}
                onkeydown={(e) => { if (e.key === 'Enter') loadPackages(); }}
              />
              <button onclick={() => loadPackages()} disabled={packageLoading}>{tr('search_btn')}</button>
            </div>
            <input class="options-input" bind:value={installOptions} placeholder={tr('package_options')} />
            <button class="bulk-btn" onclick={() => runBulk('install')} disabled={selectedPackages.size === 0}>{tr('install_selected')}</button>
          </div>
        {:else if view === 'installed' || view === 'updates'}
          <div class="package-toolbar">
            <button onclick={forceRefreshCurrentPackages} disabled={packageLoading}>{view === 'updates' ? tr('check_updates') : tr('force_refresh')}</button>
            {#if view === 'updates'}
              <button class="bulk-btn" onclick={() => runBulk('update')} disabled={selectedPackages.size === 0}>{tr('update_selected')}</button>
            {:else}
              <button class="danger-btn" onclick={() => runBulk('uninstall')} disabled={selectedPackages.size === 0}>{tr('uninstall_selected')}</button>
            {/if}
            <button onclick={exportInstalled}>{tr('export_backup')}</button>
          </div>
        {:else if view === 'bundle'}
          <div class="bundle-panel">
            <div class="package-toolbar">
              <button onclick={exportInstalled}>{tr('export_installed')}</button>
              <button onclick={importBundle} disabled={!bundleText.trim()}>{tr('parse_import')}</button>
              <button class="bulk-btn" onclick={installImported} disabled={importedItems.length === 0}>{tr('install_imported')}</button>
            </div>
            <textarea bind:value={bundleText} placeholder={tr('bundle_placeholder')}></textarea>
            {#if importedItems.length > 0}
              <div class="import-list">
                {#each importedItems as item}
                  <span>{item.manager}:{item.id}</span>
                {/each}
              </div>
            {/if}
          </div>
        {/if}

        {#if view !== 'bundle'}
          {#if packageLoading}
            <div class="package-empty">{tr('loading')}</div>
          {:else if packages.length === 0}
            <div class="package-empty">
              {view === 'discover'
                ? tr('package_empty')
                : tr('no_package_results')}
            </div>
          {:else}
            <div class="package-table">
              <div class="package-row package-head">
                <span></span>
                <span>{tr('manager')}</span>
                <span>{tr('name')}</span>
                <span>{tr('version')}</span>
                <span>{tr('available')}</span>
                <span>{tr('action')}</span>
              </div>
              {#each packages as pkg (packageKey(pkg))}
                {@const p = progressMap.get(progressKey(pkg.manager, pkg.id))}
                <div class="package-row">
                  <label class="check-cell">
                    <input type="checkbox" checked={selectedPackages.has(packageKey(pkg))} onchange={() => togglePackage(pkg)} />
                  </label>
                  <span class="manager-chip">{pkg.manager}</span>
                  <span class="pkg-name">
                    <strong>{pkg.name || pkg.id}</strong>
                    <small>{pkg.id}</small>
                    {#if pkg.description}<em>{pkg.description}</em>{/if}
                    {#if p?.message}<em class:error-text={p.status === 'error'}>{p.message}</em>{/if}
                  </span>
                  <span>{pkg.version || '-'}</span>
                  <span>{pkg.available || '-'}</span>
                  <span class="pkg-actions">
                    {#if p?.status === 'installing'}
                      <button disabled>{tr('running')}</button>
                    {:else if view === 'discover'}
                      <button onclick={() => runPackageAction('install', pkg)}>{tr('install')}</button>
                    {:else if view === 'updates'}
                      <button onclick={() => runPackageAction('update', pkg)}>{tr('update')}</button>
                    {:else}
                      <button class="danger-btn" onclick={() => runPackageAction('uninstall', pkg)}>{tr('uninstall')}</button>
                    {/if}
                  </span>
                </div>
              {/each}
            </div>
          {/if}
        {/if}
      </section>
    {/if}
  </div>

  {#if settingsVisible}
    <SettingsPanel
      {lang}
      {ui}
      onLanguageChange={switchLanguage}
      onClose={() => settingsVisible = false}
    />
  {/if}
</div>

<style>
  .bg-mesh {
    position: fixed;
    inset: 0;
    z-index: 0;
    overflow: hidden;
    background: var(--bg-base);
  }

  .app {
    position: relative;
    z-index: 1;
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .main-area {
    position: relative;
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }

  .workspace-head {
    padding: 18px 28px 12px;
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 18px;
    flex-shrink: 0;
    border-bottom: 1px solid var(--glass-border);
    background: #fff;
  }

  .brand-titlebar {
    display: flex;
    align-items: flex-start;
    gap: 14px;
    min-width: 280px;
  }

  .brand-mark {
    width: 42px;
    height: 42px;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 13px;
    font-weight: 900;
    letter-spacing: 0;
    background:
      linear-gradient(135deg, rgba(255,255,255,0.22), transparent 42%),
      linear-gradient(135deg, #0A84FF 0%, #16a34a 52%, #0f172a 100%);
    box-shadow: 0 10px 24px rgba(10, 132, 255, 0.24);
    flex-shrink: 0;
  }

  .brand-name {
    color: var(--accent-blue);
    font-size: 12px;
    font-weight: 800;
    line-height: 1;
    margin-bottom: 5px;
  }

  .workspace-head h1 {
    margin: 0;
    font-size: 24px;
    line-height: 1.2;
    color: var(--text-primary);
  }

  .workspace-head p {
    margin: 5px 0 0;
    color: var(--text-secondary);
    font-size: 13px;
  }

  .tabs,
  .manager-strip,
  .package-toolbar {
    display: flex;
    gap: 8px;
    align-items: center;
    flex-wrap: wrap;
  }

  .driver-toolbar {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 16px;
    padding: 14px;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #fff;
    box-shadow: var(--shadow-sm);
  }

  .driver-cats {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .driver-cats button {
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #f8fafc;
    color: var(--text-secondary);
    padding: 7px 11px;
    font-size: 13px;
    font-weight: 700;
  }

  .driver-cats button.active,
  .driver-cats button:hover {
    border-color: #0A84FF;
    background: #e8f3ff;
    color: var(--text-primary);
  }

  .tabs button,
  .manager-strip button,
  .package-toolbar button,
  .pkg-actions button,
  .package-search button {
    border: 1px solid var(--glass-border);
    background: #fff;
    color: var(--text-secondary);
    border-radius: 8px;
    padding: 8px 12px;
    font-size: 13px;
    font-weight: 600;
  }

  .tabs button.active,
  .manager-strip button.active,
  .package-toolbar button:hover:not(:disabled),
  .pkg-actions button:hover:not(:disabled) {
    color: var(--text-primary);
    border-color: #0A84FF;
    background: #e8f3ff;
  }

  .tabs .task-toggle {
    border-color: #94a3b8;
  }

  .tabs .task-toggle.active {
    color: #fff;
    background: #0f172a;
    border-color: #0f172a;
  }

  .task-panel {
    position: absolute;
    right: 18px;
    top: 98px;
    bottom: 18px;
    width: min(760px, calc(100% - 36px));
    z-index: 20;
    display: flex;
    flex-direction: column;
    border: 1px solid #cbd5e1;
    border-radius: 8px;
    background: #ffffff;
    box-shadow: 0 22px 70px rgba(15, 23, 42, 0.24);
    overflow: hidden;
  }

  .task-panel-head,
  .task-log-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
    padding: 12px 14px;
    border-bottom: 1px solid var(--glass-border);
    background: #f8fafc;
  }

  .task-panel-head strong,
  .task-log-head strong {
    color: var(--text-primary);
    font-size: 14px;
  }

  .task-panel-head span {
    display: block;
    margin-top: 2px;
    color: var(--text-tertiary);
    font-size: 12px;
  }

  .icon-btn {
    width: 30px;
    height: 30px;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #fff;
    color: var(--text-secondary);
    font-size: 18px;
    line-height: 1;
  }

  .task-body {
    min-height: 0;
    flex: 1;
    display: grid;
    grid-template-columns: 285px minmax(0, 1fr);
  }

  .task-list {
    min-height: 0;
    overflow-y: auto;
    border-right: 1px solid var(--glass-border);
    background: #f8fafc;
    padding: 8px;
  }

  .task-item {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 5px;
    padding: 10px;
    border: 1px solid transparent;
    border-radius: 8px;
    background: transparent;
    text-align: left;
    color: var(--text-secondary);
  }

  .task-item:hover,
  .task-item.active {
    border-color: #bfdbfe;
    background: #eff6ff;
  }

  .task-item.error {
    border-color: #fecdd3;
    background: #fff1f2;
  }

  .task-name {
    color: var(--text-primary);
    font-weight: 800;
    font-size: 13px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .task-meta {
    display: flex;
    align-items: center;
    gap: 6px;
    color: var(--text-tertiary);
    font-size: 11px;
  }

  .task-meta i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #94a3b8;
  }

  .task-meta i.running {
    background: #0A84FF;
    box-shadow: 0 0 0 4px rgba(10, 132, 255, 0.14);
  }

  .task-message {
    min-height: 16px;
    color: var(--text-secondary);
    font-size: 12px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .mini-progress {
    height: 4px;
    overflow: hidden;
    border-radius: 999px;
    background: #dbeafe;
  }

  .mini-progress b {
    display: block;
    height: 100%;
    border-radius: inherit;
    background: #0A84FF;
  }

  .task-log-pane {
    min-width: 0;
    min-height: 0;
    display: flex;
    flex-direction: column;
  }

  .task-log-head div {
    display: flex;
    gap: 6px;
  }

  .task-log-head button {
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #fff;
    color: var(--text-secondary);
    padding: 6px 9px;
    font-size: 12px;
    font-weight: 700;
  }

  .task-log {
    flex: 1;
    min-height: 0;
    margin: 0;
    padding: 12px;
    overflow: auto;
    background: #0b1020;
    color: #cbd5e1;
    font: 12px/1.55 'Cascadia Code', Consolas, monospace;
    white-space: pre-wrap;
    user-select: text;
    -webkit-user-select: text;
  }

  .task-empty {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 180px;
    color: var(--text-tertiary);
    font-size: 13px;
  }

  .task-empty.compact {
    min-height: 0;
    height: 100%;
  }

  .manager-strip button {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 7px 10px;
  }

  .category-strip button span {
    font-size: 15px;
    line-height: 1;
  }

  .manager-strip {
    padding: 12px 28px;
    border-bottom: 1px solid var(--glass-border);
    background: #f8fafc;
    flex-shrink: 0;
  }

  .manager-strip button.missing {
    opacity: 0.7;
  }

  .status-led {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: var(--accent-pink);
  }

  .status-led.on {
    background: var(--accent-green);
  }

  .package-shell {
    flex: 1;
    overflow-y: auto;
    padding: 18px 28px 28px;
  }

  .status-line {
    margin-bottom: 10px;
    color: var(--accent-teal);
    font-size: 12px;
  }

  .package-toolbar {
    margin-bottom: 16px;
    padding: 14px;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #fff;
    box-shadow: var(--shadow-sm);
  }

  .package-search {
    display: flex;
    flex: 1;
    min-width: 360px;
    gap: 8px;
  }

  .package-search input,
  .options-input,
  .bundle-panel textarea {
    border: 1px solid var(--glass-border);
    background: #fff;
    color: var(--text-primary);
    border-radius: 8px;
    padding: 11px 12px;
    outline: none;
    font-size: 14px;
  }

  .package-search input {
    flex: 1;
    min-height: 44px;
    border-color: #94a3b8;
    box-shadow: inset 0 0 0 1px rgba(148, 163, 184, 0.25);
  }

  .package-search input:focus,
  .options-input:focus,
  .bundle-panel textarea:focus {
    border-color: #0A84FF;
    box-shadow: 0 0 0 3px rgba(10, 132, 255, 0.14);
  }

  .package-search button {
    min-width: 86px;
    min-height: 44px;
    border-color: #0A84FF;
    background: #0A84FF;
    color: #fff;
    font-size: 14px;
  }

  .options-input {
    width: 220px;
  }

  .bulk-btn {
    color: #fff !important;
    background: linear-gradient(135deg, var(--accent-blue), #0070E0) !important;
  }

  .danger-btn {
    color: var(--accent-pink) !important;
  }

  button:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }

  .package-table {
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    overflow: hidden;
    background: #fff;
    box-shadow: var(--shadow-sm);
  }

  .package-row {
    display: grid;
    grid-template-columns: 34px 90px minmax(220px, 1fr) 130px 130px 140px;
    gap: 10px;
    align-items: center;
    padding: 10px 12px;
    border-bottom: 1px solid var(--glass-border);
    color: var(--text-secondary);
    font-size: 13px;
    min-height: 54px;
  }

  .package-row:last-child {
    border-bottom: none;
  }

  .package-head {
    color: var(--text-secondary);
    font-weight: 700;
    text-transform: uppercase;
    background: #f8fafc;
  }

  .check-cell input {
    width: 16px;
    height: 16px;
  }

  .manager-chip {
    color: var(--accent-teal);
    font-weight: 700;
    text-transform: uppercase;
  }

  .pkg-name {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .pkg-name strong {
    color: var(--text-primary);
    font-size: 14px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pkg-name small,
  .pkg-name em {
    color: var(--text-tertiary);
    font-style: normal;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pkg-name em.error-text {
    color: var(--accent-pink);
  }

  .pkg-actions {
    display: flex;
    gap: 6px;
  }

  .package-empty {
    height: 48vh;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-tertiary);
    background: #fff;
    border: 1px dashed var(--glass-border);
    border-radius: 8px;
  }

  .driver-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 12px;
  }

  .driver-card {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-height: 238px;
    padding: 16px;
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #fff;
    box-shadow: var(--shadow-sm);
  }

  .driver-card-head {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    align-items: flex-start;
  }

  .vendor {
    display: inline-block;
    margin-bottom: 4px;
    color: #0A84FF;
    font-size: 12px;
    font-weight: 800;
    text-transform: uppercase;
  }

  .driver-card h3 {
    margin: 0;
    color: var(--text-primary);
    font-size: 17px;
    line-height: 1.3;
  }

  .driver-card p {
    color: var(--text-secondary);
    font-size: 13px;
    line-height: 1.5;
  }

  .driver-category {
    flex-shrink: 0;
    padding: 4px 8px;
    border-radius: 999px;
    background: #f1f5f9;
    color: var(--text-secondary);
    font-size: 12px;
    font-weight: 700;
  }

  .driver-types,
  .driver-aliases {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .driver-types span,
  .driver-aliases span {
    padding: 4px 7px;
    border-radius: 6px;
    font-size: 12px;
  }

  .driver-types span {
    background: #ecfdf5;
    color: #047857;
    font-weight: 700;
  }

  .driver-aliases span {
    background: #f8fafc;
    color: var(--text-tertiary);
  }

  .driver-actions {
    display: flex;
    gap: 8px;
    margin-top: auto;
  }

  .driver-actions button {
    border: 1px solid var(--glass-border);
    border-radius: 8px;
    background: #fff;
    color: var(--text-secondary);
    padding: 9px 12px;
    font-weight: 700;
  }

  .driver-actions .primary-link {
    border-color: #0A84FF;
    background: #0A84FF;
    color: #fff;
  }

  .bundle-panel {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .bundle-panel textarea {
    min-height: 300px;
    resize: vertical;
    font-family: 'Cascadia Code', Consolas, monospace;
    font-size: 12px;
  }

  .import-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .import-list span {
    padding: 5px 8px;
    border: 1px solid var(--glass-border);
    border-radius: var(--radius-sm);
    color: var(--text-secondary);
    background: var(--glass-3);
    font-size: 11px;
  }

  @media (max-width: 980px) {
    .workspace-head {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;
      padding: 16px;
    }

    .tabs {
      flex-wrap: nowrap;
      overflow-x: auto;
      padding-bottom: 4px;
    }

    .tabs button {
      flex: 0 0 auto;
      white-space: nowrap;
    }

    .manager-strip,
    .package-shell {
      padding-left: 16px;
      padding-right: 16px;
    }

    .package-search {
      min-width: 0;
      width: 100%;
      flex-direction: column;
    }

    .options-input {
      width: 100%;
    }

    .package-row {
      grid-template-columns: 30px 74px minmax(160px, 1fr) 90px 90px 110px;
    }

    .task-body {
      grid-template-columns: 1fr;
    }

    .task-list {
      max-height: 220px;
      border-right: 0;
      border-bottom: 1px solid var(--glass-border);
    }
  }
</style>
