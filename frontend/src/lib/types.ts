export interface Tool {
  id: string;
  nameZh: string;
  nameEn: string;
  descZh: string;
  descEn: string;
  category: string;
  tags: string[];
  icon: string;
  version: string;
  size: string;
  wingetId: string;
  downloadUrl: string;
  mirrorUrl: string;
  homepage: string;
  isFree: boolean;
  installType: string;
  installed: boolean;
}

export interface Progress {
  toolId: string;
  status: string;
  percent: number;
  message: string;
  startedAt?: string;
  updatedAt?: string;
  logLines?: number;
  exitCode?: number;
}

export interface LogEntry {
  time: string;
  toolId: string;
  level: string;
  message: string;
}

export interface MirrorConfig {
  githubProxy: string;
  downloadCdn: string;
  useProxy: boolean;
  httpProxy: string;
}

export interface ProxyOption {
  name: string;
  url: string;
  description: string;
}

export type Category =
  | 'programming'
  | 'art'
  | 'planning'
  | 'audio'
  | 'qa'
  | 'pm'
  | 'ai';

export const ALL_CATEGORIES: Category[] = [
  'programming',
  'art',
  'planning',
  'audio',
  'qa',
  'pm',
  'ai',
];

export const CATEGORY_ICONS: Record<Category, string> = {
  programming: '⌨️',
  art: '🎨',
  planning: '📋',
  audio: '🎵',
  qa: '🧪',
  pm: '📊',
  ai: '🤖',
};
