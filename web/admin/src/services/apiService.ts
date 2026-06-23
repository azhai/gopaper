import m from 'mithril';

const API_BASE = '/api/admin';

interface ApiResponse<T = unknown> {
  code: number;
  message?: string;
  data?: T;
  total?: number;
  page?: number;
  pageSize?: number;
}

async function request<T>(method: string, url: string, body?: unknown): Promise<ApiResponse<T>> {
  const token = authService.getToken();
  const headers: Record<string, string> = {};

  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const opts: RequestInit = { method, headers };
  if (body) {
    opts.body = body instanceof FormData ? body : JSON.stringify(body);
  }

  const response = await fetch(`${API_BASE}${url}`, opts);

  if (response.status === 401) {
    authService.logout();
    m.route.set('/login');
    throw new Error('会话已过期');
  }

  const result = await response.json();
  return result as ApiResponse<T>;
}

export const apiService = {
  login: (username: string, password: string) =>
    request<{ token: string; expireAt: string }>('POST', '/login', { username, password }),

  getArticles: (page = 1, pageSize = 10, dir?: string, type?: 'page' | 'article') => {
    let url = `/articles?page=${page}&pageSize=${pageSize}`;
    if (dir) url += `&dir=${encodeURIComponent(dir)}`;
    if (type) url += `&type=${type}`;
    return request<unknown[]>('GET', url);
  },

  getArticle: (slug: string) =>
    request<unknown>('GET', `/articles/${slug}`),

  createArticle: (data: unknown) =>
    request('POST', '/articles', data),

  updateArticle: (slug: string, data: unknown) =>
    request('PUT', `/articles/${slug}`, data),

  deleteArticle: (slug: string) =>
    request('DELETE', `/articles/${slug}`),

  previewArticle: (content: string) =>
    request<{ html: string }>('POST', '/articles/preview', { content }),

  getImages: (page = 1, pageSize = 20) =>
    request<unknown[]>('GET', `/images?page=${page}&pageSize=${pageSize}`),

  uploadImage: (file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return request<{ fileName: string; uploadPath: string }>('POST', '/images', formData);
  },

  deleteImage: (fileName: string) =>
    request('DELETE', `/images/${fileName}`),

  refreshCache: () =>
    request<{ articleCount: number; duration: string }>('POST', '/cache/refresh'),

  getDirs: () =>
    request<DirInfo[]>('GET', '/dirs'),

  getSettings: () =>
    request<Record<string, unknown>>('GET', '/settings'),

  updateSettings: (data: unknown) =>
    request('PUT', '/settings', data),

  getLayouts: () =>
    request<LayoutConfig>('GET', '/layouts'),

  saveLayouts: (data: unknown) =>
    request('PUT', '/layouts', data),

  getRegions: (name: string) =>
    request<Region[]>('GET', `/layouts/${name}/regions`),
};

export interface Region {
  name: string;
  title: string;
}

export interface TemplateInfo {
  name: string;
  title: string;
  file: string;
  desc: string;
  regions: Region[];
}

export interface LayoutConfig {
  templates: TemplateInfo[];
}

export interface DirInfo {
  dirPath: string;
  title: string;
  dirType: string;
  layout: string;
  sortOrder: string;
  navOrder: number;
  articleCount: number;
}

export const authService = {
  getToken: (): string | null => localStorage.getItem('admin_token'),

  setToken: (token: string) => localStorage.setItem('admin_token', token),

  logout: () => localStorage.removeItem('admin_token'),

  isLoggedIn: (): boolean => !!localStorage.getItem('admin_token'),
};