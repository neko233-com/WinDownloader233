<script lang="ts">
  import { onMount } from 'svelte';
  import type { Tool, Progress, Category } from './lib/types';
  import { ALL_CATEGORIES, CATEGORY_ICONS } from './lib/types';
  import Sidebar from './components/Sidebar.svelte';
  import Header from './components/Header.svelte';
  import ToolList from './components/ToolList.svelte';
  import SettingsPanel from './components/SettingsPanel.svelte';
  import { AppService } from '../bindings/github.com/neko233/WinDownloader233';

  // Reactive state
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

  // Load tools when category, query, or tags change
  let filteredTools = $derived.by(() => {
    if (query.length > 0) {
      return tools; // search results already filtered
    }
    if (activeTags.length > 0) {
      return tools.filter(t =>
        activeTags.every(tag => t.tags.includes(tag))
      );
    }
    return tools;
  });

  async function loadStrings() {
    try {
      categoryNames = await AppService.GetCategoryNames();
      ui = await AppService.GetUIStrings();
    } catch (e) {
      console.error('Failed to load strings:', e);
      // Fallback
      categoryNames = {
        programming: '程序开发', art: '美术设计', planning: '策划文档',
        audio: '音频制作', qa: '测试 QA', pm: '项目管理', ai: 'AI 环境',
      };
      ui = {
        search: '搜索工具...', install: '安装', installed: '已安装',
        settings: '设置', refresh: '刷新列表',
      };
    }
  }

  async function loadTools() {
    loading = true;
    try {
      if (query.length > 0) {
        tools = await AppService.SearchTools(query);
      } else {
        tools = await AppService.GetToolsByCategory(category);
      }
      allTags = await AppService.GetAllTags();
    } catch (e) {
      console.error('Failed to load tools:', e);
      tools = [];
    } finally {
      loading = false;
    }
  }

  async function selectCategory(cat: Category) {
    category = cat;
    query = '';
    activeTags = [];
    await loadTools();
  }

  async function onSearch(q: string) {
    query = q;
    if (q.length > 0) {
      loading = true;
      try {
        tools = await AppService.SearchTools(q);
      } catch (e) {
        console.error(e);
      } finally {
        loading = false;
      }
    } else {
      await loadTools();
    }
  }

  async function toggleTag(tag: string) {
    const idx = activeTags.indexOf(tag);
    if (idx >= 0) {
      activeTags = activeTags.filter(t => t !== tag);
    } else {
      activeTags = [...activeTags, tag];
    }
  }

  async function installTool(tool: Tool) {
    try {
      const msg = await AppService.InstallTool(tool.id);
      statusMessage = msg;
      // Poll for progress
      pollProgress(tool.id);
    } catch (e) {
      console.error('Install failed:', e);
      statusMessage = `Error: ${e}`;
    }
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
          // Refresh tools to update installed status
          await loadTools();
        }
      } catch (e) {
        clearInterval(interval);
      }
    }, 500);
  }

  async function uninstallTool(tool: Tool) {
    try {
      const msg = await AppService.UninstallTool(tool.id);
      statusMessage = msg;
      pollProgress(tool.id);
    } catch (e) {
      console.error(e);
    }
  }

  async function refreshRegistry() {
    loading = true;
    try {
      const msg = await AppService.RefreshRegistry();
      statusMessage = msg;
      await loadTools();
    } catch (e) {
      console.error(e);
    } finally {
      loading = false;
    }
  }

  async function switchLanguage(newLang: string) {
    lang = newLang;
    await AppService.SetLanguage(newLang);
    await loadStrings();
    await loadTools();
  }

  function getToolName(tool: Tool): string {
    return lang === 'zh' ? tool.nameZh : tool.nameEn;
  }

  function getToolDesc(tool: Tool): string {
    return lang === 'zh' ? tool.descZh : tool.descEn;
  }

  function getCategoryLabel(cat: string): string {
    return categoryNames[cat] || cat;
  }

  onMount(async () => {
    // Init
    try {
      const initMsg = await AppService.InitRegistry();
      statusMessage = initMsg;
    } catch (e) {
      console.error('Init failed:', e);
    }

    await loadStrings();
    await loadTools();

    // Clear status after 5s
    setTimeout(() => { statusMessage = ''; }, 5000);
  });
</script>

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
  :global(*) {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
  }

  :global(html, body) {
    height: 100%;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Microsoft YaHei', sans-serif;
    background: #0f1117;
    color: #e1e4e8;
  }

  :global(#app) {
    height: 100%;
  }

  .app {
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
