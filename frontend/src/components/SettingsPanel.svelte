<script lang="ts">
  import { AppService } from '../../bindings/github.com/neko233/WinDownloader233';

  interface Props {
    lang: string;
    ui: Record<string, string>;
    onLanguageChange: (lang: string) => void;
    onClose: () => void;
  }

  let { lang, ui, onLanguageChange, onClose }: Props = $props();

  // Mirror settings
  let useProxy = $state(false);
  let githubProxy = $state('https://ghfast.top/');
  let httpProxy = $state('');

  // Load current config
  async function loadConfig() {
    try {
      const cfg = await AppService.GetMirrorConfig();
      useProxy = cfg.useProxy;
      githubProxy = cfg.githubProxy;
      httpProxy = cfg.httpProxy;
    } catch (e) {
      console.error(e);
    }
  }

  async function saveMirrorConfig() {
    try {
      await AppService.SetMirrorConfig({
        githubProxy,
        downloadCdn: githubProxy,
        useProxy,
        httpProxy,
      });
    } catch (e) {
      console.error(e);
    }
  }

  function selectProxy(url: string) {
    if (url) {
      githubProxy = url;
      saveMirrorConfig();
    }
  }

  const proxyOptions = [
    { name: 'ghfast.top', url: 'https://ghfast.top/', desc: '高速稳定 (推荐)' },
    { name: 'gh-proxy.com', url: 'https://gh-proxy.com/', desc: '免费公共代理' },
    { name: 'mirror.ghproxy.com', url: 'https://mirror.ghproxy.com/', desc: 'GitHub 镜像' },
  ];

  // Init
  loadConfig();
</script>

<div class="overlay" role="dialog" aria-modal="true" aria-label="Settings" onclick={onClose} onkeydown={() => {}}>
  <div class="panel" role="document" onclick={(e) => e.stopPropagation()} onkeydown={() => {}}>
    <div class="panel-header">
      <h2>{ui['settings'] || '设置'}</h2>
      <button class="close-btn" onclick={onClose}>✕</button>
    </div>

    <div class="panel-body">
      <!-- Language -->
      <section class="setting-section">
        <h3 class="setting-title">🌐 {ui['language'] || '语言'}</h3>
        <div class="lang-options">
          <button
            class="lang-btn"
            class:active={lang === 'zh'}
            onclick={() => onLanguageChange('zh')}
          >
            中文
          </button>
          <button
            class="lang-btn"
            class:active={lang === 'en'}
            onclick={() => onLanguageChange('en')}
          >
            English
          </button>
        </div>
      </section>

      <!-- Mirror Proxy -->
      <section class="setting-section">
        <h3 class="setting-title">🚀 {ui['mirror'] || '镜像加速'}</h3>
        <p class="setting-desc">
          {lang === 'zh'
            ? '启用 GitHub 镜像代理，解决中国网络访问 GitHub 缓慢的问题'
            : 'Enable GitHub mirror proxy for faster access in China'}
        </p>

        <label class="toggle-row">
          <span>{lang === 'zh' ? '启用镜像代理' : 'Enable Mirror Proxy'}</span>
          <input type="checkbox" bind:checked={useProxy} onchange={saveMirrorConfig} />
          <span class="toggle-slider"></span>
        </label>

        {#if useProxy}
          <div class="proxy-options">
            <p class="proxy-label">{lang === 'zh' ? '选择代理:' : 'Select proxy:'}</p>
            {#each proxyOptions as opt}
              <button
                class="proxy-btn"
                class:active={githubProxy === opt.url}
                onclick={() => selectProxy(opt.url)}
              >
                <span class="proxy-name">{opt.name}</span>
                <span class="proxy-desc">{opt.desc}</span>
              </button>
            {/each}

            <div class="custom-proxy">
              <label class="proxy-label" for="custom-proxy-input">{lang === 'zh' ? '自定义代理:' : 'Custom proxy:'}</label>
              <input
                id="custom-proxy-input"
                type="text"
                class="proxy-input"
                placeholder="https://your-proxy.com/"
                bind:value={githubProxy}
                onchange={saveMirrorConfig}
              />
            </div>
          </div>

          <div class="http-proxy-section">
            <label class="proxy-label" for="http-proxy-input">HTTP {lang === 'zh' ? '代理:' : 'Proxy:'}</label>
            <input
              id="http-proxy-input"
              type="text"
              class="proxy-input"
              placeholder="http://127.0.0.1:7890"
              bind:value={httpProxy}
              onchange={saveMirrorConfig}
            />
            <p class="setting-hint">
              {lang === 'zh' ? '用于下载工具时走本地代理（如 Clash、V2Ray）' : 'Use local proxy (Clash, V2Ray, etc.) for downloads'}
            </p>
          </div>
        {/if}
      </section>

      <!-- Registry Info -->
      <section class="setting-section">
        <h3 class="setting-title">📋 {lang === 'zh' ? '数据源' : 'Registry'}</h3>
        <p class="setting-desc">
          {lang === 'zh'
            ? '工具列表由 GitHub 上的 JSON 文件驱动，无需更新应用即可获取最新工具列表。本地数据 + 远程数据对比时间戳，取最新的。'
            : 'Tool list is driven by a GitHub JSON file. No app update needed. Local + remote data compared by timestamp, newest wins.'}
        </p>
      </section>
    </div>
  </div>
</div>

<style>
  .overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    backdrop-filter: blur(4px);
  }

  .panel {
    background: #161822;
    border: 1px solid #2d3148;
    border-radius: 16px;
    width: 500px;
    max-width: 90vw;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 20px 24px;
    border-bottom: 1px solid #2d3148;
  }

  .panel-header h2 {
    font-size: 18px;
    font-weight: 600;
    color: #e1e4e8;
  }

  .close-btn {
    background: none;
    border: none;
    color: #6e7681;
    font-size: 18px;
    cursor: pointer;
    padding: 4px 8px;
    border-radius: 4px;
  }

  .close-btn:hover {
    background: #1c2030;
    color: #c9d1d9;
  }

  .panel-body {
    padding: 20px 24px;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    gap: 24px;
  }

  .setting-section {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .setting-title {
    font-size: 15px;
    font-weight: 600;
    color: #c9d1d9;
  }

  .setting-desc {
    font-size: 13px;
    color: #8b949e;
    line-height: 1.5;
  }

  .lang-options {
    display: flex;
    gap: 8px;
  }

  .lang-btn {
    padding: 8px 20px;
    border: 1px solid #2d3148;
    border-radius: 8px;
    background: transparent;
    color: #8b949e;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.15s;
    font-family: inherit;
  }

  .lang-btn:hover {
    border-color: #58a6ff;
    color: #c9d1d9;
  }

  .lang-btn.active {
    background: #1f3a5f;
    border-color: #58a6ff;
    color: #58a6ff;
    font-weight: 600;
  }

  .toggle-row {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    font-size: 14px;
    color: #c9d1d9;
    user-select: none;
  }

  .toggle-row input {
    display: none;
  }

  .toggle-slider {
    position: relative;
    width: 40px;
    height: 22px;
    background: #2d3148;
    border-radius: 11px;
    transition: background 0.2s;
    margin-left: auto;
  }

  .toggle-slider::after {
    content: '';
    position: absolute;
    top: 3px;
    left: 3px;
    width: 16px;
    height: 16px;
    background: #8b949e;
    border-radius: 50%;
    transition: transform 0.2s, background 0.2s;
  }

  .toggle-row input:checked + .toggle-slider {
    background: #238636;
  }

  .toggle-row input:checked + .toggle-slider::after {
    transform: translateX(18px);
    background: #fff;
  }

  .proxy-options, .http-proxy-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-left: 8px;
    border-left: 2px solid #2d3148;
  }

  .proxy-label {
    font-size: 13px;
    color: #8b949e;
  }

  .proxy-btn {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 8px 12px;
    border: 1px solid #2d3148;
    border-radius: 8px;
    background: transparent;
    color: #c9d1d9;
    cursor: pointer;
    transition: all 0.15s;
    text-align: left;
    font-family: inherit;
  }

  .proxy-btn:hover {
    border-color: #58a6ff;
  }

  .proxy-btn.active {
    background: #1f3a5f;
    border-color: #58a6ff;
  }

  .proxy-name {
    font-size: 13px;
    font-weight: 500;
  }

  .proxy-desc {
    font-size: 11px;
    color: #8b949e;
  }

  .custom-proxy {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .proxy-input {
    padding: 8px 12px;
    background: #0f1117;
    border: 1px solid #2d3148;
    border-radius: 6px;
    color: #c9d1d9;
    font-size: 13px;
    font-family: 'Consolas', monospace;
    outline: none;
  }

  .proxy-input:focus {
    border-color: #58a6ff;
  }

  .proxy-input::placeholder {
    color: #484f58;
  }

  .setting-hint {
    font-size: 11px;
    color: #6e7681;
    line-height: 1.4;
  }
</style>
