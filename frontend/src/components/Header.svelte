<script lang="ts">
  interface Props {
    query: string;
    allTags: string[];
    activeTags: string[];
    ui: Record<string, string>;
    loading: boolean;
    statusMessage: string;
    onSearch: (q: string) => void;
    onToggleTag: (tag: string) => void;
    onRefresh: () => void;
  }

  let { query, allTags, activeTags, ui, loading, statusMessage, onSearch, onToggleTag, onRefresh }: Props = $props();

  let inputValue = $state('');
  let searchTimeout: ReturnType<typeof setTimeout>;

  function handleInput() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      onSearch(inputValue);
    }, 300);
  }

  function clearSearch() {
    inputValue = '';
    onSearch('');
  }
</script>

<header class="header">
  <div class="header-top">
    <div class="search-box">
      <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/>
        <path d="m21 21-4.35-4.35"/>
      </svg>
      <input
        type="text"
        class="search-input"
        placeholder={ui['search'] || '搜索工具...'}
        bind:value={inputValue}
        oninput={handleInput}
      />
      {#if inputValue}
        <button class="search-clear" onclick={clearSearch}>✕</button>
      {/if}
    </div>

    <div class="header-actions">
      <button class="action-btn" onclick={onRefresh} title={ui['refresh'] || '刷新列表'} disabled={loading}>
        <svg class="action-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M21.5 2v6h-6M2.5 22v-6h6M2 11.5a10 10 0 0 1 18.8-4.3M22 12.5a10 10 0 0 1-18.8 4.3"/>
        </svg>
        {#if loading}
          <span class="spinner"></span>
        {/if}
      </button>
    </div>
  </div>

  {#if statusMessage}
    <div class="status-bar">{statusMessage}</div>
  {/if}

  {#if allTags.length > 0}
    <div class="tag-bar">
      {#each allTags.slice(0, 20) as tag}
        <button
          class="tag-chip"
          class:active={activeTags.includes(tag)}
          onclick={() => onToggleTag(tag)}
        >
          {tag}
        </button>
      {/each}
    </div>
  {/if}
</header>

<style>
  .header {
    padding: 16px 24px;
    border-bottom: 1px solid #2d3148;
    background: #0f1117;
    flex-shrink: 0;
  }

  .header-top {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search-box {
    flex: 1;
    display: flex;
    align-items: center;
    background: #161822;
    border: 1px solid #2d3148;
    border-radius: 8px;
    padding: 0 12px;
    transition: border-color 0.15s;
  }

  .search-box:focus-within {
    border-color: #58a6ff;
  }

  .search-icon {
    width: 18px;
    height: 18px;
    color: #6e7681;
    flex-shrink: 0;
  }

  .search-input {
    flex: 1;
    background: transparent;
    border: none;
    outline: none;
    color: #c9d1d9;
    font-size: 14px;
    padding: 10px 8px;
    font-family: inherit;
  }

  .search-input::placeholder {
    color: #484f58;
  }

  .search-clear {
    background: none;
    border: none;
    color: #6e7681;
    cursor: pointer;
    font-size: 14px;
    padding: 4px;
  }

  .search-clear:hover {
    color: #c9d1d9;
  }

  .header-actions {
    display: flex;
    gap: 8px;
  }

  .action-btn {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 38px;
    height: 38px;
    border: 1px solid #2d3148;
    border-radius: 8px;
    background: #161822;
    color: #8b949e;
    cursor: pointer;
    transition: all 0.15s;
  }

  .action-btn:hover:not(:disabled) {
    background: #1c2030;
    color: #c9d1d9;
  }

  .action-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .action-icon {
    width: 18px;
    height: 18px;
  }

  .spinner {
    position: absolute;
    inset: 4px;
    border: 2px solid transparent;
    border-top-color: #58a6ff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .status-bar {
    margin-top: 8px;
    padding: 6px 12px;
    background: #1a2332;
    border: 1px solid #1f3a5f;
    border-radius: 6px;
    color: #58a6ff;
    font-size: 12px;
  }

  .tag-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 12px;
  }

  .tag-chip {
    padding: 4px 10px;
    border: 1px solid #2d3148;
    border-radius: 12px;
    background: transparent;
    color: #8b949e;
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .tag-chip:hover {
    border-color: #58a6ff;
    color: #c9d1d9;
  }

  .tag-chip.active {
    background: #1f3a5f;
    border-color: #58a6ff;
    color: #58a6ff;
  }
</style>
