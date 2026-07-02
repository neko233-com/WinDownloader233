<script lang="ts">
  import { onMount } from 'svelte';
  import type { Tool, Progress, Category } from './lib/types';
  import { ALL_CATEGORIES, CATEGORY_ICONS } from './lib/types';
  import Sidebar from './components/Sidebar.svelte';
  import Header from './components/Header.svelte';
  import ToolList from './components/ToolList.svelte';
  import SettingsPanel from './components/SettingsPanel.svelte';
  import { AppService } from '../bindings/github.com/neko233/WinDownloader233';

  let category: Category = $state('programming');
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

  let filteredTools = $derived.by(() => {
    if (query.length > 0) return tools;
    if (activeTags.length > 0) {
      return tools.filter(t => activeTags.every(tag => t.tags.includes(tag)));
    }
    return tools;
  });

  async function loadStrings() {
    try {
      categoryNames = await AppService.GetCategoryNames();
      ui = await AppService.GetUIStrings();
    } catch {
      categoryNames = {
        programming: '程序开发', art: '美术设计', planning: '策划文档',
        audio: '音频制作', qa: '测试 QA', pm: '项目管理', ai: 'AI 环境',
      };
      ui = { search: '搜索工具...', install: '安装', installed: '已安装', settings: '设置', refresh: '刷新列表' };
    }
  }

  async function loadTools() {
    loading = true;
    try {
      tools = query.length > 0
        ? await AppService.SearchTools(query)
        : await AppService.GetToolsByCategory(category);
      allTags = await AppService.GetAllTags();
    } catch { tools = []; }
    finally { loading = false; }
  }

  async function selectCategory(cat: Category) {
    category = cat; query = ''; activeTags = [];
    await loadTools();
  }

  async function onSearch(q: string) {
    query = q;
    if (q.length > 0) {
      loading = true;
      try { tools = await AppService.SearchTools(q); }
      catch {} finally { loading = false; }
    } else { await loadTools(); }
  }

  async function toggleTag(tag: string) {
    const idx = activeTags.indexOf(tag);
    activeTags = idx >= 0 ? activeTags.filter(t => t !== tag) : [...activeTags, tag];
  }

  async function installTool(tool: Tool) {
    try {
      statusMessage = await AppService.InstallTool(tool.id);
      pollProgress(tool.id);
    } catch (e) { statusMessage = `Error: ${e}`; }
  }

  function pollProgress(toolId: string) {
    const interval = setInterval(async () => {
      try {
        const p = await AppService.GetInstallProgress(toolId);
        progressMap = new Map(progressMap);
        progressMap.set(toolId, p);
        if (p.status === 'done' || p.status === 'error') {
          clearInterval(interval);
          statusMessage = p.message;
          await loadTools();
        }
      } catch { clearInterval(interval); }
    }, 500);
  }

  async function uninstallTool(tool: Tool) {
    try {
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
    await loadTools();
    setTimeout(() => { statusMessage = ''; }, 5000);
  });
</script>

<!-- Animated gradient mesh background -->
<div class="bg-mesh">
  <div class="orb orb-1"></div>
  <div class="orb orb-2"></div>
  <div class="orb orb-3"></div>
  <div class="orb orb-4"></div>
</div>

<div class="app">
  <Sidebar
    {category}
    {categoryNames}
    {CATEGORY_ICONS}
    {ALL_CATEGORIES}
    onSelect={selectCategory}
    onSettings={() => settingsVisible = true}
  />

  <div class="main-area">
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
    background: #000;
  }

  .orb {
    position: absolute;
    border-radius: 50%;
    filter: blur(100px);
    opacity: 0.35;
    animation: float 20s ease-in-out infinite;
  }

  .orb-1 {
    width: 600px; height: 600px;
    background: radial-gradient(circle, #0A84FF 0%, transparent 70%);
    top: -15%; left: -10%;
    animation-delay: 0s;
    animation-duration: 25s;
  }

  .orb-2 {
    width: 500px; height: 500px;
    background: radial-gradient(circle, #BF5AF2 0%, transparent 70%);
    top: 50%; right: -8%;
    animation-delay: -5s;
    animation-duration: 22s;
  }

  .orb-3 {
    width: 450px; height: 450px;
    background: radial-gradient(circle, #30D158 0%, transparent 70%);
    bottom: -10%; left: 30%;
    animation-delay: -10s;
    animation-duration: 28s;
  }

  .orb-4 {
    width: 350px; height: 350px;
    background: radial-gradient(circle, #FF9F0A 0%, transparent 70%);
    top: 20%; left: 50%;
    animation-delay: -15s;
    animation-duration: 30s;
    opacity: 0.2;
  }

  @keyframes float {
    0%, 100% { transform: translate(0, 0) scale(1); }
    25% { transform: translate(40px, -30px) scale(1.05); }
    50% { transform: translate(-20px, 20px) scale(0.95); }
    75% { transform: translate(30px, 40px) scale(1.02); }
  }

  .app {
    position: relative;
    z-index: 1;
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .main-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    min-width: 0;
  }
</style>
